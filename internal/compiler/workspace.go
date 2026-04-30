package compiler

import (
	"fmt"
	"os"
)

// Workspace holds a pair of temporary directories for one compilation job:
// InDir for the source files and OutDir for latexmk output.
// Call Close() (typically via defer) to remove both directories.
type Workspace struct {
	InDir  string
	OutDir string
}

// NewWorkspace creates two temporary directories under os.TempDir().
func NewWorkspace() (Workspace, error) {
	inDir, err := os.MkdirTemp("", "latex-in-*")
	if err != nil {
		return Workspace{}, fmt.Errorf("create input workspace: %w", err)
	}

	outDir, err := os.MkdirTemp("", "latex-out-*")
	if err != nil {
		_ = os.RemoveAll(inDir)
		return Workspace{}, fmt.Errorf("create output workspace: %w", err)
	}

	return Workspace{InDir: inDir, OutDir: outDir}, nil
}

// Close removes both temporary directories. Safe to call on a zero Workspace.
func (w Workspace) Close() error {
	var inErr, outErr error
	if w.InDir != "" {
		inErr = os.RemoveAll(w.InDir)
	}
	if w.OutDir != "" {
		outErr = os.RemoveAll(w.OutDir)
	}
	if inErr != nil {
		return inErr
	}
	return outErr
}
