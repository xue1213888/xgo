package xgo

import (
	"sync/atomic"
)

type limiterCloseWaitStartNum struct {
	cnt int64
	now int64
}

func NewLimitUntilStartNum(cnt int64) limiterListenClosedEventer {
	return &limiterCloseWaitStartNum{
		cnt: cnt,
		now: 0,
	}
}

func (l *limiterCloseWaitStartNum) Closed() bool {
	return atomic.LoadInt64(&l.now) >= l.cnt
}

func (l *limiterCloseWaitStartNum) Add() {
	atomic.AddInt64(&l.now, 1)
}
