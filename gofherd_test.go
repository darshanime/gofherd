package gofherd

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func randomStatus() Status {
	return Status(rand.Intn(len(statusStrings)))
}

func getBasicGopherd(maxRetries, workUnits, gofherdSize int, status Status) *Gofherd {
	gf := New(func(w Work) Status { return status })
	gf.SetHerdSize(gofherdSize)
	gf.SetMaxRetries(maxRetries)
	inputChan := gf.InputChan()

	go func() {
		for i := 0; i < workUnits; i++ {
			inputChan <- Work{ID: fmt.Sprintf("%d", i)}
		}
		close(inputChan)
	}()
	return gf
}

func TestGopherdSuccess(t *testing.T) {
	maxRetries := 10
	workUnits := 1
	gofherdSize := 1
	gf := getBasicGopherd(maxRetries, workUnits, gofherdSize, Success)
	gf.Start()
	outputChan := gf.OutputChan()
	for i := 0; i < workUnits; i++ {
		w := <-outputChan
		if w.Status() != Success || w.RetryCount() != 0 {
			t.Fatalf("did not receive expected status in output, expected: %s, got: %s\n", Success, w.Status())
		}
	}
}

func TestGopherdRetries(t *testing.T) {
	maxRetries := 1000
	workUnits := 1000
	gofherdSize := 1
	gf := getBasicGopherd(maxRetries, workUnits, gofherdSize, Retry)
	gf.Start()

	outputChan := gf.OutputChan()
	for i := 0; i < workUnits; i++ {
		w := <-outputChan
		if w.Status() != Failure || w.RetryCount() != maxRetries {
			t.Fatalf("did not receive expected status in output, expected: %s, got: %s\n", Failure, w.Status())
		}
	}
}

func TestGopherdFailure(t *testing.T) {
	maxRetries := 10
	workUnits := 1
	gofherdSize := 2
	gf := getBasicGopherd(maxRetries, workUnits, gofherdSize, Failure)
	gf.Start()

	outputChan := gf.OutputChan()
	for i := 0; i < workUnits; i++ {
		w := <-outputChan
		if w.Status() != Failure || w.RetryCount() != 0 {
			t.Fatalf("did not receive expected status in output, expected: %s, got: %s\n", Failure, w.Status())
		}
	}
}

func TestGopherdNew(t *testing.T) {
	gf := New(func(Work) Status { return Success })

	inputChan := gf.InputChan()
	if inputChan != gf.input {
		t.Fatal("could not get input chan using InputChan()")
	}

	outputChan := gf.OutputChan()
	if outputChan != gf.output {
		t.Fatal("could not get output chan using OutputChan()")
	}

	gf.SetHerdSize(10)
	if gf.herdSize != 10 {
		t.Fatal("could not set herd size using SetHerdSize()")
	}

	gf.SetMaxRetries(10)
	if gf.maxRetries != 10 {
		t.Fatal("could not set maxRetries using SetMaxRetries()")
	}
}

func TestPushToOutputChanWithSuccess(t *testing.T) {
	maxRetries := 0
	workUnits := 1
	gofherdSize := 1
	gf := getBasicGopherd(maxRetries, workUnits, gofherdSize, Success)

	oldVal := testutil.ToFloat64(successMetric)
	expectedNewVal := oldVal + 1.0
	gf.Start()

	outputChan := gf.OutputChan()
	for i := 0; i < workUnits; i++ {
		<-outputChan
		if newVal := testutil.ToFloat64(successMetric); newVal != expectedNewVal {
			t.Fatalf("did not receive expected val in success metric, expected: %f, got: %f\n", expectedNewVal, newVal)
		}
	}
}

func TestPushToOutputChanWithFailure(t *testing.T) {
	maxRetries := 0
	workUnits := 1
	gofherdSize := 1
	gf := getBasicGopherd(maxRetries, workUnits, gofherdSize, Failure)

	oldVal := testutil.ToFloat64(failureMetric)
	expectedNewVal := oldVal + 1.0
	gf.Start()

	outputChan := gf.OutputChan()
	for i := 0; i < workUnits; i++ {
		<-outputChan
		if newVal := testutil.ToFloat64(failureMetric); newVal != expectedNewVal {
			t.Fatalf("did not receive expected val in failure metric, expected: %f, got: %f\n", expectedNewVal, newVal)
		}
	}
}

func TestPushToRetryChan(t *testing.T) {
	maxRetries := 5
	workUnits := 1
	gofherdSize := 1
	gf := getBasicGopherd(maxRetries, workUnits, gofherdSize, Retry)

	oldVal := testutil.ToFloat64(retryMetric)
	expectedNewVal := oldVal + float64(maxRetries)
	gf.Start()

	outputChan := gf.OutputChan()
	for i := 0; i < workUnits; i++ {
		<-outputChan
		if newVal := testutil.ToFloat64(retryMetric); newVal != expectedNewVal {
			t.Fatalf("did not receive expected val in retry metric, expected: %f, got: %f\n", expectedNewVal, newVal)
		}
	}
}
