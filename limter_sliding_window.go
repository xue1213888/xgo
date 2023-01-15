package xgo

import (
	"sync"
	"sync/atomic"
	"time"
)

type slidingWindowComponents struct {
	sync.Mutex
	dur int64
	cap int64

	buf   []int64
	sendx int64
	recvx int64
	hold  int64

	closed limiterListenClosedEventer
}

func NewSlidingWindowsLimiter(dur time.Duration, cnt int64, closed limiterListenClosedEventer) Limiter {
	return &slidingWindowComponents{
		dur:    dur.Milliseconds(),
		cap:    cnt,
		buf:    make([]int64, cnt),
		closed: closed,
	}
}

// Wait 会阻塞启动，如果返回 false 就不要继续启动了
func (s *slidingWindowComponents) Next() bool {
SPIN:
	if s.closed.Closed() {
		return false
	}
	currTime := time.Now().UnixMilli()
	if hold := atomic.LoadInt64(&s.hold); hold < s.cap {
		s.Lock()
		if s.hold != hold {
			s.Unlock()
			goto SPIN
		}
		s.hold++
		s.sendx++
		s.sendx %= s.cap
		s.buf[s.sendx] = currTime
		s.Unlock()
		s.closed.Add()
		return true
	}
	s.Lock()
	n := 0
	for {
		s.recvx %= s.cap
		shoudOut := currTime - s.dur
		if s.buf[s.recvx] <= shoudOut {
			s.hold--
			s.recvx++
			s.recvx %= s.cap
			n++
			if s.hold == 0 {
				break
			}
		} else {
			break
		}
	}
	s.Unlock()
	if n > 0 {
		goto SPIN
	}
	sleep := currTime - s.buf[s.recvx]
	time.Sleep(time.Duration(sleep))
	goto SPIN
}
