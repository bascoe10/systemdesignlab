// Package obs centralises Prometheus metrics so every service exposes the
// exact same metric names. The Grafana dashboards (shared across all levels)
// are built against these names — treat them as a public API.
package obs

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "HTTP requests by route and status code.",
	}, []string{"route", "method", "code"})

	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request latency.",
		Buckets: []float64{.001, .0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5},
	}, []string{"route"})

	cacheRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cache_requests_total",
		Help: "Cache operations by outcome (hit, miss, bypass, error).",
	}, []string{"op", "result"})

	cacheNodeRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cache_node_requests_total",
		Help: "Cache operations routed to each node by the hash ring.",
	}, []string{"node"})

	CacheNodeMemUsed = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cache_node_memory_used_bytes",
		Help: "Memory used per cache node.",
	}, []string{"node"})

	CacheNodeMemMax = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cache_node_memory_max_bytes",
		Help: "Memory limit per cache node (0 = unlimited).",
	}, []string{"node"})

	CacheNodeItems = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cache_node_items",
		Help: "Items stored per cache node.",
	}, []string{"node"})

	CacheNodeEvictions = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cache_node_evictions_total",
		Help: "Cumulative evictions per cache node (from backend stats).",
	}, []string{"node"})

	DBQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "db_query_duration_seconds",
		Help:    "Database query latency.",
		Buckets: []float64{.0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1},
	}, []string{"query"})

	DBPoolInUse = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "db_pool_in_use",
		Help: "Open connections currently in use.",
	})

	DBPoolMax = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "db_pool_max",
		Help: "Maximum size of the connection pool.",
	})

	RateLimitRejected = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gateway_ratelimit_rejected_total",
		Help: "Requests rejected by the gateway rate limiter.",
	})
)

// CacheResult is the Cluster.OnResult hook.
func CacheResult(op, result string) { cacheRequests.WithLabelValues(op, result).Inc() }

// CacheNodeOp is the Cluster.OnNodeOp hook.
func CacheNodeOp(node string) { cacheNodeRequests.WithLabelValues(node).Inc() }

// MetricsHandler serves /metrics.
func MetricsHandler() http.Handler { return promhttp.Handler() }

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Instrument wraps a handler with request count + latency metrics under a
// stable route label (avoid raw paths — they explode cardinality).
func Instrument(route string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, req)
		httpDuration.WithLabelValues(route).Observe(time.Since(start).Seconds())
		httpRequests.WithLabelValues(route, req.Method, strconv.Itoa(rec.status)).Inc()
	})
}
