package metrics

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"strconv"
	"time"
)

var (
	// HTTP metrics
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Worker pool metrics
	workerPoolActiveJobs = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "worker_pool_active_jobs",
			Help: "Number of currently active jobs in worker pool",
		},
	)

	workerPoolQueuedJobs = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "worker_pool_queued_jobs",
			Help: "Number of queued jobs waiting to be processed",
		},
	)

	// Database metrics
	dbConnectionsActive = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_active",
			Help: "Number of active database connections",
		},
	)

	dbConnectionsIdle = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_idle",
			Help: "Number of idle database connections",
		},
	)
)

// Initialize registers all metrics with Prometheus
func Initialize() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(workerPoolActiveJobs)
	prometheus.MustRegister(workerPoolQueuedJobs)
	prometheus.MustRegister(dbConnectionsActive)
	prometheus.MustRegister(dbConnectionsIdle)
}

// PrometheusHandler returns the Prometheus metrics handler
func PrometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// MetricsMiddleware records HTTP metrics
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip metrics endpoint itself
		if c.Request.URL.Path == "/api/v1/metrics" {
			c.Next()
			return
		}

		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Process request
		c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		httpRequestsTotal.WithLabelValues(method, path, status).Inc()
		httpRequestDuration.WithLabelValues(method, path).Observe(duration)
	}
}

// UpdateWorkerPoolMetrics updates worker pool metrics
func UpdateWorkerPoolMetrics(active, queued int) {
	workerPoolActiveJobs.Set(float64(active))
	workerPoolQueuedJobs.Set(float64(queued))
}

// UpdateDBMetrics updates database connection metrics
func UpdateDBMetrics(active, idle int) {
	dbConnectionsActive.Set(float64(active))
	dbConnectionsIdle.Set(float64(idle))
}
