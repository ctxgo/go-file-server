package models

import (
	"go-file-server/internal/common/models"
)

type FsPermissions struct {
	Path        string   `json:"path" binding:"required"`
	Permissions []string `json:"permissions" binding:"required"`
}

type SysRole struct {
	RoleId    int             `json:"roleId" gorm:"primaryKey;autoIncrement"`    // 角色编码
	RoleName  string          `json:"roleName" gorm:"size:128;unique;not null;"` // 角色名称
	Status    string          `json:"status" gorm:"size:4;"`                     //
	RoleKey   string          `json:"roleKey" gorm:"size:128;"`                  //角色代码
	RoleSort  int             `json:"roleSort" gorm:""`                          //角色排序
	Flag      string          `json:"flag" gorm:"size:128;"`                     //
	Remark    string          `json:"remark" gorm:"size:255;"`                   //备注
	Admin     bool            `json:"admin" gorm:"size:4;"`
	DataScope string          `json:"dataScope" gorm:"size:128;"`
	Params    string          `json:"params" gorm:"-"`
	RateLimit uint64          `json:"rateLimit"` // 文件传输限速
	MenuIds   []int           `json:"menuIds" gorm:"-"`
	DeptIds   []int           `json:"deptIds" gorm:"-"`
	SysDept   []SysDept       `json:"sysDept" gorm:"many2many:sys_role_dept;foreignKey:RoleId;joinForeignKey:role_id;references:DeptId;joinReferences:dept_id;"`
	SysMenu   []SysMenu       `json:"sysMenu" gorm:"many2many:sys_role_menu;foreignKey:RoleId;joinForeignKey:role_id;references:MenuId;joinReferences:menu_id;"`
	FsRoles   []FsPermissions `json:"fsRoles" gorm:"-"`
	models.ControlBy
	models.ModelTime
}

func (SysRole) TableName() string {
	return "sys_role"
}
