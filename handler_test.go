package gofherd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func TestMetricsHandler(t *testing.T) {
	resp := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/metrics", nil)
	if err != nil {
		t.Fatalf("failed to create a request")
	}

	handler := promhttp.Handler()
	handler.ServeHTTP(resp, req)

	http.DefaultServeMux.ServeHTTP(resp, req)
	if status := resp.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestHerdGet(t *testing.T) {
	maxRetries := 10
	workUnits := 1
	gofherdSize := 15
	gf := getBasicGopherd(maxRetries, workUnits, gofherdSize, Success)

	resp := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/herd", nil)
	if err != nil {
		t.Fatalf("failed to create a request")
	}

	handler := http.HandlerFunc(gf.herdHandler)
	handler.ServeHTTP(resp, req)

	if status := resp.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	contentTypeHeaderValue := resp.Result().Header["Content-Type"][0]
	if contentTypeHeaderValue != "application/json" {
		t.Errorf("handler returned wrong value for content type: got %v want %v",
			contentTypeHeaderValue, "application/json")
	}

	expected := `{"num":15,"msg":"success"}`
	if resp.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			resp.Body.String(), expected)
	}
}

func TestHerdPatch(t *testing.T) {
	maxRetries := 10
	workUnits := 1
	gofherdSize := 1
	gf := getBasicGopherd(maxRetries, workUnits, gofherdSize, Success)
	gf.Start()

	resp := httptest.NewRecorder()
	reader := bytes.NewReader([]byte(`{"num": 10}`))
	req, err := http.NewRequest("PATCH", "/herd", reader)
	if err != nil {
		t.Fatalf("failed to create a request")
	}

	handler := http.HandlerFunc(gf.herdHandler)
	handler.ServeHTTP(resp, req)

	if status := resp.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	contentTypeHeaderValue := resp.Result().Header["Content-Type"][0]
	if contentTypeHeaderValue != "application/json" {
		t.Errorf("handler returned wrong value for content type: got %v want %v",
			contentTypeHeaderValue, "application/json")
	}

	expected := `{"num":10,"msg":"success"}`
	if resp.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			resp.Body.String(), expected)
	}
}
