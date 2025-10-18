package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
)

var (
    Requests = prometheus.NewCounterVec(prometheus.CounterOpts{
        Name: "aegis_requests_total",
        Help: "Total requests processed by gateway",
    }, []string{"agent","tool","action","decision"})

    Latency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
        Name: "aegis_request_latency_seconds",
        Help: "Request latency in seconds",
        Buckets: prometheus.DefBuckets,
    }, []string{"agent","tool","action"})
)

func init() {
    prometheus.MustRegister(Requests)
    prometheus.MustRegister(Latency)
}
