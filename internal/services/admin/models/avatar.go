package models

type Avatar struct {
	ID     int     `gorm:"primaryKey"`
	Data   []byte  `gorm:"type:longblob;not null"`
	UserID int     `gorm:"uniqueIndex"`
	User   SysUser `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:UserID;references:UserId"`
}

func (*Avatar) TableName() string {
	return "avatar"
}
