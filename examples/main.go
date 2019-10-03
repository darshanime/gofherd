package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	gf "github.com/darshanime/gofherd"
)

func LoadWork(herd *gf.Gofherd) {
	sites, err := os.Open("sites.csv")
	if err != nil {
		panic("could not find sites.csv. Please run from inside examples dir")
	}
	sitesReader := csv.NewReader(sites)
	counter := 0
	for {
		site, err := sitesReader.Read()
		if err == io.EOF {
			break
		}
		w := gf.Work{ID: fmt.Sprintf("%d", counter), Body: site[0]}
		herd.SendWork(w)
		counter++
	}
	herd.CloseInputChan()
}

func ProcessWork(w *gf.Work) gf.Status {
	start := time.Now()
	http.Get(w.Body.(string))
	w.SetResult(time.Since(start))
	return gf.Success
}

func ReviewOutput(outputChan <-chan gf.Work) {
	for work := range outputChan {
		fmt.Printf("site: %s, latency: %s\n", work.Body.(string), work.Result())
	}
}

func main() {
	herd := gf.New(ProcessWork)
	herd.SetHerdSize(5)
	herd.SetMaxRetries(100)
	herd.SetAddr("127.0.0.1:5555")
	go LoadWork(herd)
	herd.Start()
	ReviewOutput(herd.OutputChan())
}
