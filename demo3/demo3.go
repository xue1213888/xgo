package main

import (
	"context"
	"log"
	"time"

	"github.com/xue1213888/xgo"
)

// 在 demo1 的基础上增加了一次性开启的协程数量，还带了一个协程名，当协程panic的时候外层会通过协程名来打印错误
func main() {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*3)
	x := xgo.New(ctx)
	x.Run(func() {
		for i := 0; i < 100000; i++ {
			log.Printf("%d\n", i)
			time.Sleep(time.Second)
			if x.IsDone() {
				return
			}
		}
	}, xgo.WithGoroutineName("goroutine name"), xgo.WithConcurrentNum(10))

	x.Wait()
	if err := x.Error(); err != nil {
		log.Printf("err: %v\n", err)
	}
	log.Printf("over\n")
}
