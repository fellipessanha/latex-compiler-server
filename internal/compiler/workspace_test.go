package compiler

import (
	"os"
	"testing"
)

func TestNewWorkspace_CreatesDirs(t *testing.T) {
	ws, err := NewWorkspace()
	if err != nil {
		t.Fatalf("NewWorkspace: %v", err)
	}
	defer ws.Close()

	if ws.InDir == "" || ws.OutDir == "" {
		t.Fatal("expected non-empty dir paths")
	}
	if ws.InDir == ws.OutDir {
		t.Fatal("InDir and OutDir must be distinct")
	}
	if _, err := os.Stat(ws.InDir); err != nil {
		t.Errorf("InDir does not exist: %v", err)
	}
	if _, err := os.Stat(ws.OutDir); err != nil {
		t.Errorf("OutDir does not exist: %v", err)
	}
}

func TestWorkspaceClose_RemovesDirs(t *testing.T) {
	ws, err := NewWorkspace()
	if err != nil {
		t.Fatalf("NewWorkspace: %v", err)
	}

	inDir := ws.InDir
	outDir := ws.OutDir

	if err := ws.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if _, err := os.Stat(inDir); !os.IsNotExist(err) {
		t.Errorf("expected InDir to be removed, got: %v", err)
	}
	if _, err := os.Stat(outDir); !os.IsNotExist(err) {
		t.Errorf("expected OutDir to be removed, got: %v", err)
	}
}
