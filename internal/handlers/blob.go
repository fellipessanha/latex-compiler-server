package handlers

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"latex-compiler-api/internal/compiler"
	"latex-compiler-api/internal/output"
)

type blobOptions struct {
	Entry       string                  `json:"entry"`
	CompileOpts compiler.CompileOptions `json:"compile_options"`
	OutputOpts  output.OutputOptions    `json:"output_options"`
}

// Blob handles POST /blob and POST /blob/{output}.
// The request must be multipart/form-data with:
//   - "file": the compressed archive bytes
//   - "options": JSON-encoded blobOptions
//
// @Summary   Compile LaTeX from an uploaded archive
// @Tags      compile
// @Accept    multipart/form-data
// @Produce   application/pdf,image/png,application/zip,application/x-tar,application/x-7z-compressed
// @Param     output   path      string  false  "Output format (pdf, png, zip, tar, 7zip). Defaults to pdf."  Enums(pdf,png,zip,tar,7zip)
// @Param     file     formData  file    true   "Compressed source archive (zip, tar, or 7zip)"
// @Param     options  formData  string  false  "JSON-encoded compile and output options"  example({"entry":"main.tex","compile_options":{},"output_options":{}})
// @Success   200
// @Failure   400  {object}  errorResponse
// @Failure   401  {object}  errorResponse
// @Failure   422  {object}  errorResponse
// @Failure   504  {object}  errorResponse
// @Security  BearerAuth
// @Router    /blob/{output} [post]
// @Router    /blob [post]
func Blob(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(128 << 20); err != nil { // 128 MB max in memory
		writeError(w, http.StatusBadRequest, "invalid multipart form", err.Error())
		return
	}

	// Parse options part.
	var opts blobOptions
	if s := r.FormValue("options"); s != "" {
		if err := json.Unmarshal([]byte(s), &opts); err != nil {
			writeError(w, http.StatusBadRequest, "invalid options JSON", err.Error())
			return
		}
	}

	entry := opts.Entry
	if entry == "" {
		entry = "main.tex"
	}

	// Read the uploaded file.
	file, fh, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing file part", err.Error())
		return
	}
	defer file.Close()

	// Read into memory to allow magic-byte detection.
	data, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusBadRequest, "reading file", err.Error())
		return
	}

	archiveType, err := detectArchiveType(fh.Header.Get("Content-Type"), data)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), "")
		return
	}

	outFormat := resolveOutput(r)

	compileAndFormat(w, r, outFormat, entry, opts.CompileOpts, opts.OutputOpts,
		func(ws compiler.Workspace) error {
			return extractArchive(data, archiveType, ws.InDir)
		},
	)
}

// detectArchiveType determines the archive format from Content-Type header first,
// then falls back to magic bytes.
func detectArchiveType(contentType string, data []byte) (string, error) {
	switch contentType {
	case "application/zip":
		return "zip", nil
	case "application/x-7z-compressed":
		return "7zip", nil
	case "application/x-tar", "application/gzip":
		return "tar", nil
	}

	// Magic byte detection.
	if len(data) >= 4 && bytes.Equal(data[:4], []byte{0x50, 0x4B, 0x03, 0x04}) {
		return "zip", nil
	}
	if len(data) >= 6 && bytes.Equal(data[:6], []byte{0x37, 0x7A, 0xBC, 0xAF, 0x27, 0x1C}) {
		return "7zip", nil
	}
	if len(data) >= 262 && string(data[257:262]) == "ustar" {
		return "tar", nil
	}

	return "", fmt.Errorf("unrecognised archive type; provide Content-Type or use zip/tar/7zip")
}

func extractArchive(data []byte, kind, destDir string) error {
	log.Printf("extracting %s archive (%d bytes)", kind, len(data))
	switch kind {
	case "zip":
		return extractZip(data, destDir)
	case "tar":
		return extractTar(data, destDir)
	case "7zip":
		return extract7zip(data, destDir)
	default:
		return fmt.Errorf("unknown archive kind: %s", kind)
	}
}

func extractZip(data []byte, destDir string) error {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		dest := filepath.Join(destDir, f.Name)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.Create(dest)
		if err != nil {
			rc.Close()
			return err
		}
		n, err := io.Copy(out, rc)
		rc.Close()
		out.Close()
		if err != nil {
			return err
		}
		log.Printf("  extracted (zip): %s (%d bytes)", f.Name, n)
	}
	return nil
}

func extractTar(data []byte, destDir string) error {
	tr := tar.NewReader(bytes.NewReader(data))
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar: %w", err)
		}
		if hdr.Typeflag == tar.TypeDir {
			continue
		}
		dest := filepath.Join(destDir, hdr.Name)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		out, err := os.Create(dest)
		if err != nil {
			return err
		}
		n, err := io.Copy(out, tr)
		out.Close()
		if err != nil {
			return err
		}
		log.Printf("  extracted (tar): %s (%d bytes)", hdr.Name, n)
	}
	return nil
}

func extract7zip(data []byte, destDir string) error {
	// Write to a temp file because 7z does not read from stdin.
	tmp, err := os.CreateTemp("", "latex-in-*.7z")
	if err != nil {
		return fmt.Errorf("create temp file for 7z: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp 7z: %w", err)
	}
	tmp.Close()

	cmd := exec.Command("7z", "x", "-bd", "-o"+destDir, tmpPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("7z extract: %w\n%s", err, out)
	}
	_ = filepath.Walk(destDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(destDir, path)
		log.Printf("  extracted (7zip): %s (%d bytes)", rel, info.Size())
		return nil
	})
	return nil
}
