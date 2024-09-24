package jobs

import (
	"go-file-server/internal/common/repository"
	"go-file-server/pkgs/base"
	"go-file-server/pkgs/config"
	"go-file-server/pkgs/zlog"
	"log"
	"testing"

	"gorm.io/gorm/logger"
)

func TestLdapUserSyncer_Run(t *testing.T) {
	type fields struct {
		userRepo *repository.UserRepository
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{}, // TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.Init(config.SetFile("your config path"))
			zlog.Init()
			driverOpen, err := base.GetDriverOpen(config.DatabaseCfg.Driver)
			if err != nil {
				log.Fatal(err)
			}
			db, err := base.InitDatabase(
				config.DatabaseCfg.Source,
				driverOpen,
				base.SetLogger(logger.Default.LogMode(logger.Info)),
			)
			if err != nil {
				log.Fatal(err)
			}
			c := &LDAPUserSyncer{
				userRepo: repository.NewUserRepository(db),
			}
			c.Run()
		})
	}
}
