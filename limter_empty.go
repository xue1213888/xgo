package xgo

type emptyLimiter struct {
	ch limiterListenClosedEventer
}

func (e *emptyLimiter) Next() bool {
	return e.ch.Closed()
}

func NewEmptyLimterBlockChan(ch limiterListenClosedEventer) Limiter {
	return &emptyLimiter{
		ch: ch,
	}
}


