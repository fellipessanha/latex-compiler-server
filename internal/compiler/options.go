package compiler

import "strings"

// CompileOptions controls the latexmk invocation. All fields are optional;
// zero values produce a safe, standard pdflatex compilation.
type CompileOptions struct {
	// Engine selects the LaTeX engine. Supported: "pdflatex" (default), "xelatex", "lualatex".
	Engine string `json:"engine"`

	// ShellEscape enables \write18 / shell escape. Disabled by default for security.
	ShellEscape bool `json:"shell_escape"`

	// ExtraFlags are appended directly to the latexmk invocation
	// (e.g. ["-f", "-recorder"]).
	ExtraFlags []string `json:"extra_flags"`

	// EngineArgs are injected into the engine invocation string that latexmk
	// constructs, after the base safety flags but before %O %S
	// (e.g. ["-8bit", "-etex"]).
	EngineArgs []string `json:"engine_args"`
}

// BuildLatexmkArgs returns the full argument list for a latexmk invocation.
// This is the single source of truth for CLI flag construction; adding support
// for a new engine or flag requires changes only here.
func BuildLatexmkArgs(entry string, outDir string, opts CompileOptions) []string {
	engine := opts.Engine
	if engine == "" {
		engine = "pdflatex"
	}

	shellFlag := "-no-shell-escape"
	if opts.ShellEscape {
		shellFlag = "-shell-escape"
	}

	// Build the engine invocation string passed to latexmk via -pdflatex / -xelatex / -lualatex.
	baseEngineFlags := []string{shellFlag, "-interaction=nonstopmode"}
	engineFlagStr := strings.Join(append(baseEngineFlags, opts.EngineArgs...), " ")
	engineInvocation := engine + " " + engineFlagStr + " %O %S"

	args := []string{"-pdf", "-norc", "-" + engine + "=" + engineInvocation, "-outdir=" + outDir}
	args = append(args, opts.ExtraFlags...)
	args = append(args, entry)

	return args
}
