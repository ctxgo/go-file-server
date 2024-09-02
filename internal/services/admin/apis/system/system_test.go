package system

import (
	"context"
	"go-file-server/pkgs/zlog"
	"reflect"
	"testing"
)

func Test_collectSystemDetails(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    *SystemDetails
		wantErr bool
	}{
		{args: args{ctx: context.Background()}}, // TODO: Add test cases
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zlog.Init()
			got, err := collectSystemDetails(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("collectSystemDetails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("collectSystemDetails() = %v, want %v", got, tt.want)
			}
		})
	}
}
