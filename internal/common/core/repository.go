package core

import (
	"go-file-server/internal/common/global"
	"go-file-server/pkgs/base"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db: db}
}
func (r *Repo) GetDB() *gorm.DB {
	return r.db
}

func (r *Repo) Find(dest interface{}, opts ...base.DbScope) error {
	query := r.db.Scopes(opts...)
	return query.Find(dest).Error
}

func (r *Repo) FindWithCount(dest interface{}, count *int64, opts ...base.DbScope) error {
	query := r.db.Scopes(opts...)
	return query.Find(dest).Offset(-1).Limit(-1).Count(count).Error
}

func (r *Repo) FindOne(dest interface{}, opts ...base.DbScope) error {
	query := r.db.Scopes(opts...)
	return query.First(dest).Error
}

func (r *Repo) Delete(value interface{}, opts ...base.DbScope) error {
	if len(opts) == 0 {
		return errors.Errorf(global.ErrEmptyOptsForGromDelete)
	}

	return r.WithTransaction(func(tx *gorm.DB) error {
		query := tx.Scopes(opts...)
		return query.Delete(value).Error
	})

}
func (r *Repo) DeleteWithAssociation(mode any, column string, values any, opts ...base.DbScope) error {
	return r.WithTransaction(func(tx *gorm.DB) error {
		query := tx.Scopes(opts...)
		return query.Model(mode).Association(column).Delete(values)
	})

}

func (r *Repo) Update(value interface{}, opts ...base.DbScope) error {

	if len(opts) == 0 {
		return errors.Errorf(global.ErrEmptyOptsForGromUptdae)
	}

	return r.WithTransaction(func(tx *gorm.DB) error {
		query := tx.Scopes(opts...)
		return query.Updates(value).Error
	})

}

func WithClauses(columns ...string) func(...string) base.DbScope {
	return func(assignmentColumns ...string) base.DbScope {
		return func(db *gorm.DB) *gorm.DB {
			clauseColumns := make([]clause.Column, len(columns))
			for i, name := range columns {
				clauseColumns[i] = clause.Column{Name: name}
			}
			onConflictClause := clause.OnConflict{
				Columns:   clauseColumns,
				DoUpdates: clause.AssignmentColumns(assignmentColumns),
			}
			return db.Clauses(onConflictClause)
		}

	}
}

func (r *Repo) Create(value interface{}, opts ...base.DbScope) error {
	return r.WithTransaction(func(tx *gorm.DB) error {
		newTx := tx.Scopes(opts...)
		return newTx.Create(value).Error
	})
}

func (r *Repo) Save(value interface{}, opts ...base.DbScope) error {
	return r.WithTransaction(func(tx *gorm.DB) error {
		newTx := tx.Scopes(opts...)
		return newTx.Save(value).Error
	})
}

func (r *Repo) WithTransaction(fc func(tx *gorm.DB) error) error {
	return r.db.Transaction(
		fc,
	)
}
