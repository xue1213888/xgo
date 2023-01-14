package main

import (
	"log"
	"time"

	"github.com/xue1213888/xgo"
)

// 在 demo1 的基础上增加了panic
func main() {
	x := xgo.New(xgo.WithTimeout(time.Second * 3))
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
	}, xgo.RunWithName("go"))

	x.Wait()
	if err := x.Error(); err != nil {
		log.Printf("err: %v\n", err)
	}
	log.Printf("over\n")

}
