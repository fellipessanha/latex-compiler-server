package handlers

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"latex-compiler-api/internal/compiler"
	"latex-compiler-api/internal/output"
)

type gitRequest struct {
	URL         string                  `json:"url"`
	Token       string                  `json:"token"`
	Ref         string                  `json:"ref"`
	Entry       string                  `json:"entry"`
	TargetDir   string                  `json:"target_dir"`
	CompileOpts compiler.CompileOptions `json:"compile_options"`
	OutputOpts  output.OutputOptions    `json:"output_options"`
}

// Git handles POST /git and POST /git/{output}.
//
// @Summary   Compile LaTeX from a git repository
// @Tags      compile
// @Accept    json
// @Produce   application/pdf,image/png,application/zip,application/x-tar,application/x-7z-compressed
// @Param     output  path    string      false  "Output format (pdf, png, zip, tar, 7zip). Defaults to pdf."  Enums(pdf,png,zip,tar,7zip)
// @Param     body    body    gitRequest  true   "Repository URL and compile options"
// @Success   200
// @Failure   400  {object}  errorResponse
// @Failure   401  {object}  errorResponse
// @Failure   422  {object}  errorResponse
// @Failure   504  {object}  errorResponse
// @Security  BearerAuth
// @Router    /git/{output} [post]
// @Router    /git [post]
func Git(w http.ResponseWriter, r *http.Request) {
	var req gitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body", err.Error())
		return
	}
	if req.URL == "" {
		writeError(w, http.StatusBadRequest, "url is required", "")
		return
	}
	if req.TargetDir != "" {
		if err := validateTargetDir(req.TargetDir); err != nil {
			writeError(w, http.StatusBadRequest, err.Error(), "")
			return
		}
	}

	entry := req.Entry
	if entry == "" {
		entry = "main.tex"
	}

	outFormat := resolveOutput(r)

	compileAndFormat(w, r, outFormat, entry, req.CompileOpts, req.OutputOpts,
		func(ws compiler.Workspace) error {
			return cloneRepo(req.URL, req.Token, req.Ref, req.TargetDir, ws.InDir)
		},
	)
}

// validateTargetDir rejects absolute paths and any path that would escape the
// repository root via ".." components.
func validateTargetDir(targetDir string) error {
	if filepath.IsAbs(targetDir) {
		return fmt.Errorf("target_dir must be a relative path, got %q", targetDir)
	}
	clean := filepath.Clean(targetDir)
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf("target_dir must not escape the repository root")
	}
	return nil
}

// cloneRepo performs a bare clone of url and extracts the result into destDir
// via git archive. When targetDir is non-empty only that subdirectory is
// extracted; otherwise the full tree is extracted.
func cloneRepo(rawURL, token, ref, targetDir, destDir string) error {
	cloneURL := rawURL
	if token != "" {
		cloneURL = embedToken(rawURL, token)
	}

	isCommitSHA := len(ref) == 40 && isHex(ref)

	tmpDir, err := os.MkdirTemp("", "latex-bare-*")
	if err != nil {
		return fmt.Errorf("create bare clone dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	cloneArgs := []string{"clone", "--bare", "--filter=blob:none", "--single-branch"}
	if !isCommitSHA {
		cloneArgs = append(cloneArgs, "--depth", "1")
		if ref != "" {
			cloneArgs = append(cloneArgs, "--branch", ref)
		}
	}
	cloneArgs = append(cloneArgs, cloneURL, tmpDir)

	if out, err := exec.Command("git", cloneArgs...).CombinedOutput(); err != nil {
		return fmt.Errorf("git clone (bare): %w\n%s", err, out)
	}

	archiveRef := "HEAD"
	if isCommitSHA {
		if out, err := exec.Command("git", "-C", tmpDir, "fetch", "origin", ref).CombinedOutput(); err != nil {
			return fmt.Errorf("git fetch %s: %w\n%s", ref, err, out)
		}
		archiveRef = ref
	}

	// When targetDir is set: "HEAD:<targetDir>" archives only that subtree.
	// When empty:            "HEAD"             archives the full tree.
	archiveArg := archiveRef
	if targetDir != "" {
		archiveArg = archiveRef + ":" + targetDir
	}

	archiveCmd := exec.Command("git", "-C", tmpDir, "archive", archiveArg)
	stdout, err := archiveCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("git archive pipe: %w", err)
	}
	var archiveStderr bytes.Buffer
	archiveCmd.Stderr = &archiveStderr

	if err := archiveCmd.Start(); err != nil {
		return fmt.Errorf("git archive start: %w", err)
	}

	if err := extractTarStream(stdout, destDir); err != nil {
		_ = archiveCmd.Wait()
		return fmt.Errorf("git archive extract: %w", err)
	}
	if err := archiveCmd.Wait(); err != nil {
		return fmt.Errorf("git archive %q: %w\n%s", archiveArg, err, archiveStderr.String())
	}
	return nil
}

// extractTarStream reads a tar stream from r and writes all files into destDir.
func extractTarStream(r io.Reader, destDir string) error {
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
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
		f, err := os.Create(dest)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(f, tr)
		f.Close()
		if copyErr != nil {
			return copyErr
		}
	}
}

func embedToken(rawURL, token string) string {
	for _, scheme := range []string{"https://", "http://"} {
		if strings.HasPrefix(rawURL, scheme) {
			return scheme + token + "@" + rawURL[len(scheme):]
		}
	}
	return rawURL
}

func isHex(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
