package xgo

// type goroutineCache struct {
// 	sync.Mutex
// 	ctx          context.Context
// 	goroutineCap int64
// 	running      int64
// 	freebuf      []chan func()
// 	workbuf      []chan func()
// }

// func newGoroutineCacher(ctx context.Context, cnt int64) *goroutineCache {
// 	return &goroutineCache{
// 		Mutex:        sync.Mutex{},
// 		ctx:          ctx,
// 		goroutineCap: cnt,
// 		running:      0,
// 		cachebuf:     make([]chan func(), cnt),
// 	}
// }

// func (g *goroutineCache) Start(fn func()) {

// }

// func (g *goroutineCache) newGoroutine() chan func() {
// 	ch := make(chan func())
// 	go func() {
// 	BREAK:
// 		for {
// 			select {
// 			case fn := <-ch:
// 				fn()
// 			case <-g.ctx.Done():
// 				break BREAK
// 			}
// 		}
// 	}()
// 	return ch
// }
