package main

import (
	"log"
	"time"

	"github.com/xue1213888/xgo"
)

// 一个生产者消费者模型，生产者开了10个协程，消费者开了20个协程，生产者组在运行结束会关闭通信的管道，消费者自然而然地退出，不用监听点位，当然也可以监听，但是没有必要
func main() {
	// ctx, _ := context.WithTimeout(context.Background(), time.Second*3)
	x := xgo.New(xgo.WithTimeout(time.Second * 10))
	data := make(chan int, 50)
	x.Run(func() {
		for i := 0; i < 5; i++ {
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
