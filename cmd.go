package gofherd

type Gofherd struct {
	input           chan Work
	output          chan Work
	processingLogic func(Work) Status
	gofherd         int
	retry           int
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
		gf.output <- work
	}
}

func (gf *Gofherd) Start() {
	for i := 0; i < gf.gofherd; i++ {
		go gf.initGopher()
	}
}
