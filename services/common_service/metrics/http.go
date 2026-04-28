package metrics

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type HTTPMetrics struct {
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	inFlight        prometheus.Gauge
}

var (
	registryMu sync.Mutex
	registries = make(map[string]*HTTPMetrics)
)

func getOrCreate(service string) *HTTPMetrics {
	registryMu.Lock()
	defer registryMu.Unlock()

	if m, ok := registries[service]; ok {
		return m
	}

	m := &HTTPMetrics{
		requestsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "pushpost",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests.",
		}, []string{"service", "method", "path", "status"}),
		requestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "pushpost",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP request latencies in seconds.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"service", "method", "path", "status"}),
		inFlight: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: "pushpost",
			Subsystem: "http",
			Name:      "in_flight_requests",
			Help:      "Number of in-flight HTTP requests.",
			ConstLabels: prometheus.Labels{
				"service": service,
			},
		}),
	}

	registries[service] = m

	return m
}

func Middleware(service string) func(http.Handler) http.Handler {
	metrics := getOrCreate(service)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			metrics.inFlight.Inc()
			defer metrics.inFlight.Dec()

			ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
			started := time.Now()

			next.ServeHTTP(ww, r)

			statusCode := ww.Status()
			if statusCode == 0 {
				statusCode = http.StatusOK
			}

			path := "unknown"
			if routeCtx := chi.RouteContext(r.Context()); routeCtx != nil {
				if routePattern := routeCtx.RoutePattern(); routePattern != "" {
					path = routePattern
				}
			}

			labels := []string{service, r.Method, path, strconv.Itoa(statusCode)}
			metrics.requestsTotal.WithLabelValues(labels...).Inc()
			metrics.requestDuration.WithLabelValues(labels...).Observe(time.Since(started).Seconds())
		})
	}
}

func Handler() http.Handler {
	return promhttp.Handler()
}
