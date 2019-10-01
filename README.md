## Gofherd

Gofherd (`gof-herd`), is a small framework for running user defined functions with bounded parallelism. 
It's simple interface gives you a function which accepts tasks, allows you to define a function which has "processing logic" and gives you an output channel to read results from.

### Example

```go
package main

import (
	"fmt"
	"log"
	"os"

	gf "github.com/darshanime/gofherd"
)

func LoadWork(herd *gf.Gofherd) {
	for i := 0; i < 10; i++ {
		w := gf.Work{ID: fmt.Sprintf("workdID:%d", i), Body: i}
		herd.SendWork(w)
	}
	herd.CloseInputChan()
}

func ProcessWork(w *gf.Work) gf.Status {
	result := w.Body.(int) + 10
	w.SetResult(result)
	return gf.Success
}

func ReviewOutput(outputChan <-chan gf.Work) {
	for work := range outputChan {
		fmt.Printf("workdID: %s, result:%d\n", work.Status(), work.Result())
	}
}

func main() {
	herd := gf.New(ProcessWork)
	herd.SetHerdSize(1)
	logger := log.New(os.Stdout, "gofherd:", log.Ldate|log.Ltime|log.Lshortfile)
	herd.SetLogger(logger)
	go LoadWork(herd)
	herd.Start()
	ReviewOutput(herd.OutputChan())
}
```

Output:

```
$ go run main.go
workdID: success, result:10
workdID: success, result:14
workdID: success, result:12
workdID: success, result:13
workdID: success, result:17
workdID: success, result:15
workdID: success, result:16
workdID: success, result:18
workdID: success, result:19
workdID: success, result:11
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
