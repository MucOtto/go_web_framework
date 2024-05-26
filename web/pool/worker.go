package pool

import (
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
	for f := range w.task {
		if f == nil {
			continue
		}
		f()
		w.pool.PutWorker(w)
		atomic.AddInt32(&w.pool.running, -1)
	}
}
