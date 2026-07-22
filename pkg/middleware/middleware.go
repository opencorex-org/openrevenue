package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/opencorex-org/openrevenue/pkg/id"
)

type correlationTag struct{}

type key string

const CorrelationKey key = "correlation-id"

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
		if v == "" {
			v = id.New[correlationTag]().String()
		}
		w.Header().Set("X-Correlation-ID", v)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), CorrelationKey, v)))
	})
}
func Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") && r.Header.Get("Authorization") == "" {
			http.Error(w, "authorization required", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func CorrelationID(ctx context.Context) string { v, _ := ctx.Value(CorrelationKey).(string); return v }
