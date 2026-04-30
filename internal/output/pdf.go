package output

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// PDFFormatter streams the compiled PDF to the response.
type PDFFormatter struct{}

func (PDFFormatter) Format(outDir string, entry string, _ OutputOptions, w http.ResponseWriter) error {
	pdfPath, err := findPDF(outDir, entry)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(pdfPath)
	if err != nil {
		return fmt.Errorf("read pdf: %w", err)
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(data)
	return err
}

// findPDF locates the PDF in outDir that corresponds to the given entry filename.
func findPDF(outDir, entry string) (string, error) {
	base := filepath.Base(entry)
	ext := filepath.Ext(base)
	pdfName := strings.TrimSuffix(base, ext) + ".pdf"
	pdfPath := filepath.Join(outDir, pdfName)

	if _, err := os.Stat(pdfPath); err != nil {
		return "", fmt.Errorf("pdf not found at %s: %w", pdfPath, err)
	}
	return pdfPath, nil
}
