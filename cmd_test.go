package gofherd

import (
	"fmt"
	"testing"
)

func TestGopherd(t *testing.T) {
	gf := New(func(w Work) Result {
		return Success
	})
	gf.SetGopherd(10)
	inputChan := gf.InputChan()

	go func() {
		for i := 0; i < 1000; i++ {
			inputChan <- Work{ID: fmt.Sprintf("%d", i)}
		}
		close(inputChan)
	}()

	gf.Start()

	outputChan := gf.OutputChan()
	for i := 0; i < 1000; i++ {
		w := <-outputChan
		if w.Status() != Success {
			t.Fatal("did not receive success in output")
		}
	}
}
