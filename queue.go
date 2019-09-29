package gofherd

import (
	"sync"
	"sync/atomic"
)

type queue struct {
	hose   chan Work
	lock   sync.Mutex
	count  uint64
	closed bool
}

func (q *queue) Increment() {
	atomic.AddUint64(&(q.count), 1)
}

func (q *queue) SetClosedTrue() {
	q.closed = true
}

func (q *queue) Closed() bool {
	return q.closed
}

func (q *queue) Count() uint64 {
	return q.count
}

func (q *queue) Lock() {
	q.lock.Lock()
}

func (q *queue) Unlock() {
	q.lock.Unlock()
}
