package xgo

import (
	"unsafe"
)

type limiterCloseWaitChanClosed struct {
	ch interface{}
}

func NewLimitUntilChanClosed(ch interface{}) limiterListenClosedEventer {
	return &limiterCloseWaitChanClosed{
		ch: ch,
	}
}

func (l *limiterCloseWaitChanClosed) Closed() bool {
	return InterfaceChanClosedAndRecvOver(l.ch)
}
func (l *limiterCloseWaitChanClosed) Add() {}

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
	if ch.closed == 1 && ch.qcount == 0 {
		return true
	}
	return false
}
