package models

import (
	"go-file-server/internal/common/models"

	"gorm.io/gorm"
)

type SysLoginLog struct {
	gorm.Model
	Username      string `json:"username" gorm:"size:128;comment:用户名"`
	Status        string `json:"status" gorm:"size:4;comment:状态"`
	Ipaddr        string `json:"ipaddr" gorm:"size:255;comment:ip地址"`
	LoginLocation string `json:"loginLocation" gorm:"size:255;comment:归属地"`
	Browser       string `json:"browser" gorm:"size:255;comment:浏览器"`
	Os            string `json:"os" gorm:"size:255;comment:系统"`
	Platform      string `json:"platform" gorm:"size:255;comment:固件"`
	Remark        string `json:"remark" gorm:"size:255;comment:备注"`
	Msg           string `json:"msg" gorm:"size:255;comment:信息"`
	models.ControlBy
}

func (*SysLoginLog) TableName() string {
	return "sys_login_log"
}
