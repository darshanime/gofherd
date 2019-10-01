package gofherd

import (
	"sync"
	"sync/atomic"
)

type atomicBool struct {
	val *int64
}

func (a *atomicBool) setTrue() {
	atomic.StoreInt64(a.val, 1)
}

func (a *atomicBool) setFalse() {
	atomic.StoreInt64(a.val, 0)
}

func (a *atomicBool) value() bool {
	return atomic.LoadInt64(a.val) == 1
}

type queue struct {
	hose  chan Work
	mu    sync.Mutex
	num   uint64
	close atomicBool
}

func newQueue() queue {
	return queue{hose: make(chan Work), close: atomicBool{val: new(int64)}}
}

func (q *queue) increment() {
	atomic.AddUint64(&(q.num), 1)
}

func (q *queue) setClosedTrue() {
	q.close.setTrue()
}

func (q *queue) closed() bool {
	return q.close.value()
}

func (q *queue) count() uint64 {
	return q.num
}

func (q *queue) lock() {
	q.mu.Lock()
}

func (q *queue) unlock() {
	q.mu.Unlock()
}
