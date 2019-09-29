package gofherd

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Gofherd struct {
	input           chan Work
	output          chan Work
	retry           chan Work
	processingLogic func(Work) Status
	herdSize        int
	maxRetries      int
}

func New(processingLogic func(Work) Status) *Gofherd {
	return &Gofherd{
		processingLogic: processingLogic,
		input:           make(chan Work),
		output:          make(chan Work),
		retry:           make(chan Work),
	}
}

func (gf *Gofherd) InputChan() chan<- Work {
	return gf.input
}

func (gf *Gofherd) CloseInputChan() {
	close(gf.input)
}

func (gf *Gofherd) SetHerdSize(num int) {
	gf.herdSize = num
}

func (gf *Gofherd) SetMaxRetries(num int) {
	gf.maxRetries = num
}

func (gf *Gofherd) OutputChan() <-chan Work {
	return gf.output
}

func (gf *Gofherd) PushToOutputChan(work Work) {
	if work.Status() == Success {
		IncrementSuccessMetric()
	}
	if work.Status() == Failure {
		IncrementFailureMetric()
	}
	gf.output <- work
	return
}

func (gf *Gofherd) PushToRetryChan(work Work) {
	IncrementRetryMetric()
	work.IncrementRetries()
	go func() { gf.retry <- work }()
	return
}

func (gf *Gofherd) initGopher() {
	var work Work
	var ok bool
	for {
		select {
		case work, ok = <-gf.input:
			if !ok {
				goto handleRetries
			}
			gf.handleInput(work)
		case work, ok = <-gf.retry:
			if !ok {
				return
			}
			gf.handleInput(work)
		}
	}
handleRetries:
	for work := range gf.retry {
		gf.handleInput(work)
	}
	close(gf.retry)
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
	gf.output <- work
}

func (gf *Gofherd) Start() {
	go http.ListenAndServe(":2112", promhttp.Handler())

	for i := 0; i < gf.herdSize; i++ {
		go gf.initGopher()
	}
}
