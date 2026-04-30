package compiler

import (
	"strings"
	"testing"
)

func args(entry, outDir string, opts CompileOptions) []string {
	return BuildLatexmkArgs(entry, outDir, opts)
}

func containsArg(args []string, sub string) bool {
	for _, a := range args {
		if strings.Contains(a, sub) {
			return true
		}
	}
	return false
}

func TestBuildLatexmkArgs_DefaultEngine(t *testing.T) {
	a := args("main.tex", "/out", CompileOptions{})
	if !containsArg(a, "pdflatex") {
		t.Errorf("expected pdflatex in args, got %v", a)
	}
}

func TestBuildLatexmkArgs_XelatexEngine(t *testing.T) {
	a := args("main.tex", "/out", CompileOptions{Engine: "xelatex"})
	if !containsArg(a, "-xelatex=") {
		t.Errorf("expected -xelatex= flag, got %v", a)
	}
}

func TestBuildLatexmkArgs_ShellEscapeDisabled(t *testing.T) {
	a := args("main.tex", "/out", CompileOptions{ShellEscape: false})
	if !containsArg(a, "-no-shell-escape") {
		t.Errorf("expected -no-shell-escape in engine invocation, got %v", a)
	}
}

func TestBuildLatexmkArgs_ShellEscapeEnabled(t *testing.T) {
	a := args("main.tex", "/out", CompileOptions{ShellEscape: true})
	if containsArg(a, "-no-shell-escape") {
		t.Errorf("expected -no-shell-escape to be absent, got %v", a)
	}
	if !containsArg(a, "-shell-escape") {
		t.Errorf("expected -shell-escape in engine invocation, got %v", a)
	}
}

func TestBuildLatexmkArgs_ExtraFlags(t *testing.T) {
	a := args("main.tex", "/out", CompileOptions{ExtraFlags: []string{"-f", "-recorder"}})
	if !containsArg(a, "-f") || !containsArg(a, "-recorder") {
		t.Errorf("expected extra flags in args, got %v", a)
	}
}

func TestBuildLatexmkArgs_EngineArgs(t *testing.T) {
	a := args("main.tex", "/out", CompileOptions{EngineArgs: []string{"-8bit"}})
	if !containsArg(a, "-8bit") {
		t.Errorf("expected engine arg -8bit in invocation, got %v", a)
	}
	// engine args must appear before %O %S
	for _, arg := range a {
		if strings.Contains(arg, "-8bit") {
			if !strings.Contains(arg, "%O %S") {
				t.Errorf("expected %%O %%S to follow engine args in same flag, got %q", arg)
			}
		}
	}
}

func TestBuildLatexmkArgs_EntryAndOutDir(t *testing.T) {
	a := args("thesis.tex", "/output/dir", CompileOptions{})
	last := a[len(a)-1]
	if last != "thesis.tex" {
		t.Errorf("expected entry file as last arg, got %q", last)
	}
	if !containsArg(a, "-outdir=/output/dir") {
		t.Errorf("expected -outdir in args, got %v", a)
	}
}
