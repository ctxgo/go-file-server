package repository

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/services/admin/models"
	"go-file-server/pkgs/base"

	"gorm.io/gorm"
)

type UserRepository struct {
	Repo *core.Repo
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{Repo: core.NewRepo(db)}
}

func GetTableName() string {
	return (&models.SysUser{}).TableName()
}

func (r *UserRepository) Create(values *models.SysUser) error {
	return r.Repo.Create(values)
}

func (r *UserRepository) Delete(opts ...base.DbScope) error {
	return r.Repo.Delete(&models.SysUser{}, opts...)
}

func (r *UserRepository) Save(values *models.SysUser) error {
	return r.Repo.Save(values)
}

func (r *UserRepository) Update(updateFunc func(*models.SysUser), opts ...base.DbScope) error {
	sysUser := &models.SysUser{}
	updateFunc(sysUser)
	return r.Repo.Update(sysUser, opts...)
}

func (r *UserRepository) Updates(v any, opts ...base.DbScope) error {
	opts = append(opts, base.WithModel(&models.SysUser{}))
	return r.Repo.Update(v, opts...)
}

func WithUserId(userId int) base.DbScope {
	return base.WithQuery("user_id = ?", userId)
}

func WithUserSource(s string) base.DbScope {
	return base.WithQuery("source = ?", s)
}

func WithUserIds(userIds ...int) base.DbScope {
	return base.WithQuery("user_id in ?", userIds)
}

func (r *UserRepository) FindOne(opts ...base.DbScope) (user *models.SysUser, err error) {
	err = r.Repo.FindOne(&user, opts...)
	return
}

func WithUsername(name string) base.DbScope {
	return base.WithQuery("Username = ?", name)
}

func WithPhone(phone string) base.DbScope {
	return base.WithQuery("phone = ?", phone)
}

func WithUserStatus(s string) base.DbScope {
	return base.WithQuery("status = ?", s)
}

func WithPreloadDept() base.DbScope {
	return base.WithPreload("Dept")
}

// JoinWithDeptOnDeptID 添加与 `sys_dept` 的左连接。
// 使用此函数后，请确保后续查询中使用完全限定的列名以避免歧义。
func WithJoinDeptOnDeptID() base.DbScope {
	return base.WithJoins("left join `sys_dept` on `sys_dept`.`dept_id` = `sys_user`.`dept_id` ")
}

func WithPaginateByUserId(pageIndex int, pageSize int) base.DbScope {
	return func(db *gorm.DB) *gorm.DB {
		return db.Scopes(
			base.WithOrderBy("user_id", false),
			base.WithPaginate(pageIndex, pageSize),
		)
	}

}

func (r *UserRepository) Find(opts ...base.DbScope) (users []models.SysUser, c int64, err error) {
	err = r.Repo.FindWithCount(&users, &c, opts...)
	return
}
