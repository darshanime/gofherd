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
