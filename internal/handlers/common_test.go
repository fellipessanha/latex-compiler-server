package handlers

import (
	"net/http/httptest"
	"testing"
	"time"
)

// --- resolveOutput ---

func TestResolveOutput_DefaultsPDF(t *testing.T) {
	r := httptest.NewRequest("POST", "/git", nil)
	// No {output} path value set — PathValue returns "".
	if got := resolveOutput(r); got != "pdf" {
		t.Errorf("expected pdf, got %q", got)
	}
}

func TestResolveOutput_ReturnsPathValue(t *testing.T) {
	r := httptest.NewRequest("POST", "/git/png", nil)
	r.SetPathValue("output", "png")
	if got := resolveOutput(r); got != "png" {
		t.Errorf("expected png, got %q", got)
	}
}

// --- compilationTimeout ---

func TestCompilationTimeout_Default(t *testing.T) {
	t.Setenv("COMPILE_TIMEOUT_SECONDS", "")
	if got := compilationTimeout(); got != 60*time.Second {
		t.Errorf("expected 60s, got %v", got)
	}
}

func TestCompilationTimeout_CustomValue(t *testing.T) {
	t.Setenv("COMPILE_TIMEOUT_SECONDS", "120")
	if got := compilationTimeout(); got != 120*time.Second {
		t.Errorf("expected 120s, got %v", got)
	}
}

func TestCompilationTimeout_InvalidValue(t *testing.T) {
	t.Setenv("COMPILE_TIMEOUT_SECONDS", "notanumber")
	if got := compilationTimeout(); got != 60*time.Second {
		t.Errorf("expected fallback 60s, got %v", got)
	}
}
