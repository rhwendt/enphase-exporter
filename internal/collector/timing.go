package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// API call duration metrics - these are automatically registered
var (
	APICallDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "enphase_api_call_duration_seconds",
			Help:    "Duration of API calls to the Enphase gateway",
			Buckets: []float64{0.1, 0.25, 0.5, 1, 2, 3, 4, 5, 7.5, 10, 15, 20},
		},
		[]string{"endpoint"},
	)
)
