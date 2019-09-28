package gofherd

type Status int

const (
	Unprocessed Status = iota
	Success
	Retry
	Failure
)

var statusStrings = map[Status]string{
	Unprocessed: "unprocessed",
	Success:     "success",
	Retry:       "retry",
	Failure:     "failure",
}

func (r Status) String() string {
	if val, ok := statusStrings[r]; ok {
		return val
	}
	return "unknown"
}

type Work struct {
	ID     string
	Body   interface{}
	retry  int
	result interface{}
	status Status
}

func (w *Work) SetResult(result interface{}) {
	w.result = result
}

func (w *Work) setStatus(status Status) {
	w.status = status
}

func (w *Work) Status() Status {
	return w.status
}
