package cronjob

import (
	"go-file-server/internal/cronjob/jobs"
	"go-file-server/pkgs/config"
	"go-file-server/pkgs/zlog"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

func InitJobs(db *gorm.DB) {
	c := cron.New(cron.WithSeconds(),
		cron.WithLogger(cron.VerbosePrintfLogger(&CornLogger{zlog.SugLog})),
	)
	RegisterLdapJob(c, db)

	if len(c.Entries()) == 0 {
		return
	}
	c.Start()
}

func RegisterLdapJob(c *cron.Cron, db *gorm.DB) {
	if !config.OAuthCfg.Enable {
		return
	}
	_, err := c.AddJob("0 0 2 * * *", jobs.NewLDAPUserSyncer(db))
	if err != nil {
		zlog.SugLog.Fatalf("无法注册ldap用户同步任务: %v", err)
	}
}
