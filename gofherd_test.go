package gofherd

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func randomStatus() Status {
	return Status(rand.Intn(len(statusStrings)))
}

func getBasicGopherd(maxRetries, workUnits, gofherdSize int, status Status) *Gofherd {
	gf := New(func(w *Work) Status { return status })
	gf.SetHerdSize(gofherdSize)
	gf.SetMaxRetries(maxRetries)

	go func() {
		for i := 0; i < workUnits; i++ {
			gf.SendWork(Work{ID: fmt.Sprintf("%d", i)})
		}
		gf.CloseInputChan()
	}()
	return gf
}

func assertAllChannelsClosed(herd *Gofherd, t *testing.T) {
	result := make(chan struct{})
	go func() {
		select {
		case work, ok := <-herd.input.hose:
			if ok {
				t.Fatalf("expected input chan to be closed, it is not, got work with ID: %s", work.ID)
			}
		}

		select {
		case work, ok := <-herd.retry.hose:
			if ok {
				t.Fatalf("expected retry chan to be closed, it is not, got work with ID: %s", work.ID)
			}
		}

		select {
		case work, ok := <-herd.output.hose:
			if ok {
				t.Fatalf("expected output chan to be closed, it is not, got work with ID: %s", work.ID)
			}
		}
		result <- struct{}{}
	}()

	select {
	case <-result:
	case <-time.After(2 * time.Second):
		t.Fatalf("did not get chan closed confirmation in 2 seconds, looks like a deadlock")
	}
}

func TestGopherdSuccess(t *testing.T) {
	maxRetries := 10
	workUnits := 1
	gofherdSize := 1
	gf := getBasicGopherd(maxRetries, workUnits, gofherdSize, Success)
	gf.Start()
	for i := 0; i < workUnits; i++ {
		w := <-gf.output.hose
		if w.Status() != Success || w.RetryCount() != 0 {
			t.Fatalf("did not receive expected status in output, expected: %s, got: %s\n", Success, w.Status())
		}
	}
	assertAllChannelsClosed(gf, t)
}

func TestGopherdRetries(t *testing.T) {
	maxRetries := 1000
	workUnits := 1000
	gofherdSize := 1
	gf := getBasicGopherd(maxRetries, workUnits, gofherdSize, Retry)
	gf.Start()

	for i := 0; i < workUnits; i++ {
		w := <-gf.output.hose
		if w.Status() != Failure || w.RetryCount() != maxRetries {
			t.Fatalf("did not receive expected status in output, expected: %s, got: %s\n", Failure, w.Status())
		}
	}
	assertAllChannelsClosed(gf, t)
}

func TestGopherdFailure(t *testing.T) {
	maxRetries := 10
	workUnits := 1
	gofherdSize := 2
	gf := getBasicGopherd(maxRetries, workUnits, gofherdSize, Failure)
	gf.Start()

	for i := 0; i < workUnits; i++ {
		w := <-gf.output.hose
		if w.Status() != Failure || w.RetryCount() != 0 {
			t.Fatalf("did not receive expected status in output, expected: %s, got: %s\n", Failure, w.Status())
		}
	}
	assertAllChannelsClosed(gf, t)
}

func TestGopherdNew(t *testing.T) {
	gf := New(func(w *Work) Status { return Success })

	gf.SetHerdSize(10)
	if gf.herdSize != 10 {
		t.Fatal("could not set herd size using SetHerdSize()")
	}

	gf.SetMaxRetries(10)
	if gf.maxRetries != 10 {
		t.Fatal("could not set maxRetries using SetMaxRetries()")
	}

	gf.SetAddr("0.0.0.0:2345")
	if gf.addr != "0.0.0.0:2345" {
		t.Fatal("could not set addr using SetAddr()")
	}
}

func TestMetricIncrementOnPushToOutputChanWithSuccess(t *testing.T) {
	maxRetries := 0
	workUnits := 1
	gofherdSize := 1
	gf := getBasicGopherd(maxRetries, workUnits, gofherdSize, Success)

	oldVal := testutil.ToFloat64(successMetric)
	expectedNewVal := oldVal + 1.0
	gf.Start()

	for i := 0; i < workUnits; i++ {
		<-gf.output.hose
		if newVal := testutil.ToFloat64(successMetric); newVal != expectedNewVal {
			t.Fatalf("did not receive expected val in success metric, expected: %f, got: %f\n", expectedNewVal, newVal)
		}
	}
	assertAllChannelsClosed(gf, t)
}

func TestMetricIncrementOnPushToOutputChanWithFailure(t *testing.T) {
	maxRetries := 0
	workUnits := 1
	gofherdSize := 1
	gf := getBasicGopherd(maxRetries, workUnits, gofherdSize, Failure)

	oldVal := testutil.ToFloat64(failureMetric)
	expectedNewVal := oldVal + 1.0
	gf.Start()

	for i := 0; i < workUnits; i++ {
		<-gf.output.hose
		if newVal := testutil.ToFloat64(failureMetric); newVal != expectedNewVal {
			t.Fatalf("did not receive expected val in failure metric, expected: %f, got: %f\n", expectedNewVal, newVal)
		}
	}
	assertAllChannelsClosed(gf, t)
}

func TestMetricIncrementOnPushToRetryChan(t *testing.T) {
	maxRetries := 5
	workUnits := 1
	gofherdSize := 1
	gf := getBasicGopherd(maxRetries, workUnits, gofherdSize, Retry)

	oldVal := testutil.ToFloat64(retryMetric)
	expectedNewVal := oldVal + float64(maxRetries)
	gf.Start()

	for i := 0; i < workUnits; i++ {
		<-gf.output.hose
		if newVal := testutil.ToFloat64(retryMetric); newVal != expectedNewVal {
			t.Fatalf("did not receive expected val in retry metric, expected: %f, got: %f\n", expectedNewVal, newVal)
		}
	}
	assertAllChannelsClosed(gf, t)
}

func TestProcessingLogicMakesUpdatesToWork(t *testing.T) {
	maxRetries := 5
	workUnits := 1
	gofherdSize := 1
	gf := New(func(w *Work) Status {
		w.Body = w.Body.(int) + 10
		return Success
	})
	gf.SetHerdSize(gofherdSize)
	gf.SetMaxRetries(maxRetries)

	go func() {
		for i := 0; i < workUnits; i++ {
			gf.SendWork(Work{ID: fmt.Sprintf("%d", i), Body: 5})
		}
		gf.CloseInputChan()
	}()

	gf.Start()

	for i := 0; i < workUnits; i++ {
		w := <-gf.output.hose
		if w.Body.(int) != 15 || w.RetryCount() != 0 {
			t.Fatalf("did not receive expected body in output, expected: %d, got: %d\n", 15, w.Body.(int))
		}
	}
	assertAllChannelsClosed(gf, t)

}
