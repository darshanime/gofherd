## Gofherd

Gofherd (`gof-herd`), is a small framework for running user defined functions with bounded parallelism. It's simple interface gives you a channel to put tasks into, allows you to define a function which has "processing logic" and gives you an output channel to read results from.


Gofherd provides:
- Bounded parallelism
  - You can configure the herd size (number of gophers) to run the tasks.
- Monitoring
  - Prometheus compatible metrics exposed for tracking progress.
- Persistence
  - Progress is saved so execution can be continued after pause/crash.
- HTTP APIs
  - To dynamically change the parallelism etc.

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
		herd.SendWork(w)
	}
	herd.CloseInputChan()
}

func ProcessWork(w *gf.Work) gf.Status {
	workBody := w.Body.(int)
	w.Body = workBody + 10
	return gf.Success
}

func ReviewOutput(outputChan <-chan gf.Work) {
	for work := range outputChan {
		fmt.Printf("workdID: %s, result:%d\n", work.Status(), work.Body)
	}
}

func main() {
	herd := gf.New(ProcessWork)
	herd.SetHerdSize(10)
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

- processing logic function with the signature `func (in Work) gf.Status`
The status can be one of:
```go
type Status struct {
    Success int
    Failure int
    Retry   int
}
```

Each unit of "work" is defined as the struct:

```go
type Work struct {
    ID   string
    Body interface{}
}
```

The `ID` field is used to track status of the work unit, retry count etc.
The `Body` field can be anything that makes sense for the usecase at hand.

On calling `gf.GetInputHose()`, a send only channel `chan<- Work` is returned. It can be populated with the `Work` entries by the user.
On calling `gf.GetOutputHose()`, a receive only channel `chan-> Work` is returned. It can be used to read the status for successfully processed work units.
