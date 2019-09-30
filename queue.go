package gofherd

import (
	"sync"
	"sync/atomic"
)

type queue struct {
	hose  chan Work
	mu    sync.Mutex
	num   uint64
	close bool
}

func (q *queue) increment() {
	atomic.AddUint64(&(q.num), 1)
}

func (q *queue) setClosedTrue() {
	q.close = true
}

func (q *queue) closed() bool {
	return q.close
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
