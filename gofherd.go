package gofherd

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Gofherd struct {
	input           queue
	output          queue
	retry           queue
	processingLogic func(*Work) Status
	herdSize        int
	maxRetries      int
	addr            string
	logger          Logger
}

func New(processingLogic func(*Work) Status) *Gofherd {
	return &Gofherd{
		processingLogic: processingLogic,
		input:           queue{hose: make(chan Work)},
		output:          queue{hose: make(chan Work)},
		retry:           queue{hose: make(chan Work)},
		addr:            "127.0.0.1:2112",
		logger:          NoOpLogger{},
	}
}

func (gf *Gofherd) SetLogger(l Logger) {
	gf.logger = l
}

func (gf *Gofherd) SendWork(work Work) {
	gf.input.Increment()
	gf.input.hose <- work
	gf.logger.Printf("Pushed to input, work: %s\n", work.ID)
}

func (gf *Gofherd) OutputChan() <-chan Work {
	return gf.output.hose
}

func (gf *Gofherd) CloseInputChan() {
	gf.input.Lock()
	defer gf.input.Unlock()
	if !gf.input.Closed() {
		close(gf.input.hose)
		gf.logger.Printf("Closed input chan\n")
		gf.input.SetClosedTrue()
		gf.MaintainRetry()
	}
}

func (gf *Gofherd) CloseOutputChan() {
	gf.output.Lock()
	defer gf.output.Unlock()
	if !gf.output.Closed() {
		close(gf.output.hose)
		gf.logger.Printf("Closed output chan\n")
		gf.output.SetClosedTrue()
	}
}

func (gf *Gofherd) SetHerdSize(num int) {
	gf.herdSize = num
}

func (gf *Gofherd) SetAddr(addr string) {
	gf.addr = addr
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
	gf.logger.Printf("Pusing to output, work: %s\n", work.ID)
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
	gf.logger.Printf("Closed retry chan\n")
	gf.retry.SetClosedTrue()
}

func (gf *Gofherd) PushToRetryChan(work Work) {
	IncrementRetryMetric()
	work.IncrementRetries()
	go func() {
		gf.retry.hose <- work
		gf.logger.Printf("Pushed to retry, work: %s\n", work.ID)
	}()
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
			gf.logger.Printf("Received work from input: %s\n", work.ID)
			gf.handleInput(work)
		case work, ok = <-gf.retry.hose:
			if !ok {
				gf.CloseOutputChan()
				return
			}
			gf.logger.Printf("Received work from retry: %s\n", work.ID)
			gf.handleInput(work)
		}
	}
handleRetries:
	for work := range gf.retry.hose {
		gf.logger.Printf("Received work from retry: %s\n", work.ID)
		gf.handleInput(work)
	}
	gf.CloseOutputChan()
}

func (gf *Gofherd) handleInput(work Work) {
	status := gf.processingLogic(&work)
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
	gf.logger.Printf("Starting server at %s\n", gf.addr)
	go http.ListenAndServe(gf.addr, promhttp.Handler())
	for i := 0; i < gf.herdSize; i++ {
		gf.logger.Printf("Starting gofher #%d\n", i)
		go gf.initGopher()
	}
}
