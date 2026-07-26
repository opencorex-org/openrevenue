package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestObservabilitySharesCorrelationAndTraceWithoutSensitiveHeaders(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	handler := Correlation(Observability(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if CorrelationID(r.Context()) != "request-42" || TraceID(r.Context()) != "trace-42" {
			t.Fatal("correlation and trace were not propagated")
		}
		w.WriteHeader(http.StatusNoContent)
	})))
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/returns/7e4b68f5-17ca-48a7-9c03-36281584ad13",
		nil,
	)
	request.Header.Set("X-Correlation-ID", "request-42")
	request.Header.Set("X-Trace-ID", "trace-42")
	request.Header.Set("Authorization", "Bearer must-not-appear")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	output := logs.String()
	for _, required := range []string{"request-42", "trace-42", `"status":204`} {
		if !strings.Contains(output, required) {
			t.Fatalf("missing %q from structured log: %s", required, output)
		}
	}
	if strings.Contains(output, "must-not-appear") || strings.Contains(output, "Authorization") {
		t.Fatalf("sensitive header leaked to log: %s", output)
	}
	if strings.Contains(output, "7e4b68f5-17ca-48a7-9c03-36281584ad13") {
		t.Fatalf("domain identifier leaked to route log: %s", output)
	}
}
