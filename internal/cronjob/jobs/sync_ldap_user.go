package jobs

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"go-file-server/internal/common/repository"
	"go-file-server/pkgs/config"
	"go-file-server/pkgs/zlog"
	"os"

	ldapv3 "github.com/go-ldap/ldap/v3"
	"github.com/thoas/go-funk"
	"gorm.io/gorm"

	dexv2 "github.com/dexidp/dex/api/v2"
	"github.com/dexidp/dex/connector/ldap"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type LDAPUserSyncer struct {
	userRepo *repository.UserRepository
}

func NewLDAPUserSyncer(db *gorm.DB) *LDAPUserSyncer {
	return &LDAPUserSyncer{
		userRepo: repository.NewUserRepository(db),
	}
}

func (s *LDAPUserSyncer) Run() {
	ldapConfig, err := getLDAPConnector()
	if err != nil {
		zlog.SugLog.Error(err)
		return
	}

	conn, err := s.getLDAPConnection(ldapConfig)
	if err != nil {
		zlog.SugLog.Error(err)
		return
	}
	defer conn.Close()

	users, count, err := s.userRepo.Find(repository.WithUserSource("ldap"), repository.WithUserStatus("2"))
	if err != nil {
		zlog.SugLog.Error(err)
		return
	}
	if count == 0 {
		return
	}

	ldapEntries, err := s.searchLDAPUsers(conn, ldapConfig)
	if err != nil {
		zlog.SugLog.Error(err)
		return
	}

	var userIDs []int
	for _, user := range users {
		if funk.Contains(ldapEntries, func(entry *ldapv3.Entry) bool {
			return user.Username == entry.GetAttributeValue(ldapConfig.UserSearch.NameAttr)
		}) {
			continue
		}
		userIDs = append(userIDs, user.UserId)
	}

	if len(userIDs) == 0 {
		return
	}

	err = s.userRepo.Updates(map[string]interface{}{"status": "1"}, repository.WithUserIds(userIDs...))
	if err != nil {
		zlog.SugLog.Error(err)
	}
}

func (s *LDAPUserSyncer) getLDAPConnection(ldapConfig *ldap.Config) (*ldapv3.Conn, error) {
	var (
		conn *ldapv3.Conn
		err  error
	)

	if ldapConfig.InsecureNoSSL {
		conn, err = ldapv3.DialURL("ldap://" + ldapConfig.Host)
		if err != nil {
			return nil, err
		}
	} else {

		clientCert, err := tls.X509KeyPair([]byte(ldapConfig.ClientKey), []byte(ldapConfig.ClientCert))
		if err != nil {
			return nil, errors.Wrap(err, "无法加载客户端证书和密钥")
		}
		tlsConfig, err := genTlsConfig(ldapConfig.RootCAData, clientCert)
		if err != nil {
			return nil, errors.Wrap(err, "无法生成证书")
		}
		conn, err = ldapv3.DialURL("ldaps://"+ldapConfig.Host, ldapv3.DialWithTLSConfig(tlsConfig))
		if err != nil {
			return nil, err
		}
	}

	err = conn.Bind(ldapConfig.BindDN, ldapConfig.BindPW)
	if err != nil {
		return nil, errors.Errorf("LDAP 绑定失败，主机：%s，错误信息：%s", ldapConfig.Host, err)
	}

	return conn, nil
}

func (s *LDAPUserSyncer) searchLDAPUsers(conn *ldapv3.Conn, ldapConfig *ldap.Config) ([]*ldapv3.Entry, error) {
	searchRequest := ldapv3.NewSearchRequest(
		ldapConfig.UserSearch.BaseDN, // 搜索的基准 DN
		ldapv3.ScopeWholeSubtree, ldapv3.NeverDerefAliases, 0, 0, false,
		ldapConfig.UserSearch.Filter,             // 搜索过滤器
		[]string{ldapConfig.UserSearch.NameAttr}, // 要返回的属性列表
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	return sr.Entries, nil
}

func getLDAPConnector() (*ldap.Config, error) {
	oAuthGrpc := config.OAuthCfg.Grpc

	transportCredentials := insecure.NewCredentials()

	if oAuthGrpc.TlsCert != "" && oAuthGrpc.TlsKey != "" && oAuthGrpc.TlsCA != "" {
		caCert, err := os.ReadFile(oAuthGrpc.TlsCA)
		if err != nil {
			return nil, fmt.Errorf("invalid CA crt file: %s", oAuthGrpc.TlsCA)
		}

		clientCert, err := tls.LoadX509KeyPair(oAuthGrpc.TlsCert, oAuthGrpc.TlsKey)
		if err != nil {
			return nil, fmt.Errorf("invalid client crt file: %s, %s", oAuthGrpc.TlsCert, oAuthGrpc.TlsKey)
		}
		tlsConfig, err := genTlsConfig(caCert, clientCert)
		if err != nil {
			return nil, err
		}
		transportCredentials = credentials.NewTLS(tlsConfig)
	}

	conn, err := grpc.NewClient(oAuthGrpc.Addr, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		return nil, errors.Wrap(err, "连接 gRPC 服务失败")
	}
	defer conn.Close()

	client := dexv2.NewDexClient(conn)
	resp, err := client.ListConnectors(context.Background(), &dexv2.ListConnectorReq{})
	if err != nil {
		return nil, errors.Wrap(err, "获取 connectors 列表失败")
	}

	for _, connector := range resp.Connectors {
		if connector.Type == "ldap" {
			var data ldap.Config
			err := json.Unmarshal(connector.Config, &data)
			if err != nil {
				return nil, errors.Wrap(err, "解析 LDAP 配置失败")
			}
			return &data, nil
		}
	}
	return nil, errors.New("未找到 LDAP 连接器")
}

func genTlsConfig(caCert []byte, clientCert tls.Certificate) (*tls.Config, error) {

	cPool := x509.NewCertPool()

	if !cPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to parse CA crt")
	}

	clientTLSConfig := &tls.Config{
		RootCAs:      cPool,
		Certificates: []tls.Certificate{clientCert},
	}
	return clientTLSConfig, nil
}
