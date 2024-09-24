package base

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DbScope = func(db *gorm.DB) *gorm.DB

func WithQuery(query string, args ...interface{}) DbScope {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(query, args...)
	}
}

func WithModel(v any) DbScope {
	return func(db *gorm.DB) *gorm.DB {
		return db.Model(v)
	}
}

func WithSelect(query string, args ...interface{}) DbScope {
	return func(db *gorm.DB) *gorm.DB {
		return db.Select(query, args...)
	}
}

func WithOrderBy(sort string, bl bool) DbScope {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order(clause.OrderByColumn{Column: clause.Column{Name: sort}, Desc: bl})
	}
}

func WithPreload(query string, args ...any) DbScope {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload(query, args)
	}
}

func WithJoins(query string, args ...any) DbScope {
	return func(db *gorm.DB) *gorm.DB {
		return db.Joins(query, args)
	}
}

func WithPaginate(pageIndex, pageSize int) DbScope {
	return func(db *gorm.DB) *gorm.DB {
		offset := (pageIndex - 1) * pageSize
		if offset < 0 {
			offset = 0
		}
		if pageSize <= 0 {
			pageSize = 10
		}
		return db.Offset(offset).Limit(pageSize)
	}
}

func WithFullAssociations() DbScope {
	return func(db *gorm.DB) *gorm.DB {
		return db.Session(&gorm.Session{FullSaveAssociations: true})
	}
}
