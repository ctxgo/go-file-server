package base

import (
	"fmt"
	"testing"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Foo struct {
	FID  int    `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"size:255;not null"`
}

type Bar struct {
	BID    int    `gorm:"primaryKey;autoIncrement"`
	FooID  int    `gorm:"index"`
	Foo    Foo    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:FooID;references:FID"`
	Detail string `gorm:"size:255;not null"`
}

func TestInitDatabase(t *testing.T) {
	type args struct {
		dsn  string
		open DriverOpen
		opts []Option
	}
	tests := []struct {
		name    string
		args    args
		want    *gorm.DB
		wantErr bool
	}{
		{}, // TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dns := `root:yourpassword@tcp(127.0.0.1:3306)/dev?charset=utf8&parseTime=True&loc=Local&timeout=1000ms`
			driverOpen, err := GetDriverOpen("mysql")
			if err != nil {
				fmt.Println(1, err)
				return
			}

			db, err := InitDatabase(dns, driverOpen, SetLogger(
				logger.Default.LogMode(logger.Info),
			))
			if err != nil {
				fmt.Println(2, err)
				return
			}

			err = db.AutoMigrate(&Foo{}, &Bar{})

			fmt.Println(3, err)

			db.Session(&gorm.Session{DryRun: true}).
				Where("name = ?", "John").
				Joins("join foo").
				Find(&Foo{})

		})
	}
}
