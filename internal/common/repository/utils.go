package repository

import (
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
)

func isPostgresDuplicateError(err error) bool {
	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" // 23505 是 PostgreSQL 中唯一违反的错误代码
	}
	return false
}

func isMySQLDuplicateError(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062 // 1062 是 MySQL 中重复条目的错误代码
	}
	return false
}

func IsDuplicateError(dbType string, err error) bool {
	switch dbType {
	case "postgres":
		return isPostgresDuplicateError(err)
	case "mysql":
		return isMySQLDuplicateError(err)
	default:
		return false
	}
}
