package repository

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/services/admin/models"
	"go-file-server/pkgs/base"

	"gorm.io/gorm"
)

type MenuRepository struct {
	Repo *core.Repo
}

func NewMenuRepository(db *gorm.DB) *MenuRepository {
	return &MenuRepository{Repo: core.NewRepo(db)}
}

func WithMenuTypes(menuTypes ...string) base.DbScope {
	return base.WithQuery("menu_type IN ?", menuTypes)
}

func WithMenuTitle(s string) base.DbScope {
	return base.WithQuery("title = ?", s)
}

func WithMenuVisible(d int) base.DbScope {
	return base.WithQuery("visible = ?", d)
}

func WithDeletedAtIsNull(s bool) base.DbScope {
	if s {
		return base.WithQuery("deleted_at IS NULL")

	}
	return base.WithQuery("deleted_at IS NOT NULL")
}

func WithOrderByMenuSort(bl bool) base.DbScope {
	return base.WithOrderBy("sort", bl)
}

func WithPreloadSysApi() base.DbScope {
	return base.WithPreload("SysApi")
}
func (r *MenuRepository) Create(value *models.SysMenu) error {
	return r.Repo.Create(value)
}

func (r *MenuRepository) Delete(opts ...base.DbScope) error {
	return r.Repo.Delete(&models.SysMenu{}, opts...)
}

func (r *MenuRepository) Save(value *models.SysMenu) error {
	return r.Repo.Save(value)
}

func (r *MenuRepository) Update(updateFunc func(*models.SysMenu), opts ...base.DbScope) error {
	sysMenu := &models.SysMenu{}
	updateFunc(sysMenu)
	return r.Repo.Update(&models.SysMenu{}, opts...)
}

func (r *MenuRepository) DelWithAssociationSysApi(data *models.SysMenu) error {
	return r.Repo.DeleteWithAssociation(data, "SysApi", data.SysApi)
}

func (r *MenuRepository) Find(opts ...base.DbScope) (menus models.SysMenuSlice, err error) {
	err = r.Repo.Find(&menus, opts...)
	return
}

func (r *MenuRepository) FindOne(opts ...base.DbScope) (menu models.SysMenu, err error) {
	err = r.Repo.Find(&menu, opts...)
	return
}
