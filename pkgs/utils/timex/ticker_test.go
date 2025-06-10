package timex

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNewImmediateTicker(t *testing.T) {
	type args struct {
		interval time.Duration
	}
	tests := []struct {
		name string
		args args
		want *ImmediateTicker
	}{
		{args: args{interval: 500 * time.Microsecond}}, // TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewImmediateTicker(tt.args.interval)
			timeOutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			for {
				select {
				case <-timeOutCtx.Done():
					time.Sleep(1 * time.Second)
					got.Stop()
					time.Sleep(5 * time.Second)

					return
				case <-got.C:
					fmt.Println("tick")
				}
			}

		})
	}
}
