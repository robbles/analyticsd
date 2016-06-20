package main

import (
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/expvar"
)

var quantiles = []int{50, 90, 99}
var sigfigs = 3

// Metrics represents a container for all application metrics.
type Metrics struct {
	RequestCount metrics.Counter
	ResponseTime metrics.TimeHistogram
}

// NewMetricsExpvar initializes and returns a Metrics exposed over the expvar system.
func NewMetricsExpvar() Metrics {
	return Metrics{
		RequestCount: expvar.NewCounter("request_count"),
		ResponseTime: metrics.NewTimeHistogram(
			time.Microsecond,
			expvar.NewHistogram("response_time", 0, int64(time.Second), sigfigs, quantiles...),
		),
	}
}
