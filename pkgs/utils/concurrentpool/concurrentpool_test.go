package concurrentpool

import (
	"go-file-server/pkgs/zlog"
	"log"
	"testing"
	"time"

	"github.com/panjf2000/ants/v2"
)

func TestNewAntsPool(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		name    string
		args    args
		want    *AntsPool
		wantErr bool
	}{
		{}, // TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool, err := NewAntsPool(WithPoolSize(6), WithAntsOptions(ants.WithNonblocking(false)))
			if err != nil {
				log.Fatal(err)
			}
			defer pool.Release()

			for i := 0; i < 5; i++ {
				i := i                      // 为闭包捕获变量
				err := pool.Submit(func() { //Parent 0、1、2、3、4
					log.Printf("Parent Task %d is running\n", i)
					time.Sleep(1 * time.Second) //  父任务持续时间

					ChildPool, err := pool.ForkChildPool(WithChildPoolId("1"), WithPoolSize(10))
					if err != nil {
						log.Printf("init ChildPool err: %v", err)
					}

					// 在父任务中启动子任务
					for _j := 0; _j < 10; _j++ {
						j := _j // 为闭包捕获变量

						err := ChildPool.Submit(func() { // Child
							log.Printf("Child Task %d of Parent Task %d is running\n", j, i)
							time.Sleep(1 * time.Second) // 子任务持续时间
						})

						if err != nil {
							log.Printf("Error submitting child task %d of parent task %d: %v", j, i, err)
						}

					}
					ChildPool.Wait()
					//time.Sleep(2 * time.Second) //  父任务持续时间

				})

				if err != nil {
					zlog.SugLog.Error(err)
				}
			}

			pool.Wait()
			//for x := range pool.ErrorsChan {
			//	fmt.Println(x)
			//}
			log.Println("All tasks are done!")
		})

	}
}
