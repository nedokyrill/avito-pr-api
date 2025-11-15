package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// PRLifecycleDurationHours отслеживает время жизни PR от создания до мерджа в часах
// Бакеты: 1h, 6h, 12h, 24h (1d), 48h (2d), 72h (3d), 168h (1 неделя), +Inf
var PRLifecycleDurationHours = prometheus.NewHistogram(prometheus.HistogramOpts{
	Name:    "pr_lifecycle_duration_hours",
	Help:    "Time from PR creation to merge in hours",
	Buckets: []float64{1, 6, 12, 24, 48, 72, 168},
})
