package gofherd

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
		gf.output <- work
		return
	}

	if work.Status() == Retry && work.RetryCount() < gf.maxRetries {
		work.IncrementRetries()
		gf.retry <- work
		return
	}
	work.setStatus(Failure)
	gf.output <- work
}

func (gf *Gofherd) Start() {
	for i := 0; i < gf.herdSize; i++ {
		go gf.initGopher()
	}
}
