package gofherd

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Gofherd struct {
	input           queue
	output          queue
	retry           queue
	processingLogic func(Work) Status
	herdSize        int
	maxRetries      int
}

func New(processingLogic func(Work) Status) *Gofherd {
	return &Gofherd{
		processingLogic: processingLogic,
		input:           queue{hose: make(chan Work)},
		output:          queue{hose: make(chan Work)},
		retry:           queue{hose: make(chan Work)},
	}
}

func (gf *Gofherd) SendWork(w Work) {
	gf.input.Increment()
	gf.input.hose <- w
}

func (gf *Gofherd) OutputChan() <-chan Work {
	return gf.output.hose
}

func (gf *Gofherd) CloseInputChan() {
	gf.input.Lock()
	defer gf.input.Unlock()
	if !gf.input.Closed() {
		close(gf.input.hose)
		gf.input.SetClosedTrue()
		gf.MaintainRetry()
	}
}

func (gf *Gofherd) CloseOutputChan() {
	gf.output.Lock()
	defer gf.output.Unlock()
	if !gf.output.Closed() {
		close(gf.output.hose)
		gf.output.SetClosedTrue()
	}
}

func (gf *Gofherd) SetHerdSize(num int) {
	gf.herdSize = num
}

func (gf *Gofherd) SetMaxRetries(num int) {
	gf.maxRetries = num
}

func (gf *Gofherd) PushToOutputChan(work Work) {
	if work.Status() == Success {
		IncrementSuccessMetric()
	}
	if work.Status() == Failure {
		IncrementFailureMetric()
	}
	gf.output.hose <- work
	gf.output.Increment()
	gf.MaintainRetry()
	return
}

func (gf *Gofherd) MaintainRetry() {
	gf.retry.Lock()
	defer gf.retry.Unlock()
	if !gf.retry.Closed() && gf.input.Closed() && gf.input.Count() == gf.output.Count() {
		gf.closeRetryChan()
	}
}

func (gf *Gofherd) closeRetryChan() {
	close(gf.retry.hose)
	gf.retry.SetClosedTrue()
}

func (gf *Gofherd) PushToRetryChan(work Work) {
	IncrementRetryMetric()
	work.IncrementRetries()
	go func() { gf.retry.hose <- work }()
	return
}

func (gf *Gofherd) initGopher() {
	var work Work
	var ok bool
	for {
		select {
		case work, ok = <-gf.input.hose:
			if !ok {
				goto handleRetries
			}
			gf.handleInput(work)
		case work, ok = <-gf.retry.hose:
			if !ok {
				gf.CloseOutputChan()
				return
			}
			gf.handleInput(work)
		}
	}
handleRetries:
	for work := range gf.retry.hose {
		gf.handleInput(work)
	}
	gf.CloseOutputChan()
}

func (gf *Gofherd) handleInput(work Work) {
	status := gf.processingLogic(work)
	work.setStatus(status)
	if work.Status() == Success || work.Status() == Failure {
		gf.PushToOutputChan(work)
		return
	}

	if work.Status() == Retry && work.RetryCount() < gf.maxRetries {
		gf.PushToRetryChan(work)
		return
	}
	work.setStatus(Failure)
	gf.PushToOutputChan(work)
}

func (gf *Gofherd) Start() {
	go http.ListenAndServe(":2112", promhttp.Handler())

	for i := 0; i < gf.herdSize; i++ {
		go gf.initGopher()
	}
}
