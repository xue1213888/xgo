package xgo

import (
	"context"
	"time"
)

type options func(g *goroutiner)

// WithContext 传入 ctx
func WithContext(ctx context.Context) options {
	nctx, cancel := context.WithCancel(ctx)
	return func(g *goroutiner) {
		g.ctx = nctx
		g.cancel = cancel
	}
}

// WithCancel 传入 cancel
func WithCancel(ctx context.Context, cancel context.CancelFunc) options {
	return func(g *goroutiner) {
		g.ctx = ctx
		g.cancel = cancel
	}
}

// WithTimeout 传入 timeout
func WithTimeout(t time.Duration) options {
	ctx, cancel := context.WithTimeout(context.Background(), t)
	return func(g *goroutiner) {
		g.ctx = ctx
		g.cancel = cancel
	}
}

func WithMaxGoroutineNum(cnt int64) options {
	return func(g *goroutiner) {
		g.goroutineCap = make(chan struct{}, cnt)
	}
}
