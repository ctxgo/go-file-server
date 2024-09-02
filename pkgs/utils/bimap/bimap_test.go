package bimap

import (
	"fmt"
	"testing"
)

func TestNewBiMap(t *testing.T) {
	tests := []struct {
		name string
	}{
		{}, // TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			bimap := NewBiMap[string, string]()

			bimap.Insert("aa1", "bb1")
			bimap.Insert("aa1", "bbs")

			bimap.Insert("aa2", "bb2")
			bimap.Insert("aa2", "bb2")
			bimap.Insert("bb1", "bb2")
			v, ok := bimap.GetKey("aa2")
			fmt.Println(v, ok)
			k, ok := bimap.GetValue("bbs")
			fmt.Println(k, ok)
		})
	}
}
