package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"
	foundation "github.com/opencorex-org/openrevenue/pkg/domain"
	"github.com/opencorex-org/openrevenue/pkg/id"
)

type correlationTag struct{}

type key string

const CorrelationKey key = "correlation-id"
const traceKey key = "trace-id"
const domainContextKey key = "domain-context"

var rejectedDomainContexts atomic.Uint64
var identifierSegment = regexp.MustCompile(`(?i)/[0-9a-f]{8}-[0-9a-f-]{27,}(/|$)`)

func Security(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}
func Correlation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := r.Header.Get("X-Correlation-ID")
		if parsed, err := foundation.NewCorrelationID(v); v == "" || err != nil {
			v = id.New[correlationTag]().String()
		} else {
			v = parsed.String()
		}
		w.Header().Set("X-Correlation-ID", v)
		traceID := r.Header.Get("X-Trace-ID")
		if _, err := foundation.NewCorrelationID(traceID); traceID == "" || err != nil {
			traceID = v
		}
		w.Header().Set("X-Trace-ID", traceID)
		ctx := context.WithValue(r.Context(), CorrelationKey, v)
		ctx = context.WithValue(ctx, traceKey, traceID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func Observability(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		writer := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(writer, r)
		route := chi.RouteContext(r.Context()).RoutePattern()
		if route == "" {
			route = identifierSegment.ReplaceAllString(r.URL.Path, "/{id}$1")
		}
		slog.InfoContext(r.Context(), "http request",
			"method", r.Method,
			"route", route,
			"status", writer.status,
			"duration_ms", time.Since(started).Milliseconds(),
			"correlation_id", CorrelationID(r.Context()),
			"trace_id", TraceID(r.Context()),
		)
	})
}
func Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") && r.Header.Get("Authorization") == "" {
			w.Header().Set("Content-Type", "application/problem+json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"type":  "https://openrevenue.org/problems/unauthorized",
				"title": "Authentication required", "status": http.StatusUnauthorized,
				"correlationId": CorrelationID(r.Context()),
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}
func CorrelationID(ctx context.Context) string { v, _ := ctx.Value(CorrelationKey).(string); return v }
func TraceID(ctx context.Context) string       { v, _ := ctx.Value(traceKey).(string); return v }

func RequireDomainContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenant, err := foundation.NewTenantID(r.Header.Get("X-Tenant-ID"))
		if err != nil {
			writeContextProblem(w, "A valid X-Tenant-ID header is required.")
			return
		}
		jurisdiction, err := foundation.NewJurisdiction(r.Header.Get("X-Jurisdiction-Code"))
		if err != nil {
			writeContextProblem(w, "A valid X-Jurisdiction-Code header is required.")
			return
		}
		actorKind := foundation.ActorKind(r.Header.Get("X-Actor-Type"))
		if actorKind == "" {
			actorKind = foundation.ActorUser
		}
		actor, err := foundation.NewActor(actorKind, r.Header.Get("X-Actor-ID"))
		if err != nil {
			writeContextProblem(w, "A valid X-Actor-ID and X-Actor-Type are required.")
			return
		}
		correlation, err := foundation.NewCorrelationID(CorrelationID(r.Context()))
		if err != nil {
			writeContextProblem(w, "X-Correlation-ID is invalid.")
			return
		}
		scope, err := foundation.NewContext(tenant, jurisdiction, actor, correlation)
		if err != nil {
			writeContextProblem(w, "The request domain context is invalid.")
			return
		}
		if value := r.Header.Get("X-Causation-ID"); value != "" {
			causation, causationErr := foundation.NewCorrelationID(value)
			if causationErr != nil {
				writeContextProblem(w, "X-Causation-ID is invalid.")
				return
			}
			scope, err = scope.WithCausationID(causation)
			if err != nil {
				writeContextProblem(w, "X-Causation-ID is invalid.")
				return
			}
		}
		ctx := context.WithValue(r.Context(), domainContextKey, scope)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func DomainContext(ctx context.Context) (foundation.Context, bool) {
	scope, ok := ctx.Value(domainContextKey).(foundation.Context)
	return scope, ok
}

func writeContextProblem(w http.ResponseWriter, detail string) {
	rejectedDomainContexts.Add(1)
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"type":          "https://openrevenue.org/problems/domain-context",
		"title":         "Invalid domain context",
		"status":        http.StatusBadRequest,
		"detail":        detail,
		"correlationId": w.Header().Get("X-Correlation-ID"),
	})
}

func RejectedDomainContexts() uint64 {
	return rejectedDomainContexts.Load()
}
