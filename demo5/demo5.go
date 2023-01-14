package main

import (
	"log"
	"time"

	"github.com/xue1213888/xgo"
)

// 生产者运行一段时间会发生panic，管道也会关闭
func main() {
	// ctx, _ := context.WithTimeout(context.Background(), time.Second*3)
	x := xgo.New(xgo.WithTimeout(time.Second * 3))
	data := make(chan int, 50)
	x.Run(func() {
		for i := 0; i < 100000; i++ {
			if i == 1 {
				panic("生产数据方出现错误")
			}
			data <- i
			time.Sleep(time.Second)
			if x.IsDone() {
				return
			}
		}
	}, xgo.RunWithName("producer"), xgo.RunWithNum(10), xgo.RunWithClearup(func() {
		log.Printf("close chan")
		close(data)
	}))

	x.Run(func() {
		for d := range data {
			log.Printf("consumer %d", d)
		}
	}, xgo.RunWithName("consumer"), xgo.RunWithNum(20))

	x.Wait()
	if err := x.Error(); err != nil {
		log.Printf("err: %v\n", err)
	}
	log.Printf("over\n")
}
