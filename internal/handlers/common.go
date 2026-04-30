package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"latex-compiler-api/internal/compiler"
	"latex-compiler-api/internal/output"
)

// compilationTimeout reads COMPILE_TIMEOUT_SECONDS from the environment.
// Defaults to 60 seconds if unset or invalid.
func compilationTimeout() time.Duration {
	if s := os.Getenv("COMPILE_TIMEOUT_SECONDS"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			return time.Duration(n) * time.Second
		}
	}
	return 60 * time.Second
}

// resolveOutput returns the output format from the {output} path parameter,
// defaulting to "pdf" when the segment is empty.
func resolveOutput(r *http.Request) string {
	out := r.PathValue("output")
	if out == "" {
		return "pdf"
	}
	return out
}

// compileAndFormat runs the compilation and streams the formatted result.
// It handles timeout context creation, workspace lifecycle, and error responses.
func compileAndFormat(
	w http.ResponseWriter,
	r *http.Request,
	outFormat string,
	entry string,
	compileOpts compiler.CompileOptions,
	outputOpts output.OutputOptions,
	populate func(ws compiler.Workspace) error,
) {
	formatter, ok := output.Get(outFormat)
	if !ok {
		writeError(w, http.StatusBadRequest, "unknown output format: "+outFormat, "")
		return
	}

	ws, err := compiler.NewWorkspace()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "workspace setup failed", err.Error())
		return
	}
	defer ws.Close()

	if err := populate(ws); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "source preparation failed", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), compilationTimeout())
	defer cancel()

	_, err = compiler.Compile(ctx, ws, compileOpts, entry)
	if err != nil {
		if ctx.Err() != nil {
			writeError(w, http.StatusGatewayTimeout, "compilation timed out", err.Error())
		} else {
			writeError(w, http.StatusUnprocessableEntity, "compilation failed", err.Error())
		}
		return
	}

	if err := formatter.Format(ws.OutDir, entry, outputOpts, w); err != nil {
		// Headers may already be partially written; log but don't double-write.
		_ = err
	}
}

type errorResponse struct {
	Error string `json:"error"`
	Log   string `json:"log,omitempty"`
}

func writeError(w http.ResponseWriter, code int, msg, log string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: msg, Log: log})
}
