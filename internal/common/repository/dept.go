package repository

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/services/admin/models"
	"go-file-server/pkgs/base"

	"gorm.io/gorm"
)

type DeptRepository struct {
	Repo *core.Repo
}

func NewDeptRepository(db *gorm.DB) *DeptRepository {
	return &DeptRepository{Repo: core.NewRepo(db)}
}

func WithByDeptId(id int) base.DbScope {
	return base.WithQuery("dept_id = ?", id)
}

func (r *DeptRepository) FindOne(opts ...base.DbScope) (sysDepts models.SysDept, err error) {
	err = r.Repo.FindOne(&sysDepts, opts...)
	return
}

func (r *DeptRepository) Find(opts ...base.DbScope) (sysDepts []models.SysDept, err error) {
	err = r.Repo.Find(&sysDepts, opts...)
	return
}

func (r *DeptRepository) Create(value *models.SysDept) error {
	return r.Repo.Create(value)
}

func (r *DeptRepository) Save(value *models.SysDept) error {
	return r.Repo.Save(value)
}

func WithDeptIds(ids ...int) base.DbScope {
	return base.WithQuery("dept_id in ?", ids)
}

func WithDeptName(name string) base.DbScope {
	return base.WithQuery("dept_name = ?", name)

}

func (r *DeptRepository) Delete(opts ...base.DbScope) error {
	return r.Repo.Delete(&models.SysDept{}, opts...)
}

func (r *DeptRepository) Update(updateFunc func(*models.SysDept), opts ...base.DbScope) error {
	sysDept := &models.SysDept{}
	updateFunc(sysDept)
	return r.Repo.Update(sysDept, opts...)
}

func WithLikeDeptPath(deptPath string) base.DbScope {
	return base.WithQuery("`sys_dept`.`dept_path` LIKE ?", "%"+deptPath+"%")
}

func (r *DeptRepository) GetFilteredDeptIdsForRole(roleId int) ([]int, error) {
	var deptIds []int
	if err := r.Repo.GetDB().Table("sys_role_dept").
		Select("sys_role_dept.dept_id").
		Joins("LEFT JOIN sys_dept on sys_dept.dept_id = sys_role_dept.dept_id").
		Where("role_id = ?", roleId).
		Where(`sys_role_dept.dept_id NOT IN (SELECT sys_dept.parent_id FROM sys_role_dept
		 LEFT JOIN sys_dept ON sys_dept.dept_id = sys_role_dept.dept_id WHERE role_id = ?)`, roleId).
		Find(&deptIds).Error; err != nil {
		return nil, err
	}
	return deptIds, nil
}
