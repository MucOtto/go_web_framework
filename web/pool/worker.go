package pool

import (
	"github.com/MucOtto/web/log"
	"sync/atomic"
	"time"
)

type Worker struct {
	pool *Pool
	//task 任务队列
	task chan func()
	//lastTime 执行任务的最后的时间
	lastTime time.Time
}

func (w *Worker) run() {
	atomic.AddInt32(&w.pool.running, 1)
	go w.running()
}

func (w *Worker) running() {
	defer func() {
		if err := recover(); err != nil {
			atomic.AddInt32(&w.pool.running, -1)
			w.pool.workerCache.Put(w)
			if w.pool.PanicHandler != nil {
				w.pool.PanicHandler(err)
			} else {
				log.Default().Error(err)
			}
			w.pool.PutWorker(w)
		}
	}()
	for f := range w.task {
		if f == nil {
			atomic.AddInt32(&w.pool.running, -1)
			w.pool.workerCache.Put(w)
			break
		}
		f()
		w.pool.PutWorker(w)
	}
}
