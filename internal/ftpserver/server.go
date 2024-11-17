package ftpserver

import (
	"crypto/tls"
	"fmt"
	"go-file-server/internal/common/middlewares"
	"go-file-server/internal/common/repository"
	"go-file-server/internal/common/types"
	"go-file-server/internal/services/admin/apis/fs/utils"
	"go-file-server/internal/services/admin/models"
	"go-file-server/internal/services/normal/apis/auth"
	"go-file-server/pkgs/cache"
	"sync"
	"time"

	Casbin "github.com/casbin/casbin/v2"
	goCache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go4.org/syncutil/singleflight"
	"gorm.io/gorm"

	serverlib "github.com/fclairamb/ftpserverlib"
)

const ftpserverKey = "ftpserverKey"

// Server structure
type Server struct {
	logger           *zap.SugaredLogger
	addr             string
	passivePortRange *serverlib.PortRange
	publicHost       string
	nbClients        uint32
	nbClientsSync    sync.Mutex
	zeroClientEvent  chan error
	session          *goCache.Cache
	userRepo         *repository.UserRepository
	roleRepo         *repository.RoleRepository
	loginLogRepo     *repository.LoginLogRepository
	fsRepo           *repository.FsRepository
	casbinEnforcer   *Casbin.CachedEnforcer
	requestGroup     singleflight.Group
	cache            cache.AdapterCache
	limiterManager   *utils.LimiterManager
}

// ErrTimeout is returned when an operation timeouts
var ErrTimeout = errors.New("timeout")

// ErrNotImplemented is returned when we're using something that has not been implemented yet
// var ErrNotImplemented = errors.New("not implemented")

// ErrNotEnabled is returned when a feature hasn't been enabled
var ErrNotEnabled = errors.New("not enabled")

type opt func(*Server)

func WithLogger(log *zap.SugaredLogger) opt {
	return func(s *Server) {
		s.logger = log
	}
}

func WithAddr(addr string) opt {
	return func(s *Server) {
		s.addr = addr
	}
}

func WithPublicHost(host string) opt {
	return func(s *Server) {
		s.publicHost = host
	}
}

func WithPassivePortRange(start, end int) opt {
	return func(s *Server) {
		s.passivePortRange = &serverlib.PortRange{
			Start: start,
			End:   end,
		}
	}
}

// NewServer creates a server instance
func NewServer(svcCtx *types.SvcCtx, opts ...opt) (*Server, error) {
	server := &Server{
		session:        goCache.New(8*time.Hour, 10*time.Hour),
		userRepo:       repository.NewUserRepository(svcCtx.Db),
		roleRepo:       repository.NewRoleRepository(svcCtx.Db),
		loginLogRepo:   repository.NewLoginLogRepository(svcCtx.Db),
		fsRepo:         repository.NewFsRepository(svcCtx.FsIndexer),
		casbinEnforcer: svcCtx.CasbinEnforcer,
		cache:          svcCtx.Cache,
		limiterManager: utils.NewLimiterManager(30*time.Minute, 30*time.Minute),
	}

	for _, x := range opts {
		x(server)
	}
	if server.logger == nil {
		server.logger = zap.NewNop().Sugar()
	}
	return server, nil
}

// GetSettings returns some general settings around the server setup
func (s *Server) GetSettings() (*serverlib.Settings, error) {
	return &serverlib.Settings{
		ListenAddr:               s.addr,
		PassiveTransferPortRange: s.passivePortRange,
		PublicHost:               s.publicHost,
	}, nil
}

// ClientConnected is called to send the very first welcome message
func (s *Server) ClientConnected(cc serverlib.ClientContext) (string, error) {
	s.nbClientsSync.Lock()
	defer s.nbClientsSync.Unlock()
	s.nbClients++
	s.logger.Infof(
		"Client connected, clientId: %d, remoteAddr: %s, nbClients: %d.",
		cc.ID(), cc.RemoteAddr().String(), s.nbClients,
	)
	return "ftpserver", nil
}

