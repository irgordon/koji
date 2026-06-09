package http

import (
	"context"
	nethttp "net/http"

	"koji/internal/auth"
)

type principalContextKey struct{}
type requestIDContextKey struct{}

func withPrincipal(r *nethttp.Request, principal auth.Principal) *nethttp.Request {
	ctx := context.WithValue(r.Context(), principalContextKey{}, principal)
	return r.WithContext(ctx)
}

func principalFromRequest(r *nethttp.Request) (auth.Principal, bool) {
	principal, ok := r.Context().Value(principalContextKey{}).(auth.Principal)
	return principal, ok
}

func withRequestID(r *nethttp.Request, id string) *nethttp.Request {
	ctx := context.WithValue(r.Context(), requestIDContextKey{}, id)
	return r.WithContext(ctx)
}

func requestIDFromRequest(r *nethttp.Request) (string, bool) {
	id, ok := r.Context().Value(requestIDContextKey{}).(string)
	return id, ok && id != ""
}
