package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

// MetricsMiddleware records Prometheus metrics for each HTTP request.
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Process request
		c.Next()

		duration := time.Since(startTime)
		statusCode := c.Writer.Status()
		path := c.FullPath() // Use the route path for consistent labeling
		if path == "" {
			path = "not_found"
		}

		// Record metrics
		httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration.Seconds())
		httpRequestsTotal.WithLabelValues(c.Request.Method, path, strconv.Itoa(statusCode)).Inc()
	}
}
