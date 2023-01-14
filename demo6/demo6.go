package main

import (
	"log"
	"time"

	"github.com/xue1213888/xgo"
)

func main() {
	// ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	x := xgo.New()

	ch := make(chan int, 20)
	x.Run(func() {
		for i := 0; i < 10; i++ {
			ch <- i
			time.Sleep(time.Second)
			if x.IsDone() {
				return
			}
		}
	}, xgo.RunWithClearup(func() {
		close(ch)
	}), xgo.RunWithNum(10))

	x.Run(func() {
		for v := range ch {
			log.Println(v)
			if x.IsDone() {
				return
			}
		}
	}, xgo.RunWithNum(10))

	x.Run(func() {
		for i := 0; i < 2; i++ {
			time.Sleep(time.Second * 3)
			x.Block(func() {
				time.Sleep(time.Second * 3)
			})
			if x.IsDone() {
				return
			}
		}
	})
	x.Wait()
	log.Println(x.Error())
}
