package gofherd

import "testing"

func TestWork(t *testing.T) {
	w := Work{ID: "abc"}

	w.setStatus(Success)
	if w.Status() != Success {
		t.Fatal("could not set status as success for work")
	}

	w.SetResult("xyz")
	result := w.Result().(string)
	if result != "xyz" {
		t.Fatal("could not set result as xyz for work")
	}
}
