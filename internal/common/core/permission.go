package core

import (
	"go-file-server/internal/common/global"
	"go-file-server/pkgs/base"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type DataPermission struct {
	DataScope string
	UserId    int
	DeptId    int
	RoleId    int
}

func PermissionAction(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var p = new(DataPermission)
		var err error
		if userId := GetUserId(c); userId != 0 {
			p, err = newDataPermission(db, userId)
			if err != nil {
				ErrRep().
					SetMsg("数据权限鉴定错误").
					SendGin(c)
				return
			}
		}
		c.Set(global.PermissionKey, p)
	}
}

func ExtractPermission(c *gin.Context) *DataPermission {
	p, exists := c.Get(global.PermissionKey)
	if !exists {
		return &DataPermission{}
	}
	return p.(*DataPermission)
}

func newDataPermission(tx *gorm.DB, userId interface{}) (*DataPermission, error) {
	var err error
	p := &DataPermission{}

	err = tx.Table("sys_user").
		Select("sys_user.user_id", "sys_role.role_id", "sys_user.dept_id", "sys_role.data_scope").
		Joins("left join sys_role on sys_role.role_id = sys_user.role_id").
		Where("sys_user.user_id = ?", userId).
		Scan(p).Error
	if err != nil {
		err = errors.Errorf("获取用户数据出错 msg:" + err.Error())
		return nil, err
	}
	return p, nil
}

func WithDeptDbScope(c *gin.Context) base.DbScope {
	return func(db *gorm.DB) *gorm.DB {
		p := ExtractPermission(c)
		switch p.DataScope {
		case "2":
			//return db.Where("dept_id in (select dept_id from sys_role_dept  where sys_role_dept.role_id = ?)", p.RoleId)
			return db.Where(`dept_id in (SELECT sd.dept_id
			FROM sys_dept sd
			JOIN (
				SELECT dept_id, dept_path 
				FROM sys_dept
				WHERE dept_id IN (SELECT dept_id FROM sys_role_dept WHERE role_id = ?)
			) as parent_dept ON sd.dept_path LIKE CONCAT(parent_dept.dept_path, '%');
			)`, p.RoleId)
		case "3", "5":
			return db.Where("dept_id = ? ", p.DeptId)
		case "4":
			return db.Where("dept_path like ? ", "%/"+strconv.Itoa(p.DeptId)+"/%")
		default:
			return db
		}
	}
}

func WithPermissionDbScope(tableName string, c *gin.Context) base.DbScope {
	return permissionDbScope("sys_user.create_by", c)
}
func WithUserPermissionDbScope(tableName string, c *gin.Context) base.DbScope {
	return permissionDbScope("sys_user.user_id", c)
}

func permissionDbScope(column string, c *gin.Context) base.DbScope {
	return func(db *gorm.DB) *gorm.DB {
		p := ExtractPermission(c)
		switch p.DataScope {
		case "2":
			return db.Where(column+" in (select sys_user.user_id from sys_role_dept left join sys_user on sys_user.dept_id=sys_role_dept.dept_id where sys_role_dept.role_id = ?)", p.RoleId)
		case "3":
			return db.Where(column+" in (SELECT user_id from sys_user where dept_id = ? )", p.DeptId)
		case "4":
			return db.Where(column+" in (SELECT user_id from sys_user where sys_user.dept_id in(select dept_id from sys_dept where dept_path like ? ))", "%/"+strconv.Itoa(p.DeptId)+"/%")
		case "5":
			return db.Where(column+" = ?", p.UserId)
		default:
			return db
		}
	}
}
