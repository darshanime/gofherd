package gofherd

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Gofherd is the core struct, orchestrating all functionality.
// It offers all the public methods of Gofherd.
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

// New initializes a new Gofherd struct. It takes in the processing logic function
// with the signature `func(*gf.Work) gf.Status`
func New(processingLogic func(*Work) Status) *Gofherd {
	return &Gofherd{
		processingLogic: processingLogic,
		input:           queue{hose: make(chan Work)},
		output:          queue{hose: make(chan Work)},
		retry:           queue{hose: make(chan Work)},
		addr:            "127.0.0.1:2112",
		logger:          noOpLogger{},
	}
}

// SetLogger is used to setup logging. If not specified, gofherd emits no logs.
func (gf *Gofherd) SetLogger(l Logger) {
	gf.logger = l
}

// SendWork enques Work onto the input chan.
func (gf *Gofherd) SendWork(work Work) {
	gf.input.increment()
	gf.input.hose <- work
	gf.logger.Printf("Pushed to input, work: %s\n", work.ID)
}

// OutputChan returns the output chan, it will be closed when the processing is complete,
// enabling it to be read in a `for range` loop.
func (gf *Gofherd) OutputChan() <-chan Work {
	return gf.output.hose
}

// CloseInputChan is to closed the input chan. Closing the input chan when the tasks are completed
// will allow gofherd to shutdown gracefully.
func (gf *Gofherd) CloseInputChan() {
	gf.input.lock()
	defer gf.input.unlock()
	if !gf.input.closed() {
		close(gf.input.hose)
		gf.logger.Printf("Closed input chan\n")
		gf.input.setClosedTrue()
		gf.maintainRetry()
	}
}

func (gf *Gofherd) closeOutputChan() {
	gf.output.lock()
	defer gf.output.unlock()
	if !gf.output.closed() {
		close(gf.output.hose)
		gf.logger.Printf("Closed output chan\n")
		gf.output.setClosedTrue()
	}
}

// SetHerdSize sets the herd size. The passed number is the number of
// gofhers spawned up for processing.
func (gf *Gofherd) SetHerdSize(num int) {
	gf.herdSize = num
}

// SetAddr accepts the `addr` string where the started server will be spun up.
func (gf *Gofherd) SetAddr(addr string) {
	gf.addr = addr
}

// SetMaxRetries is the maximum number of times a Work unit will be tried before giving up.
func (gf *Gofherd) SetMaxRetries(num int) {
	gf.maxRetries = num
}

func (gf *Gofherd) pushToOutputChan(work Work) {
	if work.Status() == Success {
		incrementSuccessMetric()
	}
	if work.Status() == Failure {
		incrementFailureMetric()
	}
	gf.logger.Printf("Pusing to output, work: %s\n", work.ID)
	gf.output.hose <- work
	gf.output.increment()
	gf.maintainRetry()
	return
}

func (gf *Gofherd) maintainRetry() {
	gf.retry.lock()
	defer gf.retry.unlock()
	if !gf.retry.closed() && gf.input.closed() && gf.input.count() == gf.output.count() {
		gf.closeRetryChan()
	}
}

func (gf *Gofherd) closeRetryChan() {
	close(gf.retry.hose)
	gf.logger.Printf("Closed retry chan\n")
	gf.retry.setClosedTrue()
}

func (gf *Gofherd) pushToRetryChan(work Work) {
	incrementRetryMetric()
	work.incrementRetries()
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
				gf.closeOutputChan()
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
	gf.closeOutputChan()
}

func (gf *Gofherd) handleInput(work Work) {
	status := gf.processingLogic(&work)
	work.setStatus(status)
	if work.Status() == Success || work.Status() == Failure {
		gf.pushToOutputChan(work)
		return
	}

	if work.Status() == Retry && work.retryCount() < gf.maxRetries {
		gf.pushToRetryChan(work)
		return
	}
	work.setStatus(Failure)
	gf.pushToOutputChan(work)
}

// Start will start the processing and start the server. The function will return immediately.
func (gf *Gofherd) Start() {
	gf.logger.Printf("Starting server at %s\n", gf.addr)
	go http.ListenAndServe(gf.addr, promhttp.Handler())
	for i := 0; i < gf.herdSize; i++ {
		gf.logger.Printf("Starting gofher #%d\n", i)
		go gf.initGopher()
	}
}
