package models

import "go-file-server/internal/common/models"

type UserToken struct {
	ID      int    `json:"id" gorm:"primaryKey"`
	Token   string `json:"token" gorm:"not null"`
	UserID  int    `json:"user_id"`
	Remark  string `json:"remark"`
	models.ModelTime
	User SysUser `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:UserID;references:UserId"`
}

func (*UserToken) TableName() string {
	return "user_token"
}
