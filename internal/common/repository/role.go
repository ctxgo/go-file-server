package repository

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/services/admin/models"
	"go-file-server/pkgs/base"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RoleRepository struct {
	Repo *core.Repo
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{Repo: core.NewRepo(db)}
}

func (r *RoleRepository) Create(values *models.SysRole) error {
	return r.Repo.Create(values)
}

func (r *RoleRepository) Update(updateFunc func(*models.SysRole), opts ...base.DbScope) error {
	sysRole := &models.SysRole{}
	updateFunc(sysRole)
	return r.Repo.Update(sysRole, opts...)
}

func (r *RoleRepository) Save(values *models.SysRole, opts ...base.DbScope) error {
	return r.Repo.Save(values, opts...)
}

func (r *RoleRepository) DelWithAssociationSysDept(data *models.SysRole) error {
	return r.Repo.DeleteWithAssociation(data, "SysDept", data.SysDept)
}

func (r *RoleRepository) DelWithAssociationSysMenu(data *models.SysRole) error {
	return r.Repo.DeleteWithAssociation(data, "SysMenu", data.SysMenu)
}

func WithPreloadSysMenu() base.DbScope {
	return base.WithPreload("SysMenu")
}

func WithPreloadSysDept() base.DbScope {
	return base.WithPreload("SysDept")
}

func WithMenuId(menuId int) base.DbScope {
	return base.WithQuery("menu_id = ?", menuId)
}

func WithMenuIds(menuIds ...int) base.DbScope {
	return base.WithQuery("menu_id in ?", menuIds)
}
func WithMenuName(name string) base.DbScope {
	return base.WithQuery("menu_name = ?", name)
}

func WithOrderByRoleSort(bl bool) base.DbScope {
	return base.WithOrderBy("sort", bl)
}

func WithPaginateByRoleId(pageSize int, pageIndex int) base.DbScope {
	return func(db *gorm.DB) *gorm.DB {
		return db.Scopes(
			base.WithOrderBy("role_id", false),
			base.WithPaginate(pageSize, pageIndex),
		)
	}

}

func WithRoleId(roleId int) base.DbScope {
	return base.WithQuery("role_id = ?", roleId)
}

func WithRoleKey(v string) base.DbScope {
	return base.WithQuery("role_key = ?", v)
}

func WithRoleName(v string) base.DbScope {
	return base.WithQuery("role_name = ?", v)
}

func WithRoleStatus(v string) base.DbScope {
	return base.WithQuery("status = ?", v)
}

func WithRoleRoleKey(v string) base.DbScope {
	return base.WithQuery("role_key = ?", v)
}

func WithRoleDataScope(v string) base.DbScope {
	return base.WithQuery("data_scope = ?", v)
}

func (r *RoleRepository) FindOne(opts ...base.DbScope) (role *models.SysRole, err error) {
	err = r.Repo.FindOne(&role, opts...)
	return
}

func (r *RoleRepository) Find(opts ...base.DbScope) (roles []models.SysRole, err error) {
	err = r.Repo.Find(&roles, opts...)
	return
}

func (r *RoleRepository) FindWithCount(opts ...base.DbScope) (roles []models.SysRole, c int64, err error) {
	err = r.Repo.FindWithCount(&roles, &c, opts...)
	return
}

func (r *RoleRepository) CascadeDelete(v *models.SysRole, opts ...base.DbScope) error {
	opts = append(opts, base.WithSelect(clause.Associations))
	return r.Repo.Delete(v, opts...)
}

func (r *RoleRepository) CascadeUpdate(updateSysRole *models.SysRole, opts ...base.DbScope) error {

	return r.Repo.WithTransaction(func(tx *gorm.DB) error {
		var model = models.SysRole{}
		var mlist = make([]models.SysMenu, 0)
		tx.Preload("SysMenu").First(&model, updateSysRole.RoleId)
		tx.Preload("SysApi").Where("menu_id in ?", updateSysRole.MenuIds).Find(&mlist)
		err := tx.Model(&model).Association("SysMenu").Delete(model.SysMenu)
		if err != nil {
			return err
		}
		updateSysRole.SysMenu = mlist
		return tx.Session(&gorm.Session{FullSaveAssociations: true}).Debug().Save(&updateSysRole).Error
	})

}
