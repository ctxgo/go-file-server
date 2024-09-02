package middlewares

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/global"
	"go-file-server/internal/common/types"
	"go-file-server/pkgs/config"
	"go-file-server/pkgs/zlog"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

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
func JwtAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenStr string
		var err error
		defer func() {
			if err != nil {
				zlog.SugLog.Error(err)
				core.ErrRep().
					SetHttpCode(global.UnauthorizedError).
					SetBizCode(global.BizUnauthorizedErr).
					SetMsg(err.Error()).
					SendGin(c)
			}
		}()

		tokenStr, err = GetToken(c)
		if err != nil {
			return
		}

		jwtClaims, err := ParseToken(tokenStr)
		if err != nil {
			return
		}
		c.Set(global.JwtPayloadKey, jwtClaims)

	}
}

func CreateToken(f func(*types.JwtClaims)) (string, string, error) {
	expiresAt := time.Now().Add(time.Minute * time.Duration(config.JwtCfg.Timeout))
	var claims = types.JwtClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt.Unix()},
	}
	f(&claims)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(config.JwtCfg.Secret))
	return tokenStr, expiresAt.Format(time.RFC3339), err
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
