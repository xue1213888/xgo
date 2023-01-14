package main

import (
	"log"
	"time"

	"github.com/xue1213888/xgo"
)

// 简单启动一个协程，拥有超时时间，并等待退出
func main() {
	// ctx, _ := context.WithTimeout(context.Background(), time.Second*3)
	x := xgo.New(xgo.WithTimeout(time.Second * 3))
	x.Run(func() {
		for i := 0; i < 100000; i++ {
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
