package httpapi

import "net/http"

// securityHeaders sets conservative defaults on every response. Handlers that need different
// values (for example the Swagger UI assets) override them afterwards.
// HSTS is deliberately absent: it belongs to the TLS-terminating ingress.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Cache-Control", "no-store")
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}
