package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	pkgauth "github.com/mdh/erp-audit/api/pkg/auth"
	"github.com/mdh/erp-audit/api/pkg/audit"
	"github.com/mdh/erp-audit/api/pkg/middleware"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newRouter(roles []string, twoFAVerified bool) *gin.Engine {
	r := gin.New()
	// Simulate AuthMiddleware by pre-setting context values
	r.Use(func(c *gin.Context) {
		c.Set(pkgauth.CtxRoles, roles)
		c.Set(pkgauth.CtxTwoFAVerified, twoFAVerified)
		c.Next()
	})
	r.Use(middleware.Require2FA())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}

func TestRequire2FA_AuditStaff_NoVerificationNeeded(t *testing.T) {
	r := newRouter([]string{"AUDIT_STAFF"}, false)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("AUDIT_STAFF without 2FA should be allowed; got %d", w.Code)
	}
}

func TestRequire2FA_FirmPartner_WithVerification(t *testing.T) {
	r := newRouter([]string{"FIRM_PARTNER"}, true)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("FIRM_PARTNER with 2FA verified should be allowed; got %d", w.Code)
	}
}

func TestRequire2FA_FirmPartner_Without2FA_Blocked(t *testing.T) {
	r := newRouter([]string{"FIRM_PARTNER"}, false)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("FIRM_PARTNER without 2FA should be blocked (403); got %d", w.Code)
	}
	body := w.Body.String()
	if !containsStr(body, "TWO_FA_REQUIRED") {
		t.Errorf("want error code TWO_FA_REQUIRED in response, got: %s", body)
	}
}

func TestRequire2FA_SuperAdmin_Without2FA_Blocked(t *testing.T) {
	r := newRouter([]string{"SUPER_ADMIN"}, false)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("SUPER_ADMIN without 2FA should be blocked (403); got %d", w.Code)
	}
}

func TestRequire2FA_SuperAdmin_With2FA_Allowed(t *testing.T) {
	r := newRouter([]string{"SUPER_ADMIN"}, true)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("SUPER_ADMIN with 2FA verified should be allowed; got %d", w.Code)
	}
}

func TestRequire2FA_AuditManager_NoVerificationNeeded(t *testing.T) {
	r := newRouter([]string{"AUDIT_MANAGER"}, false)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("AUDIT_MANAGER without 2FA should be allowed; got %d", w.Code)
	}
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// ── AuditIDMiddleware tests ───────────────────────────────────────────────────

func TestAuditIDMiddleware_NoAuditLog_NoHeader(t *testing.T) {
	r := gin.New()
	r.Use(middleware.AuditIDMiddleware())
	r.GET("/ping", func(c *gin.Context) {
		// No audit.Logger.Log() call → slot stays zero
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	if h := w.Header().Get("X-Audit-ID"); h != "" {
		t.Errorf("want no X-Audit-ID when no Log() was called, got %q", h)
	}
}

func TestAuditIDMiddleware_NilLogger_NoHeader(t *testing.T) {
	r := gin.New()
	r.Use(middleware.AuditIDMiddleware())
	r.POST("/mutate", func(c *gin.Context) {
		// nil Logger.Log() is a no-op — slot stays zero
		var l *audit.Logger
		_, _ = l.Log(c.Request.Context(), audit.Entry{Module: "test", Action: "CREATE"})
		c.JSON(http.StatusCreated, gin.H{"ok": true})
	})

	req, _ := http.NewRequest(http.MethodPost, "/mutate", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d", w.Code)
	}
	if h := w.Header().Get("X-Audit-ID"); h != "" {
		t.Errorf("nil logger should not set X-Audit-ID header, got %q", h)
	}
}
