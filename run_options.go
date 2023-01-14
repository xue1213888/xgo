package xgo

// goroutine 运行配置
type runConfig struct {
	goroutineName string   // 这个是goroutine名
	concurrentNum int      // 这个是当前这个run要启动多少个goroutine
	clearup       []func() // 这个是当运行结束时需要执行的清理函数，意思整个run结束的defer
	limiter       Limiter  // 如果里面传了 limiter 我们就需要按照 limiter 限流措施无限开启 goroutine
}

type runOptions func(r *runConfig)

func execRunOptions(o ...runOptions) *runConfig {
	r := new(runConfig)
	for _, v := range o {
		v(r)
	}
	if r.goroutineName == "" {
		r.goroutineName = "xgo"
	}
	if r.concurrentNum == 0 {
		r.concurrentNum = 1
	}
	return r
}

func RunWithName(name string) runOptions {
	return func(g *runConfig) {
		g.goroutineName = name
	}
}

func RunWithNum(num int) runOptions {
	return func(g *runConfig) {
		g.concurrentNum = num
	}
}

// 限流
func RunWithLimiter(limit Limiter) runOptions {
	return func(g *runConfig) {
		g.limiter = limit
	}
}

func RunWithClearup(f ...func()) runOptions {
	return func(g *runConfig) {
		g.clearup = append(g.clearup, f...)
	}
}
