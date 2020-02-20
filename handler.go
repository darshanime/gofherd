package gofherd

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type herd struct {
	Num int64  `json:"num"`
	Msg string `json:"msg"`
}

func (gf *Gofherd) herdHandler(w http.ResponseWriter, r *http.Request) {
	var response []byte
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		foo := herd{Num: gf.herdSize, Msg: "success"}
		response, _ = json.Marshal(foo)
		fmt.Fprintf(w, string(response))
		return
	case http.MethodPatch:
		var herdPatch herd
		json.NewDecoder(r.Body).Decode(&herdPatch)
		status, msg := gf.updateHerdSize(herdPatch.Num)
		response, _ = json.Marshal(herd{Num: gf.herdSize, Msg: msg})
		if status == Retry {
			w.WriteHeader(http.StatusBadRequest)
		}
		if status == Success {
			w.WriteHeader(http.StatusOK)
		}
		fmt.Fprintf(w, string(response))
		return
	}

}
