package plugin

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type RetryPlugin struct {
	MaxAttempts          int
	DelayBetweenAttempts time.Duration
}

func (p *RetryPlugin) Name() string {
	return "retryPlugin"
}

func (p *RetryPlugin) Initialize(db *gorm.DB) error {
	db.Callback().Create().Replace(
		"gorm:create",
		p.wrapWithRetry("create", db.Callback().Create().Get("gorm:create")),
	)

	db.Callback().Update().Replace(
		"gorm:update",
		p.wrapWithRetry("update", db.Callback().Update().Get("gorm:update")),
	)

	db.Callback().Delete().Replace(
		"gorm:delete",
		p.wrapWithRetry("delete", db.Callback().Delete().Get("gorm:delete")),
	)

	db.Callback().Query().Replace(
		"gorm:query",
		p.wrapWithRetry("query", db.Callback().Query().Get("gorm:query")),
	)

	return nil
}

func (p *RetryPlugin) wrapWithRetry(operation string, originalFunc func(*gorm.DB)) func(*gorm.DB) {
	return func(scope *gorm.DB) {
		for attempt := 1; attempt <= p.MaxAttempts; attempt++ {
			originalFunc(scope)
			if scope.Error == nil || errors.Is(scope.Error, gorm.ErrRecordNotFound) {
				return
			}
			tableName := ""
			if scope.Statement != nil && scope.Statement.Schema != nil {
				tableName = scope.Statement.Schema.Name // 获取表名
			}
			if code, b := isConnectionError(scope.Error); b {
				scope.Logger.Error(context.Background(), "%s on table '%s' failed state is %s, attempt %d of %d. Retrying in %v...\n",
					operation, tableName, code, attempt, p.MaxAttempts, p.DelayBetweenAttempts)
				time.Sleep(p.DelayBetweenAttempts)
				scope.Error = nil
			} else {
				scope.Logger.Error(context.Background(), "%s on table '%s' failed, error: %v", operation, tableName, scope.Error)
				return
			}
		}
	}
}

// isConnectionError checks if the error is a connection-related error.
func isConnectionError(err error) (string, bool) {
	switch e := err.(type) {
	case *pgconn.PgError:
		return isPostgresConnectionError(e)
	case *mysql.MySQLError:
		return isMySQLConnectionError(e)
	case net.Error:
		return e.Error(), true
	default:
		if strings.Contains(err.Error(), "connection refused") {
			return "Connection refused", true
		}
	}
	return "", false
}

// Global map for retryable PostgreSQL error codes
var retryablePostgresErrorCodes = map[string]string{
	"08000": "Connection exception",
	"08001": "SQL client unable to establish SQL connection",
	"08003": "Connection does not exist",
	"08004": "SQL server rejected connection",
	"08006": "Connection failure",
	"08007": "Transaction resolution unknown",
	"53300": "Too many clients already",
	"57P01": "Admin shutdown",
}

func isPostgresConnectionError(pgErr *pgconn.PgError) (string, bool) {

	if description, exists := retryablePostgresErrorCodes[pgErr.Code]; exists {
		return fmt.Sprintf("code: %s, description: %s", pgErr.Code, description), true
	}
	return "", false
}

var retryableMySQLErrorCodes = map[uint16]string{
	2002: "Cannot connect to MySQL server",
	2003: "Can't connect to MySQL server on specified port",
	2006: "MySQL server has gone away",
	2013: "Lost connection to MySQL server during query",
	2024: "SSL connection error",
	2048: "Invalid connection attributes",
	2055: "Lost connection to MySQL server at reading authorization packet",
}

func isMySQLConnectionError(mysqlErr *mysql.MySQLError) (string, bool) {

	if description, exists := retryableMySQLErrorCodes[mysqlErr.Number]; exists {
		return fmt.Sprintf("code: %d,description: %s", mysqlErr.Number, description), true
	}
	return "", false
}
