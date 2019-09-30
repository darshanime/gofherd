package gofherd

import "testing"

func TestQueueMethods(t *testing.T) {
	s := queue{hose: make(chan Work)}
	s.increment()
	if s.count() != 1 {
		t.Fatalf("state count did not work as expected. expected:%d, got:%d", 1, s.count())
	}

	s.setClosedTrue()
	if s.closed() != true {
		t.Fatalf("state setClosedTrue did not work as expected. expected:%t, got:%t", true, s.closed())
	}
}
