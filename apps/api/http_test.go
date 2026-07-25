package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	app "github.com/opencorex-org/openrevenue/internal/administration/application"
)

func TestOperationalAndAuthenticationBoundaries(t *testing.T) {
	router := Router(app.New(nil))
	health := httptest.NewRecorder()
	router.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/health", nil))
	if health.Code != http.StatusOK {
		t.Fatalf("health status = %d", health.Code)
	}
	if health.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("secure headers missing")
	}

	protected := httptest.NewRecorder()
	router.ServeHTTP(protected, httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit-events", nil))
	if protected.Code != http.StatusUnauthorized {
		t.Fatalf("protected status = %d", protected.Code)
	}
}
