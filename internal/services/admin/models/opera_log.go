package models

import (
	"go-file-server/internal/common/models"
	"time"

	"gorm.io/gorm"
)

type SysOperaLog struct {
	gorm.Model
	Title         string    `json:"title" gorm:"size:255;comment:操作模块"`
	BusinessType  string    `json:"businessType" gorm:"size:128;comment:操作类型"`
	BusinessTypes string    `json:"businessTypes" gorm:"size:128;comment:BusinessTypes"`
	Method        string    `json:"method" gorm:"size:128;comment:函数"`
	RequestMethod string    `json:"requestMethod" gorm:"size:128;comment:请求方式 GET POST PUT DELETE"`
	OperatorType  string    `json:"operatorType" gorm:"size:128;comment:操作类型"`
	OperName      string    `json:"operName" gorm:"size:128;comment:操作者"`
	DeptName      string    `json:"deptName" gorm:"size:128;comment:部门名称"`
	OperUrl       string    `json:"operUrl" gorm:"size:255;comment:访问地址"`
	OperIp        string    `json:"operIp" gorm:"size:128;comment:客户端ip"`
	OperLocation  string    `json:"operLocation" gorm:"size:128;comment:访问位置"`
	OperParam     string    `json:"operParam" gorm:"text;comment:请求参数"`
	Status        string    `json:"status" gorm:"size:4;comment:操作状态 1:正常 2:关闭"`
	OperTime      time.Time `json:"operTime" gorm:"comment:操作时间"`
	JsonResult    string    `json:"jsonResult" gorm:"size:255;comment:返回数据"`
	Remark        string    `json:"remark" gorm:"size:255;comment:备注"`
	LatencyTime   string    `json:"latencyTime" gorm:"size:128;comment:耗时"`
	UserAgent     string    `json:"userAgent" gorm:"size:255;comment:ua"`
	models.ControlBy
}

func (*SysOperaLog) TableName() string {
	return "sys_opera_log"
}
