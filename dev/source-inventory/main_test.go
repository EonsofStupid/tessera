package main

import (
	"strings"
	"testing"
)

func TestMatches(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, pattern, path string
		want                bool
	}{
		{"recursive child", "internal/**", "internal/api/flows/flows.go", true},
		{"recursive root", "internal/**", "internal", true},
		{"recursive sibling", "internal/**", "internalize/file.go", false},
		{"file wildcard", "backend/v1/domain/flow*.go", "backend/v1/domain/flow_executor.go", true},
		{"wildcard no slash crossing", "backend/v1/domain/flow*.go", "backend/v1/domain/flow/test.go", false},
		{"exact", "main.go", "main.go", true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := matches(test.pattern, test.path); got != test.want {
				t.Fatalf("matches(%q, %q) = %v, want %v", test.pattern, test.path, got, test.want)
			}
		})
	}
}

func TestClassifierFailsClosedForNewInheritedRootPath(t *testing.T) {
	t.Parallel()
	c := classifier{manifest: manifest{
		BaselineImport:        baselineImport{SourceID: "core", Relationship: "derived", Overrides: []baselineOverride{{Pattern: "proto/**", SourceID: "proto", Relationship: "derived"}}},
		PathRules:             []pathRule{{ID: "native-domain", Patterns: []string{"backend/v1/**"}, SourceID: "native", Relationship: "native"}},
		ReviewedPathPatterns:  []string{"backend/**", "internal/**", "proto/**"},
		DefaultClassification: classification{SourceID: "native", Relationship: "native", RuleID: "default"},
	}, imported: map[string]struct{}{"internal/core.go": {}, "proto/api.proto": {}}}
	tests := []struct{ path, source, rule, wantError string }{
		{"backend/v1/seat.go", "native", "native-domain", ""},
		{"internal/core.go", "core", "baseline-import", ""},
		{"proto/api.proto", "proto", "baseline-import-override", ""},
		{"docs/contract.md", "native", "default", ""},
		{"internal/new.go", "", "", "has no explicit rule"},
	}
	for _, test := range tests {
		got, err := c.classify(test.path)
		if test.wantError != "" {
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("classify(%q) error = %v", test.path, err)
			}
			continue
		}
		if err != nil || got.SourceID != test.source || got.RuleID != test.rule {
			t.Fatalf("classify(%q) = %#v, %v", test.path, got, err)
		}
	}
}

func TestValidateManifestAndBuildBOM(t *testing.T) {
	t.Parallel()
	sources := []source{
		{ID: "native", Repository: "https://example.test/native", Revision: "SELF", LicenseExpression: "LicenseRef-Pending", LicenseReference: "decision"},
		{ID: "core", Repository: "https://example.test/core", Revision: "abc", LicenseExpression: "AGPL-3.0-only", LicenseReference: "license"},
	}
	m := manifest{
		Schema:                manifestSchema,
		BaselineImport:        baselineImport{Commit: "baseline", SourceID: "core", Relationship: "derived"},
		Sources:               sources,
		DefaultClassification: classification{SourceID: "native", Relationship: "native", RuleID: "default"},
	}
	if err := validateManifest(m); err != nil {
		t.Fatal(err)
	}
	c := classifier{manifest: m, sources: map[string]source{"native": sources[0], "core": sources[1]}, imported: map[string]struct{}{"internal/core.go": {}}}
	got, err := buildBOM(c, []string{"docs/contract.md", "internal/core.go"}, "provenance/source-manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	if got.Schema != bomSchema || got.PathCount != 2 || len(got.ClassificationSHA256) != 64 || len(got.Summary) != 2 || len(got.Groups) != 2 {
		t.Fatalf("unexpected BOM: %#v", got)
	}
}

func TestValidateManifestRejectsAbsorbedThirdPartyLicenseOnNomenNative(t *testing.T) {
	t.Parallel()
	m := manifest{
		Schema:                manifestSchema,
		BaselineImport:        baselineImport{Commit: "commit", SourceID: "nomen-native", Relationship: "native"},
		Sources:               []source{{ID: "nomen-native", Repository: "https://example.test", Revision: "SELF", LicenseExpression: "AGPL-3.0-only", LicenseReference: "LICENSE"}},
		DefaultClassification: classification{SourceID: "nomen-native", Relationship: "native", RuleID: "default"},
	}
	if err := validateManifest(m); err == nil || !strings.Contains(err.Error(), "AngryVibes") {
		t.Fatalf("error = %v", err)
	}
}

func TestValidateManifestRejectsUnknownAndIncompleteSources(t *testing.T) {
	t.Parallel()
	m := manifest{
		Schema:                manifestSchema,
		BaselineImport:        baselineImport{Commit: "commit", SourceID: "missing", Relationship: "derived"},
		Sources:               []source{{ID: "native", Repository: "https://example.test", Revision: "SELF", LicenseExpression: "LicenseRef-Pending", LicenseReference: "decision"}},
		DefaultClassification: classification{SourceID: "native", Relationship: "native", RuleID: "default"},
	}
	if err := validateManifest(m); err == nil || !strings.Contains(err.Error(), "unknown source") {
		t.Fatalf("error = %v", err)
	}
	m.BaselineImport.SourceID = "native"
	m.Sources[0].LicenseReference = ""
	if err := validateManifest(m); err == nil || !strings.Contains(err.Error(), "incomplete") {
		t.Fatalf("error = %v", err)
	}
}

func TestNormalizationHelpers(t *testing.T) {
	t.Parallel()
	parts := splitNUL("one\x00two\x00")
	if len(parts) != 2 || parts[0] != "one" || parts[1] != "two" {
		t.Fatalf("splitNUL = %#v", parts)
	}
	if got := string(normalizeLineEndings([]byte("one\r\ntwo\r\n"))); got != "one\ntwo\n" {
		t.Fatalf("normalizeLineEndings = %q", got)
	}
}
