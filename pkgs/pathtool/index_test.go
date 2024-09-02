package pathtool

import (
	"fmt"
	"log"
	"os"
	"runtime/trace"
	"testing"
	"time"

	"github.com/blevesearch/bleve/v2"
)

func TestNewFileIndexer(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *FileIndexer
		wantErr bool
	}{
		{}, // TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			f, _ := os.Create("trace.out")
			defer f.Close()

			// 开始 CPU 分析
			if err := trace.Start(f); err != nil {
				log.Fatalf("failed to start trace: %v", err)
			}
			st := time.Now()
			fileIndexer, err := NewFileIndexer("/tmp/testbasedir", WithStorageType(UseDisk))
			if err != nil {
				log.Fatalf("%+v", err)
			}
			// 可以根据需求调用 RandomSearch 和 Search 方法
			log.Printf("index done, time: %f", time.Now().Sub(st).Seconds())
			trace.Stop()

			go func() {

				for {
					query := bleve.NewPrefixQuery("/tmp/testbasedir/")
					query.SetField("Path")
					searchRequest := bleve.NewSearchRequest(query)
					searchRequest.Fields = []string{"*"} // 请求加载所有字段

					results, err := fileIndexer.Index.Search(searchRequest)
					if err != nil {
						log.Fatalf("Search failed: %v", err)
					}
					fmt.Printf("Total hits: %d\n", results.Total)
					for _, hit := range results.Hits {
						fmt.Printf("Found file at %s\n", hit.ID)
					}
					time.Sleep(1 * time.Second)
				}
			}()

			//os.Rename("/tmp/testbasedir/aa2", "/tmp/testbasedir/aa3/aa2")
			time.Sleep(1 * time.Second)
			os.Remove("/tmp/testbasedir/aa2")
			fileIndexer.DelResource("/tmp/testbasedir/aa2")
			os.MkdirAll("/tmp/testbasedir/aa3/aa2", 0755)
			time.Sleep(3 * time.Second)
			fileIndexer.AddResource("/tmp/testbasedir/aa3/aa2")
			time.Sleep(3 * time.Second)

		})
	}
}
