package main

import (
	"context"
	"log"
	"time"

	"github.com/xue1213888/xgo"
)

// 在 demo1 的基础上增加了panic
func main() {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*3)
	x := xgo.New(ctx)
	x.Run(func() {
		for i := 0; i < 100000; i++ {
			if i == 1 {
				panic("aaaaa")
			}
			log.Printf("%d\n", i)
			time.Sleep(time.Second)
			if x.IsDone() {
				return
			}
		}
	})

	x.Wait()
	if err := x.Error(); err != nil {
		log.Printf("err: %v\n", err)
	}
	log.Printf("over\n")

}
