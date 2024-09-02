package casbin

import (
	"fmt"
	"go-file-server/pkgs/base"
	"testing"

	"github.com/casbin/casbin/v2"
	"gorm.io/gorm/logger"
)

func TestNewEnforcer(t *testing.T) {
	type args struct {
		opts []Option
	}
	tests := []struct {
		name    string
		args    args
		want    *casbin.CachedEnforcer
		wantErr bool
	}{
		{}, // TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := []Option{}
			dns := `root:yourpassword@tcp(127.0.0.1:3306)/dev?charset=utf8&parseTime=True&loc=Local&timeout=1000ms`
			driverOpen, err := base.GetDriverOpen("mysql")
			if err != nil {
				fmt.Println(1, err)
				return
			}

			db, err := base.InitDatabase(dns, driverOpen, base.SetLogger(
				logger.Default.LogMode(logger.Info),
			))
			if err != nil {
				fmt.Println(err)
				return
			}
			ops = append(ops, WithGormDB(db))

			casbinEnforcer, err := NewEnforcer(
				ops...,
			)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("query")
			data := casbinEnforcer.GetFilteredPolicy(0, "test")
			fmt.Println(data)
			data = casbinEnforcer.GetFilteredPolicy(0, "test")
			fmt.Println(data)
			// b, err := casbinEnforcer.Enforce("test", "/api/v1/fs/test/foo", "get")
			// fmt.Println(1, b, err)
			//b, err = casbinEnforcer.AddPolicy("test", "/api/v1/fs/test/foo.*", "get", "test")
			b, err := casbinEnforcer.AddNamedPolicies("p",
				[][]string{
					{"test", "/api/v1/fs/test/foo", "get", "test"},
				},
			)
			fmt.Println(2, b, err)

			b, err = casbinEnforcer.Enforce("test", "/api/v1/fs/test/foo", "get")
			fmt.Println(3, b, err)
			b, err = casbinEnforcer.RemoveFilteredPolicy(0, "test", "", "", "test")
			fmt.Println(4, b, err)
			b, err = casbinEnforcer.Enforce("test", "/api/v1/fs/test/foo", "get")
			fmt.Println(5, b, err)

		})
	}
}
