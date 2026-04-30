package output

import (
	"testing"
)

func TestGet_KnownFormats(t *testing.T) {
	cases := []struct {
		name string
		kind string // for ArchiveFormatter; empty for others
	}{
		{"pdf", ""},
		{"png", ""},
		{"zip", "zip"},
		{"tar", "tar"},
		{"7zip", "7zip"},
	}
	for _, tc := range cases {
		f, ok := Get(tc.name)
		if !ok {
			t.Errorf("Get(%q): expected true, got false", tc.name)
			continue
		}
		if f == nil {
			t.Errorf("Get(%q): returned nil formatter", tc.name)
			continue
		}
		if tc.kind != "" {
			af, ok := f.(ArchiveFormatter)
			if !ok {
				t.Errorf("Get(%q): expected ArchiveFormatter, got %T", tc.name, f)
				continue
			}
			if af.Kind != tc.kind {
				t.Errorf("Get(%q): Kind = %q, want %q", tc.name, af.Kind, tc.kind)
			}
		}
	}
}

func TestGet_UnknownFormat(t *testing.T) {
	if _, ok := Get("unknown"); ok {
		t.Error("expected false for unknown format, got true")
	}
	if _, ok := Get(""); ok {
		t.Error("expected false for empty string, got true")
	}
}
