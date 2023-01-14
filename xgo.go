package xgo

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

type Xgo interface {
	Run(func(), ...runOptions)
	Wait()
	Cancel()
	IsDone() bool
	Error() error
	Block(fn func()) error
}

type goroutiner struct {
	waitGroup     *sync.WaitGroup    // 内部的等待组
	ctx           context.Context    // 内部的ctx，内部会把上层的ctx包装一下，目的是为了拿到一个cancel函数
	cancel        context.CancelFunc // 内部的取消函数
	err           error              // 判断优先级高到低，第一个触发的panic，ctx取消，父ctx取消
	closed        int32              // 如果其中一个goroutine或者阻塞函数触发了panic，我们都会关闭整个组
	blockchan     chan struct{}      // 当前阻塞的信号通知
	blocked       int32              // 当前的阻塞状态
	unlimited     int32
	unlimitedChan chan struct{}
}

func New(o ...options) Xgo {
	g := new(goroutiner)
	g.waitGroup = new(sync.WaitGroup)
	g.unlimited = 0
	g.unlimitedChan = make(chan struct{})
	close(g.unlimitedChan)
	for _, v := range o {
		v(g)
	}
	if g.ctx == nil {
		ctx, cancel := context.WithCancel(context.Background())
		g.ctx = ctx
		g.cancel = cancel
	}
	return g
}

func (g *goroutiner) Run(f func(), options ...runOptions) {
	run := execRunOptions(options...)
	name := run.goroutineName
	concurrentNum := run.concurrentNum
	limiter := run.limiter
	if limiter != nil {
		if atomic.AddInt32(&g.unlimited, 1) == 1 {
			g.unlimitedChan = make(chan struct{})
		}
	}
	wg := &sync.WaitGroup{}
	for i := 0; ; i++ {
		if limiter == nil {
			if i >= concurrentNum {
				break
			}
		} else {
			if !limiter.Wait() {
				if atomic.AddInt32(&g.unlimited, -1) == 0 {
					close(g.unlimitedChan)
				}
				break
			}
			if g.IsDone() {
				atomic.StoreInt32(&g.unlimited, 0)
				close(g.unlimitedChan)
				break
			}
		}
		wg.Add(1)
		g.waitGroup.Add(1)
		go func() {
			defer func() {
				if err := recover(); err != nil {
					if atomic.CompareAndSwapInt32(&(g.closed), 0, 1) {
						g.Cancel()
						g.err = fmt.Errorf("%s: %v", name, err)
					} else {
						panic("内部执行异常")
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
		defer func() {
			if err := recover(); err != nil {
				if atomic.CompareAndSwapInt32(&(g.closed), 0, 1) {
					g.Cancel()
					g.err = fmt.Errorf("%s: %v", name, err)
				} else {
					panic("内部执行异常")
				}
			}
		}()
		for i := 0; i < len(run.clearup); i++ {
			run.clearup[i]()
		}
	}()
}

func (g *goroutiner) Wait() {
	if _, ok := <-g.unlimitedChan; !ok {
		g.waitGroup.Wait()
	}
}

func (g *goroutiner) Cancel() {
	g.cancel()
}

// 注册一个点位，用来退出协程，详情看demo
func (g *goroutiner) IsDone() bool {
	if atomic.LoadInt32(&g.blocked) == 0 {
		select {
		case <-g.ctx.Done():
			return true
		default:
			return false
		}
	}
	select {
	case <-g.blockchan:
		return false
	case <-g.ctx.Done():
		return true
	}

}

func (g *goroutiner) Error() error {
	if g.err != nil {
		return g.err
	} else {
		return g.ctx.Err()
	}
}

func (g *goroutiner) Block(fn func()) error {
	if fn == nil {
		return errors.New("阻塞必须要有一个阻塞方法")
	}
	if atomic.LoadInt32(&g.closed) == 1 {
		return errors.New("当前xgo已经关闭")
	}
	if !atomic.CompareAndSwapInt32(&g.blocked, 0, 1) {
		return errors.New("阻塞信息挂载失败，当前有goroutine正阻塞在历史阻塞方法中")
	}
	select {
	case <-g.ctx.Done():
		return g.Error()
	default:
		g.blockchan = make(chan struct{})
		go func() {
			defer func() {
				if err := recover(); err != nil {
					if atomic.CompareAndSwapInt32(&(g.closed), 0, 1) {
						g.Cancel()
						g.err = fmt.Errorf("%v", err)
					} else {
						panic("内部执行异常")
					}
				}
				if !atomic.CompareAndSwapInt32(&g.blocked, 1, 0) {
					panic("内部执行异常")
				}
				close(g.blockchan)
			}()
			fn()
		}()
		return nil
	}
}
