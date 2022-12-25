# New
傻瓜式控制并发，防止协程泄露

## xgo.Run
启动协程，可以设置并发数，设置协程的名字，也可以设置结束时执行的函数

## xgo.Wait
等待我们的xgo启动的协程全部退出，与sync.WaitGroup一直

## xgo.Cancel
手动停止所有xgo启动的协程

## xgo.IsDone
需要在我们的实际任务中加这个函数来触发退出，让我们少写很多select ctx这些的监听退出

## xgo.Error
返回我们xgo启动协程的错误，如果是ctx超时，也会返回错误，手动取消也会返回错误
