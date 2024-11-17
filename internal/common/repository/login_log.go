package repository

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/services/admin/models"
	"go-file-server/pkgs/base"

	"gorm.io/gorm"
)

type LoginLogRepository struct {
	Repo *core.Repo
}

func NewLoginLogRepository(db *gorm.DB) *LoginLogRepository {
	return &LoginLogRepository{Repo: core.NewRepo(db)}
}

func (r *LoginLogRepository) Create(log *models.SysLoginLog) error {
	return r.Repo.Create(log)
}

func WithLoginPaginateById(pageSize int, pageIndex int) base.DbScope {
	return func(db *gorm.DB) *gorm.DB {
		return db.Scopes(
			base.WithOrderBy("id", true),
			base.WithPaginate(pageSize, pageIndex),
		)
	}

}

func WithLoginUsername(s string) base.DbScope {
	return base.WithQuery("username = ?", s)
}

func (r *LoginLogRepository) Find(opts ...base.DbScope) (logs []models.SysLoginLog, c int64, err error) {
	err = r.Repo.FindWithCount(&logs, &c, opts...)
	return
}

func WithLoginStatus(s string) base.DbScope {
	return base.WithQuery("status = ?", s)
}

func WithLoginIpaddr(s string) base.DbScope {
	return base.WithQuery("ipaddr = ?", s)
}

func WithLoginIds(ids ...int) base.DbScope {
	return base.WithQuery("id in ?", ids)

}

func (r *LoginLogRepository) Delete(opts ...base.DbScope) error {
	return r.Repo.Delete(&models.SysLoginLog{}, opts...)
}
