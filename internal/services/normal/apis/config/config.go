package config

import (
	"go-file-server/internal/common/core"
	"go-file-server/pkgs/config"

	"github.com/gin-gonic/gin"
)

type Config struct {
	LdapEnabled bool `json:"ldapEnabled"`
}

type ConfigHandler gin.HandlerFunc

func NewConfigHandler() ConfigHandler {
	return func(c *gin.Context) {
		data := Config{LdapEnabled: config.OAuthCfg.Enable}
		core.OKRep(data).SendGin(c)
	}
}
