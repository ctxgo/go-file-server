package models

import (
	"go-file-server/internal/common/models"
)

// Menu 菜单中的类型枚举值
const (
	// Directory 目录
	Directory string = "M"
	// Menu 菜单
	Menu string = "C"
	// Button 按钮
	Button string = "F"
	// 路径
	Path string = "P"
)

type SysMenu struct {
	MenuId     int       `json:"menuId" gorm:"primaryKey;autoIncrement"`
	MenuName   string    `json:"menuName" gorm:"size:128;"`
	Title      string    `json:"title" gorm:"size:128;"`
	Icon       string    `json:"icon" gorm:"size:128;"`
	Path       string    `json:"path" gorm:"size:128;"`
	Paths      string    `json:"paths" gorm:"size:128;"`
	MenuType   string    `json:"menuType" gorm:"size:1;"`
	Action     string    `json:"action" gorm:"size:16;"`
	Permission string    `json:"permission" gorm:"size:255;"`
	ParentId   int       `json:"parentId" gorm:"size:11;"`
	NoCache    bool      `json:"noCache" gorm:"size:8;"`
	Breadcrumb string    `json:"breadcrumb" gorm:"size:255;"`
	Component  string    `json:"component" gorm:"size:255;"`
	Sort       int       `json:"sort" gorm:"size:4;"`
	Visible    string    `json:"visible" gorm:"size:1;"`
	IsFrame    string    `json:"isFrame" gorm:"size:1;DEFAULT:0;"`
	Apis       []int     `json:"apis" gorm:"-"`
	DataScope  string    `json:"dataScope" gorm:"-"`
	Params     string    `json:"params" gorm:"-"`
	RoleId     int       `gorm:"-"`
	Children   []SysMenu `json:"children,omitempty" gorm:"-"`
	IsSelect   bool      `json:"is_select" gorm:"-"`
	SysApi     []SysApi  `json:"sysApi" gorm:"many2many:sys_menu_api_rule"`

	models.ControlBy
	models.ModelTime
}

func (*SysMenu) TableName() string {
	return "sys_menu"
}

func (e *SysMenu) GetId() interface{} {
	return e.MenuId
}

type SysMenuSlice []SysMenu

func (x SysMenuSlice) Len() int           { return len(x) }
func (x SysMenuSlice) Less(i, j int) bool { return x[i].Sort < x[j].Sort }
func (x SysMenuSlice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

func (x SysMenuSlice) RemoveDuplicatesByKey(keyExtractor func(*SysMenu) any) SysMenuSlice {
	seen := make(map[any]bool)
	result := SysMenuSlice{}

	for _, item := range x {
		key := keyExtractor(&item)
		if !seen[key] {
			seen[key] = true
			result = append(result, item)
		}
	}

	return result
}
