package auth

import (
	"fmt"
	"os"
)

// New reads the AUTH_PROVIDER environment variable and returns the corresponding
// Authenticator. Adding a new strategy requires only a new file in this package
// and a new case here — no other code needs to change.
//
// Supported values:
//   - "bearer" (default): validates Authorization: Bearer <token> against API_SECRET
//   - "none": no authentication (dev / trusted-network deployments)
func New() (Authenticator, error) {
	provider := os.Getenv("AUTH_PROVIDER")
	switch provider {
	case "bearer", "":
		return NewBearer()
	case "none":
		return NoOp{}, nil
	default:
		return nil, fmt.Errorf("unknown AUTH_PROVIDER: %q", provider)
	}
}
