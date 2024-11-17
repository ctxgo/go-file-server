package repository

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/services/admin/models"
	"go-file-server/pkgs/base"

	"gorm.io/gorm"
)

type OperaLogRepository struct {
	Repo *core.Repo
}

func NewOperaLogRepository(db *gorm.DB) *OperaLogRepository {
	return &OperaLogRepository{Repo: core.NewRepo(db)}
}

func WithOperaPaginateById(pageSize int, pageIndex int) base.DbScope {
	return func(db *gorm.DB) *gorm.DB {
		return db.Scopes(
			base.WithOrderBy("id", true),
			base.WithPaginate(pageSize, pageIndex),
		)
	}

}

func (r *OperaLogRepository) Create(log *models.SysOperaLog) error {
	return r.Repo.Create(log)
}

func WithOperaStatus(s string) base.DbScope {
	return base.WithQuery("status = ?", s)
}

func WithOperaBegin(s string) base.DbScope {
	return base.WithQuery("created_at >= ?", s)
}

func WithOperaEnd(s string) base.DbScope {
	return base.WithQuery("created_at <= ?", s)
}

func (r *OperaLogRepository) Find(opts ...base.DbScope) (logs []models.SysOperaLog, c int64, err error) {
	err = r.Repo.FindWithCount(&logs, &c, opts...)
	return
}

func WithOperaIds(ids ...int) base.DbScope {
	return base.WithQuery("id in ?", ids)

}

func (r *OperaLogRepository) Delete(opts ...base.DbScope) error {
	return r.Repo.Delete(&models.SysOperaLog{}, opts...)
}
