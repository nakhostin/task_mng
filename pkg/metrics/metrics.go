package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// RequestsTotal counts the total number of HTTP requests
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// RequestLatencyHistogram measures the latency of HTTP requests
	RequestLatencyHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "request_latency_histogram",
			Help:    "Histogram of HTTP request latencies in seconds",
			Buckets: prometheus.DefBuckets, // Default buckets: 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10
		},
		[]string{"method", "endpoint"},
	)

	// TasksCount tracks the current number of tasks by status
	TasksCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "tasks_count",
			Help: "Current number of tasks by status",
		},
		[]string{"status"},
	)
)

// UpdateTasksCount updates the tasks count metric
func UpdateTasksCount(status string, count float64) {
	TasksCount.WithLabelValues(status).Set(count)
}
