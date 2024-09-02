package repository

import (
	"fmt"
	"go-file-server/pkgs/pathtool"
	"log"
	"testing"
	"time"
)

func TestNewFsRepository(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		name string
		args args
		want *FsRepository
	}{
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			indexer, err := pathtool.NewFileIndexer("/tmp/basedir")
			if err != nil {
				log.Fatal(err)
			}
			repo := NewFsRepository(indexer)

			for {
				time.Sleep(3 * time.Second)
				docs, total, err := repo.Find(WithIsDir(false), WithPagination(1, 4))
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println(total, docs)
			}

		})
	}
}
