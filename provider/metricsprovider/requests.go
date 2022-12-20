package metricsprovider

import (
	"github.com/prometheus/client_golang/prometheus"
)

type HttpMetrics struct {
	TotalRequests  *prometheus.CounterVec
	ResponseStatus *prometheus.CounterVec
	Duration       *prometheus.HistogramVec
}

// TODO use sync.once
func GetHttpMetrics(reg prometheus.Registerer) *HttpMetrics {
	met := &HttpMetrics{
		TotalRequests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "requestmaker",
				Name:      "http_requests_total",
				Help:      "Number of get requests.",
			},
			[]string{},
		),
		ResponseStatus: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "requestmaker",
				Name:      "http_response_status",
				Help:      "Status of HTTP response",
			},
			[]string{"code"},
		),
		Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "requestmaker",
			Name:      "http_response",
			Help:      "Duration of the request.",
			Buckets:   []float64{0.1, 0.15, 0.2, 0.25, 0.3},
		}, []string{"code", "method", "path"}),
	}
	reg.MustRegister(met.TotalRequests, met.ResponseStatus, met.Duration)
	return met
}
