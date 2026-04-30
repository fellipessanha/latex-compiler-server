package handlers

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// --- detectArchiveType ---

func TestDetectArchiveType_ContentType(t *testing.T) {
	cases := []struct {
		ct   string
		want string
	}{
		{"application/zip", "zip"},
		{"application/x-7z-compressed", "7zip"},
		{"application/x-tar", "tar"},
		{"application/gzip", "tar"},
	}
	for _, tc := range cases {
		got, err := detectArchiveType(tc.ct, nil)
		if err != nil {
			t.Errorf("detectArchiveType(%q): unexpected error: %v", tc.ct, err)
			continue
		}
		if got != tc.want {
			t.Errorf("detectArchiveType(%q): got %q, want %q", tc.ct, got, tc.want)
		}
	}
}

func TestDetectArchiveType_ZipMagicBytes(t *testing.T) {
	data := []byte{0x50, 0x4B, 0x03, 0x04, 0x00, 0x00}
	got, err := detectArchiveType("", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "zip" {
		t.Errorf("got %q, want zip", got)
	}
}

func TestDetectArchiveType_SevenZipMagicBytes(t *testing.T) {
	data := []byte{0x37, 0x7A, 0xBC, 0xAF, 0x27, 0x1C}
	got, err := detectArchiveType("", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "7zip" {
		t.Errorf("got %q, want 7zip", got)
	}
}

func TestDetectArchiveType_Unknown(t *testing.T) {
	data := []byte("just some random text")
	if _, err := detectArchiveType("", data); err == nil {
		t.Error("expected error for unknown archive type, got nil")
	}
}

// --- extractZip ---

func TestExtractZip_ExtractsFiles(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	files := map[string]string{
		"main.tex":    `\documentclass{article}`,
		"refs/bib.bib": `@article{key, title={Test}}`,
	}
	for name, content := range files {
		fw, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	zw.Close()

	destDir := t.TempDir()
	if err := extractZip(buf.Bytes(), destDir); err != nil {
		t.Fatalf("extractZip: %v", err)
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

// --- extractTar ---

func TestExtractTar_ExtractsFiles(t *testing.T) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	files := map[string]string{
		"main.tex": `\begin{document}`,
		"img/fig.tex": `figure`,
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
	if err := extractTar(buf.Bytes(), destDir); err != nil {
		t.Fatalf("extractTar: %v", err)
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
