package auth

import (
	"errors"
	"net/http"
	"os"
	"strings"
)

// BearerAuthenticator validates requests using a shared secret bearer token.
// The secret is read from the API_SECRET environment variable.
type BearerAuthenticator struct {
	secret string
}

// NewBearer creates a BearerAuthenticator. Returns an error if API_SECRET is unset.
func NewBearer() (*BearerAuthenticator, error) {
	secret := os.Getenv("API_SECRET")
	if secret == "" {
		return nil, errors.New("API_SECRET environment variable is required for AUTH_PROVIDER=bearer")
	}
	return &BearerAuthenticator{secret: secret}, nil
}

// Authenticate checks the Authorization header for a matching bearer token.
func (b *BearerAuthenticator) Authenticate(r *http.Request) error {
	header := r.Header.Get("Authorization")
	token, found := strings.CutPrefix(header, "Bearer ")
	if !found || token != b.secret {
		return errors.New("unauthorized")
	}
	return nil
}
