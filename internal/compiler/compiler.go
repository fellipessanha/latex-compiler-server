package compiler

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
)

// Compile runs latexmk on the source files in ws.InDir, writing output to ws.OutDir.
// It returns the output directory path on success, or the combined stdout+stderr log
// wrapped in an error on failure.
// The context controls the compilation timeout; when it expires the latexmk process
// is killed and a context error is returned.
func Compile(ctx context.Context, ws Workspace, opts CompileOptions, entry string) (string, error) {
	args := BuildLatexmkArgs(entry, ws.OutDir, opts)

	cmd := exec.CommandContext(ctx, "latexmk", args...)
	cmd.Dir = ws.InDir

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		log := out.String()
		if ctx.Err() != nil {
			return "", fmt.Errorf("timeout: %w\n%s", ctx.Err(), log)
		}
		return "", fmt.Errorf("%s", log)
	}

	// Derive the expected PDF name from the entry filename.
	base := filepath.Base(entry)
	ext := filepath.Ext(base)
	pdfName := base[:len(base)-len(ext)] + ".pdf"
	_ = pdfName // callers locate outputs via ws.OutDir; pdfName is for reference

	return ws.OutDir, nil
}
