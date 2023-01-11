package xgo

import (
	"context"
	"errors"
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
	Block(fn func()) error
}

type goroutiner struct {
	waitGroup *sync.WaitGroup    // 内部的等待组
	parentCtx context.Context    // 外层传来的ctx，可以是一个带截至时间的ctx
	ctx       context.Context    // 内部的ctx
	cancel    context.CancelFunc // 内部的取消函数
	err       error              // 判断优先级高到低，第一个触发的panic，ctx取消，父ctx取消
	closed    int32              // 如果其中一个goroutine或者阻塞函数触发了panic，我们都会关闭整个组
	blockchan chan struct{}      // 当前阻塞的信号通知
	blocked   int32              // 当前的阻塞状态
}

// New 初始化一个 goroutine 组
func New(ctx context.Context) Xgo {
	nctx, cancel := context.WithCancel(ctx)
	return &goroutiner{
		waitGroup: &sync.WaitGroup{},
		parentCtx: ctx,
		ctx:       nctx,
		cancel:    cancel,
	}
}

// runConfig 运行一个 goroutine 时可以传这些参数
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
		defer func() {
			if err := recover(); err != nil {
				if atomic.CompareAndSwapInt32(&(g.closed), 0, 1) {
					g.Cancel()
					g.err = fmt.Errorf("%s: %v", n, err)
				}
			}
		}()
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
	} else if g.parentCtx.Err() != nil {
		return g.parentCtx.Err()
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
					}
				}
				if !atomic.CompareAndSwapInt32(&g.blocked, 1, 0) {
					panic("阻塞解除失败，按预期到这里阻塞应该还在持续，但阻塞已经被摘除，内部错误")
				}
				close(g.blockchan)
			}()
			fn()
		}()
		return nil
	}
}
