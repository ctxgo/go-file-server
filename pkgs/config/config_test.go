package config

import (
	"fmt"
	"testing"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{}, // TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init(SetAutomaticEnv(), SetFile("your config file path"))
			fmt.Println(ApplicationCfg)
			fmt.Println(DatabaseCfg)
			fmt.Println(JwtCfg)
			fmt.Println(CacheCfg)
			fmt.Println(OAuthCfg)

		})
	}
}
