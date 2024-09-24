package auth

import (
	"context"
	"fmt"
	"go-file-server/internal/common/core"
	"go-file-server/pkgs/config"

	"github.com/coreos/go-oidc"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

func (u *Authenticator) LoginDex(c *gin.Context) {
	oauth2Config, _, err := getOauthConfig(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}
	authCodeURL := oauth2Config.AuthCodeURL(config.OAuthCfg.State)
	core.OKRep(authCodeURL).SendGin(c)
}

func getOauthConfig(ctx context.Context) (*oauth2.Config, *oidc.Provider, error) {

	oauthCfg := config.OAuthCfg
	provider, err := oidc.NewProvider(ctx, oauthCfg.IssuerUrl)
	if err != nil {
		return nil, nil, fmt.Errorf("无法初始化 OIDC 提供者: %v", err)
	}

	oauth2Config := &oauth2.Config{
		ClientID:     oauthCfg.ClientID,
		ClientSecret: oauthCfg.ClientSecret,
		RedirectURL:  oauthCfg.RedirectUrl,
		Endpoint:     provider.Endpoint(),
		Scopes:       oauthCfg.Scopes,
	}
	var data = new(map[string]any)
	provider.Claims(data)
	return oauth2Config, provider, nil
}
