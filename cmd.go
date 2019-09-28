package gofherd

type Result int

const (
	Unprocessed Result = iota
	Success
	Retry
	Failure
)

type Work struct {
	ID     string
	Body   interface{}
	retry  int
	result interface{}
	status Result
}

func (w *Work) SetResult(result interface{}) {
	w.result = result
}

func (w *Work) setStatus(result Result) {
	w.status = result
}

func (w *Work) Status() Result {
	return w.status
}

type Gofherd struct {
	input           chan Work
	output          chan Work
	processingLogic func(Work) Result
	gofherd         int
	retry           int
}

func New(processingLogic func(Work) Result) *Gofherd {
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
		result := gf.processingLogic(work)
		work.setStatus(result)
		gf.output <- work
	}
}

func (gf *Gofherd) Start() {
	for i := 0; i < gf.gofherd; i++ {
		go gf.initGopher()
	}
}
