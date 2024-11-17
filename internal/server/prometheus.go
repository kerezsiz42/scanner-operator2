package server

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var httpResponseDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_response_duration_seconds",
		Help:    "Duration of HTTP responses in seconds",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"path", "method"},
)

func init() {
	metrics.Registry.MustRegister(httpResponseDuration)
}

func observeDuration(method, handlerPath string) func() {
	start := time.Now()
	return func() {
		duration := time.Since(start).Seconds()
		httpResponseDuration.WithLabelValues(handlerPath, method).Observe(duration)
	}
}
