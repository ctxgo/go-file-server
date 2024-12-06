package init

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	coreModels "go-file-server/internal/common/models"
	"go-file-server/internal/services/admin/models"
	"go-file-server/pkgs/base"
	"go-file-server/pkgs/config"
	"go-file-server/pkgs/zlog"
	"go-file-server/sql"
	"strings"
	"time"

	"gorm.io/gorm"
)

func initDB() (*gorm.DB, error) {
	driverOpen, err := base.GetDriverOpen(config.DatabaseCfg.Driver)
	if err != nil {
		return nil, err
	}
	db, err := base.InitDatabase(
		config.DatabaseCfg.Source,
		driverOpen,
		base.SetLogger(base.NewZapLoggerAdapter(zlog.SugLog)),
		base.SetRetryAttempts(6),
		base.SetRetryInterval(3*time.Second),
	)
	if err != nil {
		return nil, err
	}

	return db, initializeDBData(db)
}

func initializeDBData(db *gorm.DB) error {
	err := db.AutoMigrate(&coreModels.DBInitStatus{})
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		// 尝试获取一个全局锁
		var lockStatus int
		err := tx.Raw("SELECT GET_LOCK('db_init_lock', 10)").Scan(&lockStatus).Error
		if err != nil || lockStatus != 1 {
			return fmt.Errorf("failed to acquire lock: %v", err)
		}
		defer tx.Exec("SELECT RELEASE_LOCK('db_init_lock')") // 确保锁被释放

		var initStatus coreModels.DBInitStatus
		err = tx.First(&initStatus, 1).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if err == nil && initStatus.Initialized {
			return nil
		}

		err = migrateaTable(db)
		if err != nil {
			return err
		}

		if err := executeEmbeddedSQL(tx); err != nil {
			return err
		}

		return tx.Save(&coreModels.DBInitStatus{ID: 1, Initialized: true}).Error
	})
}

func migrateaTable(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Avatar{},
		&models.SysLoginLog{},
		&models.SysOperaLog{},
		&models.SysApi{},
		&models.SysMenu{},
		&models.SysRole{},
		&models.SysUser{},
		&models.UserToken{},
		&models.SysDept{},
	)
}

func executeEmbeddedSQL(db *gorm.DB) error {
	scanner := bufio.NewScanner(bytes.NewReader(sql.EmbeddedSQLData))
	var statement strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "--") { // 忽略空行和注释
			continue
		}
		statement.WriteString(line)
		if strings.HasSuffix(strings.TrimSpace(line), ";") {
			execSQL := statement.String()[:statement.Len()-1] // 移除末尾的分号
			if err := db.Exec(execSQL).Error; err != nil {
				return err
			}
			statement.Reset() // 重置字符串构建器
		}
	}
	return scanner.Err()
}
