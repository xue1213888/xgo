package xgo

type limiterListenClosedEventer interface {
	Closed() bool
	Add()
}
