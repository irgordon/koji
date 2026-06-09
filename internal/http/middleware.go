package http

import (
	"errors"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"koji/internal/auth"
)

type responseWriterInterceptor struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func newResponseWriterInterceptor(w http.ResponseWriter) *responseWriterInterceptor {
	return &responseWriterInterceptor{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (rw *responseWriterInterceptor) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriterInterceptor) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(b)
	if err == nil {
		rw.bytesWritten += n
	}
	return n, err
}

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("[CRITICAL] Panic recovered: %v\n%s", recovered, debug.Stack())

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":"Internal platform runtime failure"}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func SecurityHeadersMiddleware(devMode bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !devMode {
				applyProductionSecurityHeaders(w)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func StaticPathSafetyMiddleware(devMode bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !devMode && !strings.HasPrefix(r.URL.Path, "/api/") && !isSafeStaticRequestPath(r.URL.Path) {
				http.NotFound(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func applyProductionSecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; connect-src 'self'; base-uri 'none'; form-action 'self'; frame-ancestors 'none'")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Permissions-Policy", "accelerometer=(), camera=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), payment=(), usb=()")
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		interceptor := newResponseWriterInterceptor(w)

		next.ServeHTTP(interceptor, r)

		log.Printf(
			"request_id=%q method=%q path=%q status=%d latency=%q remote=%q bytes=%d",
			requestID(r),
			r.Method,
			r.URL.Path,
			interceptor.statusCode,
			time.Since(start).String(),
			r.RemoteAddr,
			interceptor.bytesWritten,
		)
	})
}

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := requestIDForRequest(r)
		w.Header().Set(requestIDHeader, id)
		next.ServeHTTP(w, withRequestID(r, id))
	})
}

func applyMiddlewareChain(mux *http.ServeMux, authStore *auth.Store, devMode bool) http.Handler {
	var h http.Handler = mux
	h = AuthGateMiddleware(authStore, devMode)(h)
	h = StaticPathSafetyMiddleware(devMode)(h)
	h = LoggingMiddleware(h)
	h = RecoveryMiddleware(h)
	h = RequestIDMiddleware(h)
	h = SecurityHeadersMiddleware(devMode)(h)
	return h
}

func AuthGateMiddleware(store *auth.Store, devMode bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if devMode {
				w.Header().Set("X-Koji-Auth-Bypass", "dev")
				next.ServeHTTP(w, r)
				return
			}

			if isPublicAuthSurface(r) {
				next.ServeHTTP(w, r)
				return
			}

			sessionID, ok := sessionIDFromRequest(r)
			if !ok {
				writeAuthDenied(w, r)
				return
			}
			principal, err := store.ValidateSession(r.Context(), sessionID)
			if err != nil {
				writeAuthDenied(w, r)
				return
			}
			if requiresCSRF(r) {
				token := r.Header.Get(auth.CSRFHeaderName)
				if err := store.ValidateCSRF(r.Context(), sessionID, token); err != nil {
					writeCSRFDenied(w)
					return
				}
			}

			next.ServeHTTP(w, withPrincipal(r, principal))
		})
	}
}

func isPublicAuthSurface(r *http.Request) bool {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/healthz":
		return true
	case r.Method == http.MethodGet && r.URL.Path == "/readyz":
		return true
	case r.Method == http.MethodPost && r.URL.Path == "/api/bootstrap":
		return true
	case r.Method == http.MethodPost && r.URL.Path == "/api/login":
		return true
	case r.Method == http.MethodPost && r.URL.Path == "/api/logout":
		return true
	case r.Method == http.MethodGet && r.URL.Path == "/api/session":
		return true
	default:
		return false
	}
}

func sessionIDFromRequest(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(auth.SessionCookieName)
	if err != nil || cookie.Value == "" {
		return "", false
	}
	return cookie.Value, true
}

func requiresCSRF(r *http.Request) bool {
	switch r.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return false
	default:
		return true
	}
}

func writeAuthDenied(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		writeJSONError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	http.Error(w, "Authentication required", http.StatusUnauthorized)
}

func writeCSRFDenied(w http.ResponseWriter) {
	writeJSONError(w, http.StatusForbidden, "CSRF token required")
}

func authStatusCode(err error) int {
	if errors.Is(err, auth.ErrInvalidCSRF) {
		return http.StatusForbidden
	}
	return http.StatusUnauthorized
}
