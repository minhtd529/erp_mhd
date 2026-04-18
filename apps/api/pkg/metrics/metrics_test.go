package metrics_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mdh/erp-audit/api/pkg/metrics"
)

func init() { gin.SetMode(gin.TestMode) }

func newRouter() *gin.Engine {
	r := gin.New()
	r.Use(metrics.Middleware())
	r.GET("/api/v1/clients", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	r.GET("/api/v1/clients/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"id": c.Param("id")})
	})
	r.GET("/metrics", metrics.Handler())
	return r
}

func TestMiddleware_RecordsRequest(t *testing.T) {
	r := newRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/clients", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
}

func TestMiddleware_MetricsEndpointServes(t *testing.T) {
	r := newRouter()
	// generate some traffic first
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/clients", nil)
		r.ServeHTTP(w, req)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/metrics", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("metrics endpoint: want 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "erp_audit_http_requests_total") {
		t.Error("metrics output missing erp_audit_http_requests_total")
	}
	if !strings.Contains(body, "erp_audit_http_request_duration_seconds") {
		t.Error("metrics output missing erp_audit_http_request_duration_seconds")
	}
}

func TestMiddleware_UsesRoutePattern(t *testing.T) {
	r := newRouter()
	// Hit parameterized route
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/clients/abc-123", nil)
	r.ServeHTTP(w, req)

	wm := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/metrics", nil)
	r.ServeHTTP(wm, req2)

	body := wm.Body.String()
	// Should use pattern /api/v1/clients/:id, not the literal /api/v1/clients/abc-123
	if strings.Contains(body, "abc-123") {
		t.Error("metrics must use route pattern, not literal path param value")
	}
	if !strings.Contains(body, `/api/v1/clients/:id`) {
		t.Error("metrics must contain route pattern /api/v1/clients/:id")
	}
}

func TestBusinessRecorders_NoPanic(t *testing.T) {
	// Verify business metric recorders don't panic
	metrics.RecordAuditLog()
	metrics.SetActiveEngagements(42)
	metrics.RecordInvoiceIssued()
	metrics.RecordCommissionAccrued()
	metrics.RecordAuthFailure("invalid_password")
	metrics.RecordRateLimitHit("user")
	metrics.RecordRateLimitHit("ip")
}
