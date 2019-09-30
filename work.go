package gofherd

import "sync/atomic"

// Status represents the outcome of "processing" Work.
// It can be one of Success, Retry, Failure.
type Status int

const (
	// Success represents a successful processing outcome. It won't be retried.
	Success Status = iota
	// Retry represents a failed processing outcome, but is retriable. It will be retried for MaxRetries.
	Retry
	// Failure represents a failed processing outcome and should not be retried again.
	Failure
)

var statusStrings = map[Status]string{
	Success: "success",
	Retry:   "retry",
	Failure: "failure",
}

func (r Status) String() string {
	if val, ok := statusStrings[r]; ok {
		return val
	}
	return "unknown"
}

// Work is the struct representing the work unit in Gofherd.
// It has an ID field which is a string, `Body` and `Result` which are an
// interface to store the "problem" and "solution" respectively.
type Work struct {
	ID     string
	retry  int64
	status Status
	Body   interface{}
	result interface{}
}

func (w *Work) retryCount() int64 {
	return w.retry
}

func (w *Work) incrementRetries() {
	atomic.AddInt64(&(w.retry), 1)
}

// SetResult is used to set the result for the Work unit.
func (w *Work) SetResult(result interface{}) {
	w.result = result
}

func (w *Work) setStatus(status Status) {
	w.status = status
}

// Status is used to access the status of the Work unit.
func (w *Work) Status() Status {
	return w.status
}

// Result is used to access the Result of the Work unit.
func (w *Work) Result() interface{} {
	return w.result
}
