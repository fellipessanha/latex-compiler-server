package auth

import "net/http"

// NoOp bypasses authentication entirely.
// Use AUTH_PROVIDER=none for local development or trusted-network deployments.
type NoOp struct{}

func (NoOp) Authenticate(_ *http.Request) error { return nil }
