package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ratelimiter_requests_total",
			Help: "Total number of rate limit check requests.",
		},
		[]string{"client_id", "algorithm", "result"},
	)

	CheckDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ratelimiter_check_duration_seconds",
			Help:    "Duration of rate limit check in seconds.",
			Buckets: []float64{0.001, 0.002, 0.005, 0.010, 0.025, 0.050},
		},
		[]string{"algorithm"},
	)

	CacheAgeSeconds = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ratelimiter_cache_age_seconds",
			Help: "Seconds since last successful cache refresh from Postgres.",
		},
	)
)

func Register() {
	prometheus.MustRegister(RequestsTotal, CheckDuration, CacheAgeSeconds)
}