// ClientDisconnected is called when the user disconnects, even if he never authenticated
func (s *Server) ClientDisconnected(cc serverlib.ClientContext) {
	s.nbClientsSync.Lock()
	defer s.nbClientsSync.Unlock()

	s.nbClients--
	s.logger.Infof(
		"Client disconnected, clientId: %d, remoteAddr: %s, nbClients: %d.",
		cc.ID(), cc.RemoteAddr().String(), s.nbClients,
	)
	s.considerEnd()
}

// Stop will trigger a graceful stop of the server. All currently connected clients won't be disconnected instantly.
func (s *Server) Stop() {
	s.nbClientsSync.Lock()
	defer s.nbClientsSync.Unlock()
	s.zeroClientEvent = make(chan error, 1)
	s.considerEnd()
}

// WaitGracefully allows to gracefully wait for all currently connected clients before disconnecting
func (s *Server) WaitGracefully(timeout time.Duration) error {
	s.logger.Info("Waiting for last client to disconnect...")

	defer func() { s.zeroClientEvent = nil }()

	select {
	case err := <-s.zeroClientEvent:
		return err
	case <-time.After(timeout):
		return ErrTimeout
	}
}

func (s *Server) considerEnd() {
	if s.nbClients == 0 && s.zeroClientEvent != nil {
		s.zeroClientEvent <- nil
		close(s.zeroClientEvent)
	}
}

func genFtpserverKey(user, pass string) string {
	return fmt.Sprintf("%s_%s_%s", ftpserverKey, user, pass)
}

func (s *Server) getSession(key string) (*FileServerFs, bool) {
	data, ok := s.session.Get(key)
	if ok {
		return data.(*FileServerFs), true
	}
	return nil, false
}

func (s *Server) createSession(key, user, pass string) (*FileServerFs, error) {
	userInfo, err := auth.VerifyUser(s.userRepo, auth.LoginReq{
		Username: user,
		Password: pass,
	})
	if err != nil {
		return nil, err
	}

	role, err := s.roleRepo.FindOne(repository.WithRoleId(userInfo.RoleId))
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, errors.Errorf("无任何目录权限，请联系管理员赋权")
	}

	token, _, err := middlewares.CreateToken(
		func(jc *types.JwtClaims) {
			jc.RoleId = role.RoleId
			jc.RoleKey = role.RoleKey
			jc.RoleName = role.RoleName
			jc.Username = userInfo.Username
			jc.UserId = userInfo.UserId
		},
	)
	if err != nil {
		return nil, err
	}

	fileServerFs := &FileServerFs{
		token:          token,
		user:           user,
		roleKey:        role.RoleKey,
		fsRepo:         s.fsRepo,
		roleRepo:       s.roleRepo,
		casbinEnforcer: s.casbinEnforcer,
		cache:          s.cache,
		limiterManager: s.limiterManager,
	}
	s.session.Set(key, fileServerFs, 0)
	return fileServerFs, nil
}

// AuthUser authenticates the user and selects an handling driver
func (s *Server) AuthUser(cc serverlib.ClientContext, user, pass string) (serverlib.ClientDriver, error) {
	key := genFtpserverKey(user, pass)
	session, ok := s.getSession(key)

	if ok {
		jwtClaims, err := middlewares.ParseToken(session.token)
		if err != nil {
			return nil, err
		}
		var lastTokenReset int64
		lastTokenReset, err = middlewares.GetLastTokenReset(s.cache, jwtClaims.UserId)
		if err != nil {
			return nil, err
		}
		if jwtClaims.IssuedAt > lastTokenReset {
			return session, nil
		}
	}

	result, err := s.requestGroup.Do(key, func() (any interface{}, err error) {
		var status string = "1"
		defer func() {
			if err != nil {
				status = "2"
			}

			go s.loginLogRepo.Create(&models.SysLoginLog{
				Username: user,
				Remark:   "ftp",
				Msg:      "ftp",
				Ipaddr:   cc.RemoteAddr().String(),
				Status:   status,
			})
		}()
		return s.createSession(key, user, pass)
	})

	return result.(*FileServerFs), err
}

func (s *Server) GetTLSConfig() (*tls.Config, error) {
	return nil, ErrNotEnabled
}
