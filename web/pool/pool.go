package pool

import (
	"errors"
	"sync"
	"time"
)

const DefaultExpire = 3

var (
	ErrorInValidCap    = errors.New("pool cap can not <= 0")
	ErrorInValidExpire = errors.New("pool expire can not <= 0")
	ErrorHasClosed     = errors.New("pool has bean released!!")
)

type sig struct {
}

type Pool struct {
	Workers []*Worker
	// 最大容量
	cap int32
	// 正在运行的容量
	running int32
	// 过期时间 空闲worker超过这个阈值之后进行回收
	expire time.Duration
	//release 释放资源  pool就不能使用了
	release chan sig
	//lock 去保护pool里面的相关资源的安全
	lock sync.Mutex
	//once 释放只能调用一次 不能多次调用
	once        sync.Once
	workerCache sync.Pool
}

func NewPool(cap int, expire int) (p *Pool, err error) {
	p = &Pool{}
	if cap < 0 {
		return nil, ErrorInValidCap
	}

	if expire < 0 {
		return nil, ErrorInValidExpire
	}

	p.cap = int32(cap)
	p.expire = time.Duration(expire) * time.Second

	p.workerCache.New = func() any {
		return &Worker{
			pool: p,
			task: make(chan func(), 8),
		}
	}

	return p, nil
}

func (p *Pool) Submit(task func()) error {
	if len(p.release) > 0 {
		return ErrorHasClosed
	}
	w := p.GetWorker()
	w.task <- task
	return nil
}

func (p *Pool) GetWorker() *Worker {
	n := len(p.Workers) - 1
	// 如果有空闲 直接取
	if n >= 0 {
		w := p.Workers[n]
		p.Workers[n] = nil
		p.Workers = p.Workers[:n]
		return w
	}
	// 如果没空闲新建一个
	// 满足运行数量小于容量
	if p.running < p.cap {
		get := p.workerCache.Get()
		var w *Worker
		if get == nil {
			w = &Worker{
				pool: p,
				task: make(chan func(), 8),
			}
		} else {
			w = get.(*Worker)
		}

		w.run()
		return w
	}
	// 运行数量大于容量 阻塞
	for {
		p.lock.Lock()
		idleWorkers := p.Workers
		n := len(idleWorkers) - 1
		if n < 0 {
			p.lock.Unlock()
			continue
		}
		w := idleWorkers[n]
		idleWorkers[n] = nil
		p.Workers = idleWorkers[:n]
		p.lock.Unlock()
		return w
	}
}

func (p *Pool) PutWorker(w *Worker) {
	w.lastTime = time.Now()
	p.lock.Lock()
	p.Workers = append(p.Workers, w)
	p.lock.Unlock()
}

func (p *Pool) Release() {
	p.once.Do(func() {
		//只执行一次
		p.lock.Lock()
		workers := p.Workers
		for i, w := range workers {
			w.task = nil
			w.pool = nil
			workers[i] = nil
		}
		p.Workers = nil
		p.lock.Unlock()
		p.release <- sig{}
	})
}

func (p *Pool) Restart() bool {
	if len(p.release) <= 0 {
		return true
	}
	_ = <-p.release
	go p.expireWorker()
	return true
}

func (p *Pool) expireWorker() {
	ticker := time.NewTicker(p.expire)
	n := len(p.Workers) - 1
	p.lock.Lock()
	for range ticker.C {
		if len(p.release) > 0 {
			break
		}
		if n > 0 {
			for i, worker := range p.Workers {
				if time.Now().Sub(worker.lastTime) < p.expire {
					// 先进先出 如果第一个都小于 后面的都小于
					break
				}
				n = i
				worker.task <- nil
				p.Workers[i] = nil
				p.Workers = p.Workers[i+1:]
			}
		}
	}
	p.lock.Unlock()
}
