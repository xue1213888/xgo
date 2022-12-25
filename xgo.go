package xgo

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

type Xgo interface {
	Run(func(), ...Option)
	Wait()
	Cancel()
	IsDone() bool
	Error() error
}

type goroutiner struct {
	waitGroup *sync.WaitGroup // 内部的等待组
	parentCtx context.Context
	ctx       context.Context    // 内部的ctx
	cancel    context.CancelFunc // 内部的取消函数
	err       error              // 第一个出发panic的goroutine
	closed    int32              // 判断管道有没有被关闭，整个生命周期其实就已经结束了，因为其中一个goroutine触发了的panic
}

func New(ctx context.Context) Xgo {
	nctx, cancel := context.WithCancel(ctx)
	return &goroutiner{
		waitGroup: &sync.WaitGroup{},
		parentCtx: ctx,
		ctx:       nctx,
		cancel:    cancel,
		closed:    0,
	}
}

type runConfig struct {
	goroutineName string
	concurrentNum int
	clearup       []func()
}

type Option func(g *runConfig)

func WithGoroutineName(name string) Option {
	return func(g *runConfig) {
		g.goroutineName = name
	}
}

func WithConcurrentNum(num int) Option {
	return func(g *runConfig) {
		g.concurrentNum = num
	}
}

func WithClearup(f func()) Option {
	return func(g *runConfig) {
		g.clearup = append(g.clearup, f)
	}
}

func (g *goroutiner) Run(f func(), options ...Option) {
	runConfig := new(runConfig)
	for i := 0; i < len(options); i++ {
		options[i](runConfig)
	}
	n := "unknow"
	if len(runConfig.goroutineName) != 0 {
		n = runConfig.goroutineName
	}
	num := 1
	if runConfig.concurrentNum > 1 {
		num = runConfig.concurrentNum
	}
	wg := &sync.WaitGroup{}
	wg.Add(num)
	g.waitGroup.Add(num)
	for i := 0; i < num; i++ {
		go func() {
			defer func() {
				if err := recover(); err != nil {
					// 如果函数有错误就返回错误信息
					if atomic.CompareAndSwapInt32(&(g.closed), 0, 1) {
						g.Cancel()
						g.err = fmt.Errorf("%s: %v", n, err)
					}
				}
				g.waitGroup.Done()
				wg.Done()
			}()
			f()
		}()
	}
	go func() {
		wg.Wait()
		for i := 0; i < len(runConfig.clearup); i++ {
			runConfig.clearup[i]()
		}
	}()
}

func (g *goroutiner) Wait() {
	g.waitGroup.Wait()
}

func (g *goroutiner) Cancel() {
	g.cancel()
}

func (g *goroutiner) IsDone() bool {
	select {
	case _, ok := <-g.ctx.Done():
		if !ok {
			return true
		}
	default:
		return false
	}
	return false
}

func (g *goroutiner) Error() error {
	if g.err != nil {
		return g.err
	} else if g.parentCtx.Err() != nil {
		return g.parentCtx.Err()
	} else {
		return g.ctx.Err()
	}
}
