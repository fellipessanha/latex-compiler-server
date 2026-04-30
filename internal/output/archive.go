package output

import (
	"archive/tar"
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

// ArchiveFormatter compresses the full output directory and streams the result.
// Kind must be one of: "zip", "tar", "7zip".
type ArchiveFormatter struct {
	Kind string
}

func (a ArchiveFormatter) Format(outDir string, entry string, opts OutputOptions, w http.ResponseWriter) error {
	// Optionally generate a PNG thumbnail and include it in the archive.
	if opts.IncludePNG {
		pngPath, err := renderPNG(outDir, entry)
		if err != nil {
			return fmt.Errorf("render thumbnail: %w", err)
		}
		defer os.Remove(pngPath)
		// pngPath is already inside outDir, so it will be picked up automatically.
	}

	switch a.Kind {
	case "zip":
		return writeZip(outDir, w)
	case "tar":
		return writeTar(outDir, w)
	case "7zip":
		return write7zip(outDir, w)
	default:
		return fmt.Errorf("unknown archive kind: %s", a.Kind)
	}
}

func writeZip(outDir string, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/zip")
	w.WriteHeader(http.StatusOK)

	zw := zip.NewWriter(w)
	defer zw.Close()

	return filepath.Walk(outDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(outDir, path)
		fw, err := zw.Create(rel)
		if err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(fw, f)
		return err
	})
}

func writeTar(outDir string, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/x-tar")
	w.WriteHeader(http.StatusOK)

	tw := tar.NewWriter(w)
	defer tw.Close()

	return filepath.Walk(outDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(outDir, path)
		hdr := &tar.Header{
			Name: rel,
			Mode: int64(info.Mode()),
			Size: info.Size(),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(tw, f)
		return err
	})
}

func write7zip(outDir string, w http.ResponseWriter) error {
	// 7z cannot stream to stdout reliably; write to a temp file then stream.
	tmp, err := os.CreateTemp("", "latex-out-*.7z")
	if err != nil {
		return fmt.Errorf("create temp 7z: %w", err)
	}
	tmpPath := tmp.Name()
	tmp.Close()
	defer os.Remove(tmpPath)

	cmd := exec.Command("7z", "a", "-bd", tmpPath, filepath.Join(outDir, "*"))
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("7z: %w\n%s", err, out)
	}

	f, err := os.Open(tmpPath)
	if err != nil {
		return fmt.Errorf("open 7z output: %w", err)
	}
	defer f.Close()

	w.Header().Set("Content-Type", "application/x-7z-compressed")
	w.WriteHeader(http.StatusOK)
	_, err = io.Copy(w, f)
	return err
}
