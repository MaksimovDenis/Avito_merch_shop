package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	httpRequestTotal             *prometheus.CounterVec
	httpRequestDurationHistogram *prometheus.HistogramVec
	httpRequestsDurationSummary  *prometheus.SummaryVec
}

func New() *Metrics {
	httpRequestTotal := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_total",
			Help: "The total amount of HTTP requests by path and code",
		},
		[]string{"path", "code"},
	)

	httpRequestDurationHistogram := promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "Histogram of requests duration in seconds",
			Buckets: []float64{
				0.1,  // 100ms
				0.2,  // 200ms
				0.25, // 250ms
				0.5,  // 500ms
				1,    // 1s
			},
		},
		[]string{"path"},
	)

	httpRequestsDurationSummary := promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_requests_duration_seconds_summary",
			Objectives: map[float64]float64{
				0.99: 0.001, // 0.99 +- 0.001
				0.95: 0.01,  // 0.95 +- 0.01
				0.5:  0.05,  // 0.5 +- 0.05
			},
		},
		[]string{"path"},
	)

	return &Metrics{
		httpRequestTotal:             httpRequestTotal,
		httpRequestDurationHistogram: httpRequestDurationHistogram,
		httpRequestsDurationSummary:  httpRequestsDurationSummary,
	}
}

func (hdl *Metrics) HTTPMetrics() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		ctx.Next()

		duration := time.Since(start).Seconds()
		path := ctx.Request.URL.Path
		code := strconv.Itoa(ctx.Writer.Status())

		hdl.httpRequestTotal.WithLabelValues(path, code).Inc()
		hdl.httpRequestDurationHistogram.WithLabelValues(path).Observe(duration)
	}
}
