package base

import (
	"fmt"
	"go-file-server/pkgs/base/plugin"
	"go-file-server/pkgs/utils/retry"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const defaultRetryInterval = 1 * time.Second

type Option func(*DbConfig)
type DriverOpen func(string) gorm.Dialector
type DbConfig struct {
	db              *gorm.DB
	logger          logger.Interface
	RetryAttempts   int           // 失败后重试次数
	RetryInterval   time.Duration // 重试间隔
	MaxIdleConns    int           // 设置连接池中的空闲连接的最大数量
	MaxOpenConns    int           // 设置数据库的最大连接数量
	ConnMaxLifetime time.Duration // 设置连接的最大可复用时间
}

func GetDriverOpen(Driver string) (DriverOpen, error) {
	switch Driver {
	case "postgres":
		return postgres.Open, nil
	case "mysql":
		return mysql.Open, nil
	default:
		return nil, errors.Errorf("unkonw db Driver: " + Driver)
	}

}

func SetLogger(logger logger.Interface) Option {
	return func(dc *DbConfig) {
		dc.logger = logger
	}
}

func SetRetryAttempts(r int) Option {
	return func(dc *DbConfig) {
		dc.RetryAttempts = r
	}
}

func SetRetryInterval(t time.Duration) Option {
	return func(dc *DbConfig) {
		dc.RetryInterval = t
	}
}

func SetMaxIdleConns(m int) Option {
	return func(dc *DbConfig) {
		dc.MaxIdleConns = m
	}
}

func SetMaxOpenConns(m int) Option {
	return func(dc *DbConfig) {
		dc.MaxOpenConns = m
	}
}

func SetConnMaxLifetime(t time.Duration) Option {
	return func(dc *DbConfig) {
		dc.ConnMaxLifetime = t
	}
}

func InitDatabase(dsn string, open DriverOpen, opts ...Option) (*gorm.DB, error) {

	dbCfg := &DbConfig{
		RetryInterval: defaultRetryInterval,
	}

	for _, o := range opts {
		o(dbCfg)
	}

	if dbCfg.logger == nil {
		dbCfg.logger = NewZapLoggerAdapter(zap.New(zapcore.NewNopCore()).Sugar())
	}
	if err := dbCfg.Open(dsn, open); err != nil {
		return nil, err
	}
	dbCfg.ApplyPlugin()

	return dbCfg.db, nil
}

func (dbCfg *DbConfig) Open(dsn string, open func(string) gorm.Dialector) (err error) {
	var db *gorm.DB
	retry.Retry(func() error {
		db, err = gorm.Open(open(dsn), &gorm.Config{Logger: dbCfg.logger})
		return err
	},
		retry.WithAttempts(dbCfg.RetryAttempts),
		retry.WithDelay(dbCfg.RetryInterval),
	)
	if err != nil {
		return fmt.Errorf("数据库连接失败: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	if dbCfg.MaxIdleConns > 0 {
		// SetMaxIdleConns 设置连接池中的空闲连接的最大数量
		sqlDB.SetMaxIdleConns(dbCfg.MaxIdleConns)
	}
	if dbCfg.MaxOpenConns > 0 {
		// SetMaxOpenConns 设置数据库的最大连接数量
		sqlDB.SetMaxOpenConns(dbCfg.MaxOpenConns)
	}
	if dbCfg.ConnMaxLifetime > 0 {
		// SetConnMaxLifetime 设置连接的最大可复用时间
		sqlDB.SetConnMaxLifetime(dbCfg.ConnMaxLifetime)
	}
	dbCfg.db = db
	return nil
}

func (dbCfg *DbConfig) ApplyPlugin() {
	if dbCfg.RetryAttempts > 0 {
		retryPluginInstance := &plugin.RetryPlugin{
			MaxAttempts:          dbCfg.RetryAttempts,
			DelayBetweenAttempts: dbCfg.RetryInterval,
		}
		dbCfg.db.Use(retryPluginInstance)
	}
}
