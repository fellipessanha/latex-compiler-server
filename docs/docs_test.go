package docs_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

type swaggerDoc struct {
	Swagger string                     `json:"swagger"`
	Info    struct{ Title, Version string } `json:"info"`
	Paths   map[string]json.RawMessage `json:"paths"`
}

func loadSpec(t *testing.T) swaggerDoc {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine test file path")
	}
	path := filepath.Join(filepath.Dir(file), "swagger.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read swagger.json: %v", err)
	}
	var doc swaggerDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("parse swagger.json: %v", err)
	}
	return doc
}

func TestSpecIsValid(t *testing.T) {
	doc := loadSpec(t)
	if doc.Swagger != "2.0" {
		t.Errorf("expected swagger version 2.0, got %q", doc.Swagger)
	}
	if doc.Info.Title == "" {
		t.Error("spec missing info.title")
	}
	if doc.Info.Version == "" {
		t.Error("spec missing info.version")
	}
}

func TestSpecPathCoverage(t *testing.T) {
	doc := loadSpec(t)
	required := []string{"/git/{output}", "/git", "/blob/{output}", "/blob", "/health"}
	for _, path := range required {
		if _, ok := doc.Paths[path]; !ok {
			t.Errorf("spec missing path %q", path)
		}
	}
}
