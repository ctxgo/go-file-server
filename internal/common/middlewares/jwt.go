package middlewares

import (
	"fmt"
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/global"
	"go-file-server/internal/common/repository"
	"go-file-server/internal/common/types"
	"go-file-server/pkgs/cache"
	"go-file-server/pkgs/config"
	"go-file-server/pkgs/zlog"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const lastTokenResetPrefix = "last_token_reset:"

func GetToken(c *gin.Context) (string, error) {
	bearerToken := c.Request.Header.Get("Authorization")
	if bearerToken == "" {
		token := c.Query("token")
		if token != "" {
			return token, nil
		}
		return "", errors.Errorf(global.ErrEmptyAToken)

	}

	tokenParts := strings.Split(bearerToken, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return "", errors.Errorf(global.ErrTokenMalformed)
	}

	return tokenParts[1], nil

}

// JwtAuth 中间件，检查token
func Auth(authenticator *Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenStr string
		var err error
		defer func() {
			if err != nil {
				zlog.SugLog.Error(err)
				core.ErrRep().
					SetHttpCode(global.HttpSuccess).
					SetBizCode(global.BizUnauthorizedErr).
					SetMsg(err.Error()).
					SendGin(c)
			}
		}()

		tokenStr, err = GetToken(c)
		if err != nil {
			return
		}
		var jwtClaims *types.JwtClaims
		jwtClaims, err = authenticator.ValidateToken(tokenStr)

		if err != nil {
			return
		}

		c.Set(global.JwtPayloadKey, jwtClaims)

	}
}

type Authenticator struct {
	cache         cache.AdapterCache
	userTokenRepo *repository.UserTokenRepository
}

func NewAuthenticator(
	userTokenRepo *repository.UserTokenRepository,
	cache cache.AdapterCache,
) *Authenticator {
	return &Authenticator{
		userTokenRepo: userTokenRepo,
		cache:         cache,
	}
}

func (a *Authenticator) ValidateToken(tokenStr string) (*types.JwtClaims, error) {
	jwtClaims, err := ParseToken(tokenStr)
	if err != nil {
		return jwtClaims, err
	}
	var lastTokenReset int64
	lastTokenReset, err = GetLastTokenReset(a.cache, jwtClaims.UserId)
	if err != nil {
		return jwtClaims, err
	}
	if jwtClaims.IssuedAt < lastTokenReset {
		return jwtClaims, errors.Errorf(global.ErrTokenRevoked)

	}

	if jwtClaims.IsPersonalToken {
		return jwtClaims, a.validatePersonalToken(jwtClaims, tokenStr)
	}
	return jwtClaims, nil
}

func GenPersonalTokenRevokedKey(token string) string {
	return fmt.Sprintf("%s:%s", global.PersonalTokenRevokedKey, token)
}

func (a *Authenticator) validatePersonalToken(jwtClaims *types.JwtClaims, tokenStr string) error {
	tokenKey := GenPersonalTokenRevokedKey(tokenStr)
	isValidToken, err := a.cache.Get(tokenKey)
	if err != nil && !cache.IsKeyNotFoundError(err) {
		return errors.WithStack(err)
	}

	switch isValidToken {
	case "false":
		return errors.Errorf(global.ErrTokenRevoked)
	case "true":
		return nil
	}

	_, err = a.userTokenRepo.FindOne(repository.WithUserTokenUserId(jwtClaims.UserId),
		repository.WithUserToken(tokenStr))

	if err == nil {
		return a.cache.Set(tokenKey, "true", 24*time.Hour)
	}

	if err != gorm.ErrRecordNotFound {
		return errors.WithStack(err)
	}

	err = a.cache.Set(tokenKey, "false", 24*time.Hour)
	if err != nil {
		return errors.WithStack(err)
	}
	return errors.Errorf(global.ErrTokenRevoked)
}

func CreateToken(f func(*types.JwtClaims)) (string, string, error) {
	nowTime := time.Now()
	expiresAt := nowTime.Add(time.Minute * time.Duration(config.JwtCfg.Timeout))
	var claims = types.JwtClaims{
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  nowTime.Unix(),
			ExpiresAt: expiresAt.Unix()},
	}
	f(&claims)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(config.JwtCfg.Secret))
	return tokenStr, expiresAt.Format(time.RFC3339), err
}

func getLastTokenResetKey(userID int) string {
	return fmt.Sprintf("%s%d", lastTokenResetPrefix, userID)
}

func GetLastTokenReset(lcache cache.AdapterCache, userID int) (int64, error) {
	key := getLastTokenResetKey(userID)
	lastTokenReset, err := lcache.Get(key)
	if err == nil {
		return strconv.ParseInt(lastTokenReset, 10, 64)
	}
	if cache.IsKeyNotFoundError(err) {
		return 0, nil
	}
	zlog.SugLog.Error(err)
	return 0, errors.Errorf(global.ErrServerNotOK)
}

// 更新用户的last_token_reset时间
func UpdateLastTokenReset(cache cache.AdapterCache, userID int) error {
	key := getLastTokenResetKey(userID)
	currentTime := time.Now().Unix()
	return cache.Set(key, currentTime, 0)
}

// ParseToken 解析JWT Token
func ParseToken(tokenString string) (*types.JwtClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &types.JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.JwtCfg.Secret), nil
	})
	if err != nil {
		ve, ok := err.(*jwt.ValidationError)
		if !ok {
			zlog.SugLog.Error(err)
			return nil, errors.Errorf("内部错误")
		}

		switch {
		case ve.Errors&jwt.ValidationErrorMalformed != 0:
			return nil, errors.Errorf(global.ErrTokenMalformed)
		case ve.Errors&jwt.ValidationErrorExpired != 0:
			return nil, errors.Errorf(global.ErrTokenMalformed)
		case ve.Errors&jwt.ValidationErrorNotValidYet != 0:
			return nil, errors.Errorf(global.ErrTokenNotValidYet)
		default:
			return nil, errors.Errorf(global.ErrTokenInvalid)
		}

	}
	if claims, ok := token.Claims.(*types.JwtClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.Errorf(global.ErrTokenInvalid)
}

// 更新token
func RefreshToken(tokenString string) (string, string, error) {
	jwt.TimeFunc = func() time.Time {
		return time.Unix(0, 0)
	}
	token, err := jwt.ParseWithClaims(tokenString, &types.JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return config.JwtCfg.Secret, nil
	})
	if err != nil {
		return "", "", err
	}
	if claims, ok := token.Claims.(*types.JwtClaims); ok && token.Valid {
		jwt.TimeFunc = time.Now
		claims.StandardClaims.ExpiresAt = time.Now().Add(2 * time.Hour).Unix()
		return CreateToken(func(jc *types.JwtClaims) {
			claims.StandardClaims = jc.StandardClaims
			*jc = *claims
		})
	}
	return "", "", errors.Errorf(global.ErrTokenInvalid)
}
