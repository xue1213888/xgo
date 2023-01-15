package xgo

type Limiter interface {
	Next() bool // 当 next 返回true时，我们继续创建 goroutine ，当返回 false 时，我们取消创建
}
