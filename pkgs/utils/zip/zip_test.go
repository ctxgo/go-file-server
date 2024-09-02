package zip

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNewFileZip(t *testing.T) {
	type args struct {
		outputPath string
		opts       []Options
	}
	tests := []struct {
		name string
		args args
		want Zipper
	}{
		{}, // TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := NewFileZip("/tmp/basedir/test/test.zip", WithVerbose(true))
			ctx, stop := context.WithTimeout(context.Background(), 1*time.Millisecond)
			defer stop()
			err := z.ZipWithCtx(ctx, "/tmp/basedir/bb", "/tmp/basedir/aa", "/tmp/basedir/hh")
			fmt.Println(err)
		})
	}
}
