## Gofherd

Gofherd (`gof-herd`), is a small framework for running user defined tasks with bounded parallelism. It's simple interface gives you a channel to put tasks into, allows you to define a function which has "processing logic" and gives you an output chan to read results from.


Gofherd provides:
- Bounded parallelism
  - You can configure number of gophers to run the tasks.
- Monitoring
  - Current state is exposed as Prometheus compatible metrics.
- Persistence
  - State is saved so it can be continued after pause/crash.
- HTTP APIs
  - For monitoring, to dynamically change the parallelism etc.

### Example

```go
package main

func main() {
	processinglogic := func(w Work) Result {
		fmt.Printf("got work to process: %s\n", w.ID)
		return Success
	}

	gf := New(processinglogic)
	gf.SetGopherd(10)
	inputChan := gf.InputChan()

	loadFunc := func() {
		for i := 0; i < 1000; i++ {
			w := Work{ID: fmt.Sprintf("%d", i)}
			fmt.Printf("got work to input: %s\n", w.ID)
			inputChan <- w
		}
		close(inputChan)
	}

	go loadFunc()

	gf.Start()

	outputChan := gf.OutputChan()
	for i := 0; i < 1000; i++ {
		w := <-outputChan
		fmt.Printf("got work output result: %s\n", w.Status())
	}
}
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
On calling `gf.GetOutputHose()`, a receive only channel `chan-> Work` is returned. It can be used to read the results for successfully processed work units.
