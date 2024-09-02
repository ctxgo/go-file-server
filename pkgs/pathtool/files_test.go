package pathtool

import (
	"fmt"
	"testing"
)

func TestIterDirectory(t *testing.T) {

	tests := []struct {
		name string
	}{
		{"asd"}, // TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := "/tmp"
			item := NewFiletool(dir).IterateFiles()
			for x := range item {
				fmt.Println(x)
			}
		})
	}
}

func TestMvproject(t *testing.T) {

	tests := []struct {
		name string
	}{
		{}, // TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//got := Mvproject("/tmp/",
			//	&model.Parse_request_mv{Filelist: []string{"/test"}, Dstdir: "/asd"})
			//fmt.Println(got)
			//cc := got[0].err.(*os.LinkError).Err.Error()

		})
	}
}
