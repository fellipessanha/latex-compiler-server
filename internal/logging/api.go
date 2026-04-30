package logging

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// maxLogBody is the maximum number of request body bytes shown in logs.
const maxLogBody = 512

// responseRecorder wraps ResponseWriter to capture the status code and total bytes written.
type responseRecorder struct {
	http.ResponseWriter
	status int
	size   int
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.status = code
	rr.ResponseWriter.WriteHeader(code)
}

func (rr *responseRecorder) Write(b []byte) (int, error) {
	n, err := rr.ResponseWriter.Write(b)
	rr.size += n
	return n, err
}

// withLog logs incoming request details (endpoint, headers, body) and the
// outgoing response summary (headers, binary size).
func WithLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Peek at the first maxLogBody bytes without consuming the stream.
		var bodyPreview string
		if r.Body != nil {
			chunk, _ := io.ReadAll(io.LimitReader(r.Body, maxLogBody+1))
			truncated := len(chunk) > maxLogBody
			// Restore body: prepend the peeked chunk before the remaining stream.
			r.Body = io.NopCloser(io.MultiReader(bytes.NewReader(chunk), r.Body))
			preview := chunk
			if truncated {
				preview = chunk[:maxLogBody]
			}
			bodyPreview = string(preview)
			if truncated {
				bodyPreview += fmt.Sprintf("\n[body truncated — showing first %d bytes]", maxLogBody)
			}
		}

		var sb strings.Builder
		fmt.Fprintf(&sb, "→ %s %s\n", r.Method, r.URL.Path)
		fmt.Fprint(&sb, "  headers:\n")
		for k, vs := range r.Header {
			fmt.Fprintf(&sb, "    %s: %s\n", k, strings.Join(vs, ", "))
		}
		fmt.Fprintf(&sb, "  body: %s", bodyPreview)
		log.Print(sb.String())

		rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		var rsb strings.Builder
		fmt.Fprintf(&rsb, "← %s %s → %d (%d bytes)\n", r.Method, r.URL.Path, rec.status, rec.size)
		fmt.Fprint(&rsb, "  response headers:\n")
		for k, vs := range rec.Header() {
			fmt.Fprintf(&rsb, "    %s: %s\n", k, strings.Join(vs, ", "))
		}
		log.Print(rsb.String())
	})
}
