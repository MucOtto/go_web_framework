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
	once sync.Once
}

func NewPool(cap int, expire int) (p *Pool, err error) {
	if cap < 0 {
		return nil, ErrorInValidCap
	}

	if expire < 0 {
		return nil, ErrorInValidExpire
	}

	p.cap = int32(cap)
	p.expire = time.Duration(expire) * time.Second

	return p, nil
}

func (p *Pool) Submit(task func()) error {
	w := p.GetWorker()
	w.task <- task
	return nil
}

func (p *Pool) GetWorker() *Worker {
	return nil
}
