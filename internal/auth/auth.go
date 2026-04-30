package auth

import "net/http"

// Authenticator verifies an incoming HTTP request.
// Implement this interface to add a new authentication strategy.
type Authenticator interface {
	Authenticate(r *http.Request) error
}

// withAuth wraps a handler with authentication middleware.
func WithAuth(a Authenticator, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := a.Authenticate(r); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}
