package gofherd

import "fmt"

type Gofherd struct {
	input           chan Work
	output          chan Work
	processingLogic func(Work) Status
	gofherd         int
	maxRetries      int
}

func New(processingLogic func(Work) Status) *Gofherd {
	return &Gofherd{
		processingLogic: processingLogic,
		input:           make(chan Work),
		output:          make(chan Work),
	}
}

func (gf *Gofherd) InputChan() chan<- Work {
	return gf.input
}

func (gf *Gofherd) SetGopherd(num int) {
	gf.gofherd = num
}

func (gf *Gofherd) SetMaxRetries(num int) {
	gf.maxRetries = num
}

func (gf *Gofherd) OutputChan() <-chan Work {
	return gf.output
}

func (gf *Gofherd) initGopher() {
	for {
		work, ok := <-gf.input
		if !ok {
			return
		}
		status := gf.processingLogic(work)
		work.setStatus(status)
		if status == Retry && work.RetryCount() < gf.maxRetries {
			fmt.Printf("in retry, %d\n", work.RetryCount())
			work.IncrementRetries()
			gf.input <- work
			continue
		}
		gf.output <- work
	}
}

func (gf *Gofherd) Start() {
	for i := 0; i < gf.gofherd; i++ {
		go gf.initGopher()
	}
}
