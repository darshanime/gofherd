package gofherd

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	successMetric = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gofherd_success_total",
		Help: "The total number of success events",
	})
	failureMetric = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gofherd_failure_total",
		Help: "The total number of failure events",
	})
	retryMetric = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gofherd_retry_total",
		Help: "The total number of retry events",
	})
)

func incrementSuccessMetric() {
	successMetric.Inc()
}

func incrementRetryMetric() {
	retryMetric.Inc()
}

func incrementFailureMetric() {
	failureMetric.Inc()
}
