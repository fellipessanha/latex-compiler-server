package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// --- BearerAuthenticator ---

func TestBearerAuthenticate_CorrectToken(t *testing.T) {
	a := &BearerAuthenticator{secret: "my-secret"}
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	r.Header.Set("Authorization", "Bearer my-secret")
	if err := a.Authenticate(r); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestBearerAuthenticate_WrongToken(t *testing.T) {
	a := &BearerAuthenticator{secret: "my-secret"}
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	r.Header.Set("Authorization", "Bearer wrong")
	if err := a.Authenticate(r); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestBearerAuthenticate_MissingHeader(t *testing.T) {
	a := &BearerAuthenticator{secret: "my-secret"}
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	if err := a.Authenticate(r); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestBearerAuthenticate_MalformedHeader(t *testing.T) {
	a := &BearerAuthenticator{secret: "my-secret"}
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	r.Header.Set("Authorization", "my-secret") // missing "Bearer " prefix
	if err := a.Authenticate(r); err == nil {
		t.Fatal("expected error for missing Bearer prefix, got nil")
	}
}

// --- NoOp ---

func TestNoOpAuthenticate_AlwaysNil(t *testing.T) {
	a := NoOp{}
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	if err := a.Authenticate(r); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

// --- Factory ---

func TestNew_BearerWithSecret(t *testing.T) {
	t.Setenv("AUTH_PROVIDER", "bearer")
	t.Setenv("API_SECRET", "test-secret")
	a, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := a.(*BearerAuthenticator); !ok {
		t.Fatalf("expected *BearerAuthenticator, got %T", a)
	}
}

func TestNew_BearerDefault(t *testing.T) {
	t.Setenv("AUTH_PROVIDER", "")
	t.Setenv("API_SECRET", "test-secret")
	a, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := a.(*BearerAuthenticator); !ok {
		t.Fatalf("expected *BearerAuthenticator, got %T", a)
	}
}

func TestNew_NoOp(t *testing.T) {
	t.Setenv("AUTH_PROVIDER", "none")
	a, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := a.(NoOp); !ok {
		t.Fatalf("expected NoOp, got %T", a)
	}
}

func TestNew_UnknownProvider(t *testing.T) {
	t.Setenv("AUTH_PROVIDER", "magic")
	if _, err := New(); err == nil {
		t.Fatal("expected error for unknown provider, got nil")
	}
}

func TestNew_BearerWithoutSecret(t *testing.T) {
	t.Setenv("AUTH_PROVIDER", "bearer")
	t.Setenv("API_SECRET", "")
	if _, err := New(); err == nil {
		t.Fatal("expected error when API_SECRET is empty, got nil")
	}
}
