package xgo

import (
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type Limiter interface {
	Wait() bool
}

type slidingWindowComponents struct {
	sync.Mutex
	dur int64
	cap int64

	buf   []int64
	sendx int64
	recvx int64
	hold  int64

	waitChan interface{}
}

func NewSlidingWindowsLimiter(dur time.Duration, cnt int64, chanInterface ...interface{}) Limiter {
	var c interface{}
	if len(chanInterface) != 0 {
		c = chanInterface[0]
	}
	return &slidingWindowComponents{
		dur:      dur.Milliseconds(),
		cap:      cnt,
		buf:      make([]int64, cnt),
		waitChan: c,
	}
}

// Wait 会阻塞启动，如果返回 false 就不要继续启动了
func (s *slidingWindowComponents) Wait() bool {
SPIN:
	if !InterfaceChanClosedAndRecvOver(s.waitChan) {
		return false
	}
	currTime := time.Now().UnixMilli()
	if hold := atomic.LoadInt64(&s.hold); hold < s.cap {
		s.Lock()
		if s.hold != hold {
			goto SPIN
		}
		s.hold++
		s.sendx++
		s.sendx %= s.cap
		s.buf[s.sendx] = currTime
		s.Unlock()
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

type hchan struct {
	qcount   uint // total data in the queue
	dataqsiz uint // size of the circular queue
	buf      uint // points to an array of dataqsiz elements
	elemsize uint16
	closed   uint32 //28
	elemtype uint   // 32
	sendx    uint   // 40
	recvx    uint   // 48
	recvq    waitq  // 56
	sendq    waitq
}
type waitq struct {
	first *uint
	last  *uint
}

func InterfaceChanClosedAndRecvOver(i interface{}) bool {
	ch := *(*hchan)(unsafe.Pointer(*(*uintptr)((unsafe.Pointer((uintptr)(unsafe.Pointer(&i)) + 8)))))
	// fmt.Printf("%+v\n", ch)
	if ch.closed == 1 && ch.qcount == 0 {
		// fmt.Println("ch.closed", ch.closed)
		return false
	}
	return true
}
