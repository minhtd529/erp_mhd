// Package metrics provides Prometheus instrumentation for the ERP Audit API.
package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "erp_audit",
		Subsystem: "http",
		Name:      "requests_total",
		Help:      "Total number of HTTP requests by method, path, and status code.",
	}, []string{"method", "path", "status"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "erp_audit",
		Subsystem: "http",
		Name:      "request_duration_seconds",
		Help:      "HTTP request latency distribution.",
		Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
	}, []string{"method", "path"})

	httpRequestsInFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "erp_audit",
		Subsystem: "http",
		Name:      "requests_in_flight",
		Help:      "Number of HTTP requests currently being processed.",
	})

	// Business metrics

	auditLogsWritten = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "erp_audit",
		Subsystem: "audit",
		Name:      "logs_written_total",
		Help:      "Total number of audit log entries written.",
	})

	activeEngagements = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "erp_audit",
		Subsystem: "business",
		Name:      "active_engagements",
		Help:      "Current number of active engagements.",
	})

	invoicesIssued = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "erp_audit",
		Subsystem: "business",
		Name:      "invoices_issued_total",
		Help:      "Total number of invoices issued.",
	})

	commissionsAccrued = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "erp_audit",
		Subsystem: "business",
		Name:      "commissions_accrued_total",
		Help:      "Total number of commission accrual records created.",
	})

	authFailures = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "erp_audit",
		Subsystem: "auth",
		Name:      "failures_total",
		Help:      "Total number of authentication failures by reason.",
	}, []string{"reason"})

	rateLimitHits = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "erp_audit",
		Subsystem: "ratelimit",
		Name:      "hits_total",
		Help:      "Total number of rate-limit rejections by scope.",
	}, []string{"scope"})
)

// Middleware returns a Gin middleware that records per-request Prometheus metrics.
// It uses the matched route pattern (e.g. /api/v1/clients/:id) so high-cardinality
// path params don't create unbounded label sets.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		httpRequestsInFlight.Inc()
		defer httpRequestsInFlight.Dec()

		c.Next()

		path := c.FullPath()
		if path == "" {
			path = "unmatched"
		}
		status := strconv.Itoa(c.Writer.Status())
		elapsed := time.Since(start).Seconds()

		httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(elapsed)
	}
}

// Handler returns the Prometheus HTTP handler for /metrics.
func Handler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// --- Business event recorders (called from use cases / workers) ---

// RecordAuditLog increments the audit log counter.
func RecordAuditLog() { auditLogsWritten.Inc() }

// SetActiveEngagements updates the active engagements gauge.
func SetActiveEngagements(n float64) { activeEngagements.Set(n) }

// RecordInvoiceIssued increments the invoices-issued counter.
func RecordInvoiceIssued() { invoicesIssued.Inc() }

// RecordCommissionAccrued increments the commission-accrued counter.
func RecordCommissionAccrued() { commissionsAccrued.Inc() }

// RecordAuthFailure increments the auth-failure counter for a given reason.
// reason should be one of: "invalid_password", "account_locked", "totp_invalid", "token_expired"
func RecordAuthFailure(reason string) { authFailures.WithLabelValues(reason).Inc() }

// RecordRateLimitHit increments the rate-limit counter for "user" or "ip" scope.
func RecordRateLimitHit(scope string) { rateLimitHits.WithLabelValues(scope).Inc() }
