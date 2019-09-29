package gofherd

import "testing"

func TestQueueMethods(t *testing.T) {
	s := queue{hose: make(chan Work)}
	s.Increment()
	if s.Count() != 1 {
		t.Fatalf("state Count did not work as expected. expected:%d, got:%d", 1, s.Count())
	}

	s.SetClosedTrue()
	if s.Closed() != true {
		t.Fatalf("state SetClosedTrue did not work as expected. expected:%t, got:%t", true, s.Closed())
	}
}
