package handlers

import (
	"archive/tar"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// --- validateTargetDir ---

func TestValidateTargetDir_Valid(t *testing.T) {
	cases := []string{"subdir", "nested/subdir", "./subdir", "a/b/c"}
	for _, tc := range cases {
		if err := validateTargetDir(tc); err != nil {
			t.Errorf("validateTargetDir(%q): unexpected error: %v", tc, err)
		}
	}
}

func TestValidateTargetDir_AbsolutePath(t *testing.T) {
	if err := validateTargetDir("/absolute"); err == nil {
		t.Error("expected error for absolute path, got nil")
	}
}

func TestValidateTargetDir_DotDotEscape(t *testing.T) {
	cases := []string{"../escape", "../../escape", "valid/../../escape"}
	for _, tc := range cases {
		if err := validateTargetDir(tc); err == nil {
			t.Errorf("validateTargetDir(%q): expected error, got nil", tc)
		}
	}
}

// --- embedToken ---

func TestEmbedToken_HTTPS(t *testing.T) {
	got := embedToken("https://github.com/user/repo.git", "tok")
	want := "https://tok@github.com/user/repo.git"
	if got != want {
		t.Errorf("embedToken: got %q, want %q", got, want)
	}
}

func TestEmbedToken_HTTP(t *testing.T) {
	got := embedToken("http://example.com/repo.git", "tok")
	want := "http://tok@example.com/repo.git"
	if got != want {
		t.Errorf("embedToken: got %q, want %q", got, want)
	}
}

func TestEmbedToken_SSH_Unchanged(t *testing.T) {
	url := "ssh://git@github.com/user/repo.git"
	if got := embedToken(url, "tok"); got != url {
		t.Errorf("embedToken on SSH: expected unchanged, got %q", got)
	}
}

// --- isHex ---

func TestIsHex_Valid(t *testing.T) {
	if !isHex("0123456789abcdefABCDEF") {
		t.Error("expected true for valid hex string")
	}
}

func TestIsHex_Invalid(t *testing.T) {
	cases := []string{"xyz", "0g", "hello", "abcdefg"}
	for _, tc := range cases {
		if isHex(tc) {
			t.Errorf("isHex(%q): expected false", tc)
		}
	}
}

// --- extractTarStream ---

func TestExtractTarStream_ExtractsFiles(t *testing.T) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	files := map[string]string{
		"main.tex": `\documentclass{article}\begin{document}Hello\end{document}`,
		"sub/fig.tex": `figure content`,
	}
	for name, content := range files {
		hdr := &tar.Header{Name: name, Size: int64(len(content)), Mode: 0o644, Typeflag: tar.TypeReg}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	tw.Close()

	destDir := t.TempDir()
	if err := extractTarStream(&buf, destDir); err != nil {
		t.Fatalf("extractTarStream: %v", err)
	}

	for name, want := range files {
		data, err := os.ReadFile(filepath.Join(destDir, name))
		if err != nil {
			t.Errorf("file %q not extracted: %v", name, err)
			continue
		}
		if string(data) != want {
			t.Errorf("file %q: got %q, want %q", name, data, want)
		}
	}
}
