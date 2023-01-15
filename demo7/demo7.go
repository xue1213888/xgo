package main

import (
	"log"
	"time"

	"github.com/xue1213888/xgo"
)

func main() {
	x := xgo.New()
	ch := make(chan int, 10)
	x.Run(func() {
		for i := 0; i < 10; i++ {
			log.Println("send:", i)
			ch <- i
			if x.IsDone() {
				break
			}
			time.Sleep(time.Millisecond)
		}
	}, xgo.RunWithClearup(func() {
		close(ch)
	}))

	x.Run(func() {
		i, ok := <-ch
		if !ok {
			log.Println("chan closed")
			return
		}
		log.Println("recv:", i)
	}, xgo.RunWithLimiter(xgo.NewSlidingWindowsLimiter(time.Second, 5, xgo.NewLimitUntilChanClosed(ch))),
		xgo.RunWithMaxGoroutineNum(10))

	x.Wait()
	if err := x.Error(); err != nil {
		log.Println(err)
	}
}
