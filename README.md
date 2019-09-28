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
  - To for monitoring, dynamically change the parallelism.

### Example

```go
package main

func main() {
    gf := gopherd.Init(
        Work: GetBusinessLogic(),
        Workers: 10
    )

    input := gf.GetInputHose()
    output := gf.GetOutputHose()

    go LoadInput(input chan)
    go GetOutputHose()
    gf.Start()
}

func GetBusinessLogic(input, output chan interface{}) func (interface{}) gf.Status {
    businessLogic := func (u interface{}) gf.Status {
        input = <-output
        fmt.Printf("moved one from output to input")
        return gf.Status
    }
    return businessLogic
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
