package output

import "net/http"

// OutputOptions controls post-compilation rendering behaviour.
type OutputOptions struct {
	// IncludePNG adds a page-1 thumbnail (thumbnail.png) to archive output formats.
	// Ignored for pdf and png output types.
	IncludePNG bool `json:"include_png"`
}

// Formatter writes compiled output to an HTTP response.
// Implement this interface to add a new output format.
type Formatter interface {
	Format(outDir string, entry string, opts OutputOptions, w http.ResponseWriter) error
}

// registry maps the {output} path segment to the corresponding Formatter.
// Adding a new format requires only a new file in this package and one line here.
var registry = map[string]Formatter{
	"pdf":  PDFFormatter{},
	"png":  PNGFormatter{},
	"zip":  ArchiveFormatter{Kind: "zip"},
	"tar":  ArchiveFormatter{Kind: "tar"},
	"7zip": ArchiveFormatter{Kind: "7zip"},
}

// Get returns the Formatter for the given output name, or false if unknown.
func Get(output string) (Formatter, bool) {
	f, ok := registry[output]
	return f, ok
}
