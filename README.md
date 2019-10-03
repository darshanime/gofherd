## Gofherd

Gofherd (`gof-herd`), is a small framework for running user defined functions with bounded parallelism. 
It's simple interface gives you a function which accepts tasks, allows you to define a function which has "processing logic" and gives you an output channel to read results from.

Gofherd provides:
- Bounded parallelism
  - You can configure number of gophers to run the tasks.
- Monitoring
  - Current state is exposed as Prometheus compatible metrics on `/metrics`
- Dynamic parallelism
  - Using `GET`/`PATCH` calls on `/herd`

### Example

```go
package main

import (
	"fmt"

	gf "github.com/darshanime/gofherd"
)

func LoadWork(herd *gf.Gofherd) {
	for i := 0; i < 10; i++ {
		w := gf.Work{ID: fmt.Sprintf("workdID:%d", i), Body: i}
		// blocking call, returns when work picked up by a goroutine
		herd.SendWork(w)
	}
	// done pushing all work, close input chan to indicate that
	herd.CloseInputChan()
}

// this is the signature of "processing logic" function
func ProcessWork(w *gf.Work) gf.Status {
	result := w.Body.(int) + 10
	// set the result after processing
	w.SetResult(result)
	return gf.Success
}

func ReviewOutput(outputChan <-chan gf.Work) {
	// output chan is closed when all results are received
	for work := range outputChan {
		fmt.Printf("%s, status: %s, result:%d\n", work.ID, work.Status(), work.Result())
	}
}

func main() {
	// passing "processing logic" when initializing work
	herd := gf.New(ProcessWork)
	// setting a herd size of 0
	herd.SetHerdSize(0)
	// disabling retries
	herd.SetMaxRetries(0)
	// bind on 127.0.0.1:5555
	herd.SetAddr("127.0.0.1:5555")
	go LoadWork(herd)
	herd.Start()
	ReviewOutput(herd.OutputChan())
}
```

Get current herd size: `curl -XGET localhost:5555/herd`
Now, we can increase the herd size using `curl -XPATCH 127.0.0.1:5555/herd -d '{"num": 10}'`

Output:

```
$ go run main.go
workdID:0, status: success, result:10
workdID:1, status: success, result:11
workdID:2, status: success, result:12
workdID:3, status: success, result:13
workdID:5, status: success, result:15
workdID:4, status: success, result:14
workdID:7, status: success, result:17
workdID:6, status: success, result:16
workdID:9, status: success, result:19
workdID:8, status: success, result:18
```


### Specification

When initializing `gofherd`, it takes:

- processing logic function with the signature `func ProcessWork(w *gf.Work) gf.Status`
The status can be one of:
```go
const (
	// Success represents a successful processing outcome. It won't be retried.
	Success Status = iota
	// Retry represents a failed processing outcome, but is retriable. It will be retried for MaxRetries.
	Retry
	// Failure represents a failed processing outcome and should not be retried again.
	Failure
)
```

Each unit of "work" is defined as the struct:

```go
type Work struct {
	ID     string
	retry  int64
	status Status
	Body   interface{}
	result interface{}
}
```

The `ID` field is used to track status of the work unit, retry count etc.
The `Body` field can be anything that makes sense for the usecase at hand. It is for the input problem, there is a `Result` field which has the output answer.

For sending work, `gf.SendWork` can be used. It is a blocking call and will return when a member of the herd is accepts the work.
On calling `gf.OutputChan()`, a receive only channel `<-chan Work` is returned which can be used to read the status for successfully processed work units. It will be closed by gofherd on completion.

#### Logging

Gofherd accepts 

```go
type Logger interface {
	Printf(format string, v ...interface{})
}
```

Example Usage:

```go
logger := log.New(os.Stdout, "gofherd:", log.Ldate|log.Ltime|log.Lshortfile)
herd.SetLogger(logger)
```

#### Contributing

Run the tests

```bash
go test -race -v ./...
```
