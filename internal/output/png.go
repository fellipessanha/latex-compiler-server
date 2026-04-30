package output

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

// PNGFormatter renders page 1 of the compiled PDF as a PNG thumbnail
// using pdftoppm (from poppler-utils).
type PNGFormatter struct{}

func (PNGFormatter) Format(outDir string, entry string, _ OutputOptions, w http.ResponseWriter) error {
	pngPath, err := renderPNG(outDir, entry)
	if err != nil {
		return err
	}
	defer os.Remove(pngPath)

	data, err := os.ReadFile(pngPath)
	if err != nil {
		return fmt.Errorf("read png: %w", err)
	}

	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(data)
	return err
}

// renderPNG converts page 1 of the PDF to PNG at 150 dpi using pdftoppm.
// Returns the path to the generated file; the caller is responsible for removal.
func renderPNG(outDir, entry string) (string, error) {
	pdfPath, err := findPDF(outDir, entry)
	if err != nil {
		return "", err
	}

	// pdftoppm appends "-1" to the output stem, producing <stem>-1.png
	stem := filepath.Join(outDir, "thumbnail")
	cmd := exec.Command("pdftoppm",
		"-png",
		"-r", "150",
		"-singlefile",
		"-f", "1", "-l", "1",
		pdfPath, stem,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("pdftoppm: %w\n%s", err, out)
	}

	pngPath := stem + ".png"
	if _, err := os.Stat(pngPath); err != nil {
		return "", fmt.Errorf("png not produced at %s: %w", pngPath, err)
	}
	return pngPath, nil
}
