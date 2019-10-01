package gofherd

import "testing"

func TestQueueMethods(t *testing.T) {
	s := newQueue()
	s.increment()
	if s.count() != 1 {
		t.Fatalf("state count did not work as expected. expected:%d, got:%d", 1, s.count())
	}

	s.setClosedTrue()
	if s.closed() != true {
		t.Fatalf("state setClosedTrue did not work as expected. expected:%t, got:%t", true, s.closed())
	}
}

func TestAtomicBool(t *testing.T) {
	close := atomicBool{val: new(int64)}

	close.setTrue()
	if close.value() != true {
		t.Fatalf("state setTrue did not work as expected. expected:%t, got:%t", true, close.value())
	}

	close.setFalse()
	if close.value() != false {
		t.Fatalf("state setFalse did not work as expected. expected:%t, got:%t", false, close.value())
	}

}
