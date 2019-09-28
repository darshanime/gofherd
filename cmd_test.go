package gofherd

import (
	"fmt"
	"testing"
)

func TestGopherdBasic(t *testing.T) {
	WorkUnits := 1
	GofherdSize := 2

	gf := New(func(w Work) Status {
		return Success
	})
	gf.SetGopherd(GofherdSize)
	inputChan := gf.InputChan()

	go func() {
		for i := 0; i < WorkUnits; i++ {
			inputChan <- Work{ID: fmt.Sprintf("%d", i)}
		}
		close(inputChan)
	}()

	gf.Start()

	outputChan := gf.OutputChan()
	for i := 0; i < WorkUnits; i++ {
		w := <-outputChan
		if w.Status() != Success {
			t.Fatal("did not receive success in output")
		}
	}
}

func TestGopherdRetries(t *testing.T) {
	MaxRetries := 10
	WorkUnits := 1
	GofherdSize := 2

	gf := New(func(w Work) Status {
		return Retry
	})
	gf.SetGopherd(GofherdSize)
	gf.SetMaxRetries(MaxRetries)
	inputChan := gf.InputChan()

	go func() {
		for i := 0; i < WorkUnits; i++ {
			inputChan <- Work{ID: fmt.Sprintf("%d", i)}
		}
	}()

	gf.Start()

	outputChan := gf.OutputChan()
	for i := 0; i < WorkUnits; i++ {
		w := <-outputChan
		if w.Status() != Failure && w.RetryCount() != MaxRetries {
			t.Fatal("did not receive expected retry behaviour")
		}
	}
}
