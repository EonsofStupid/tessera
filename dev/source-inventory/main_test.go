package main

import (
	"strings"
	"testing"
)

func TestMatches(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		pattern string
		path    string
		want    bool
	}{
		{name: "recursive child", pattern: "internal/**", path: "internal/api/flows/flows.go", want: true},
		{name: "recursive root", pattern: "internal/**", path: "internal", want: true},
		{name: "recursive sibling", pattern: "internal/**", path: "internalize/file.go", want: false},
		{name: "file wildcard", pattern: "backend/v1/domain/flow*.go", path: "backend/v1/domain/flow_executor.go", want: true},
		{name: "file wildcard no slash crossing", pattern: "backend/v1/domain/flow*.go", path: "backend/v1/domain/flow/test.go", want: false},
		{name: "exact", pattern: "main.go", path: "main.go", want: true},
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

func TestClassifier(t *testing.T) {
	t.Parallel()
	c := classifier{
		manifest: manifest{
			BaselineImport: baselineImport{
				SourceID:     "core",
				Relationship: "derived",
				Overrides:    []baselineOverride{{Pattern: "proto/**", SourceID: "proto", Relationship: "derived"}},
			},
			PathRules: []pathRule{
				{ID: "native-domain", Patterns: []string{"backend/v1/**"}, SourceID: "native", Relationship: "native"},
			},
			ReviewedPathPatterns:  []string{"backend/**", "internal/**", "proto/**"},
			DefaultClassification: classification{SourceID: "native", Relationship: "native", RuleID: "default"},
		},
		imported: map[string]struct{}{
			"internal/core.go": {},
			"proto/api.proto":  {},
		},
	}

	tests := []struct {
		name       string
		path       string
		wantSource string
		wantRule   string
		wantError  string
	}{
		{name: "path rule precedes baseline", path: "backend/v1/seat.go", wantSource: "native", wantRule: "native-domain"},
		{name: "baseline core", path: "internal/core.go", wantSource: "core", wantRule: "baseline-import"},
		{name: "baseline override", path: "proto/api.proto", wantSource: "proto", wantRule: "baseline-import-override"},
		{name: "default native", path: "docs/contract.md", wantSource: "native", wantRule: "default"},
		{name: "unmapped reviewed", path: "internal/new.go", wantError: "has no explicit rule"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := c.classify(test.path)
			if test.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantError) {
					t.Fatalf("classify(%q) error = %v, want containing %q", test.path, err, test.wantError)
				}
				return
			}
			if err != nil {
				t.Fatalf("classify(%q): %v", test.path, err)
			}
			if got.SourceID != test.wantSource || got.RuleID != test.wantRule {
				t.Fatalf("classify(%q) = source %q rule %q, want source %q rule %q", test.path, got.SourceID, got.RuleID, test.wantSource, test.wantRule)
			}
		})
	}
}

func TestValidateManifestRejectsUnknownSource(t *testing.T) {
	t.Parallel()
	m := manifest{
		Schema: manifestSchema,
		BaselineImport: baselineImport{
			Commit:       "commit",
			SourceID:     "missing",
			Relationship: "derived",
		},
		Sources: []source{{
			ID:                "native",
			Repository:        "https://example.test/native",
			Revision:          "SELF",
			LicenseExpression: "LicenseRef-Pending",
		}},
		DefaultClassification: classification{SourceID: "native", Relationship: "native", RuleID: "default"},
	}
	if err := validateManifest(m); err == nil || !strings.Contains(err.Error(), "unknown source") {
		t.Fatalf("validateManifest() error = %v, want unknown source", err)
	}
}

func TestBuildBOM(t *testing.T) {
	t.Parallel()
	sources := []source{
		{ID: "native", Repository: "https://example.test/native", Revision: "SELF", LicenseExpression: "LicenseRef-Pending"},
		{ID: "core", Repository: "https://example.test/core", Revision: "abc", LicenseExpression: "AGPL-3.0-only"},
	}
	c := classifier{
		manifest: manifest{
			BaselineImport: baselineImport{
				Commit:       "baseline",
				SourceID:     "core",
				Relationship: "derived",
			},
			Sources:               sources,
			DefaultClassification: classification{SourceID: "native", Relationship: "native", RuleID: "default"},
		},
		sources: map[string]source{
			"native": sources[0],
			"core":   sources[1],
		},
		imported: map[string]struct{}{"internal/core.go": {}},
	}

	got, err := buildBOM(c, []string{"docs/contract.md", "internal/core.go"}, "provenance/source-manifest.json")
	if err != nil {
		t.Fatalf("buildBOM(): %v", err)
	}
	if got.Schema != bomSchema || got.PathCount != 2 || len(got.ClassificationSHA256) != 64 {
		t.Fatalf("buildBOM() metadata = schema %q paths %d digest %q", got.Schema, got.PathCount, got.ClassificationSHA256)
	}
	if len(got.Summary) != 2 || got.Summary[0] != (sourceSummary{SourceID: "core", Paths: 1}) || got.Summary[1] != (sourceSummary{SourceID: "native", Paths: 1}) {
		t.Fatalf("buildBOM() summary = %#v", got.Summary)
	}
	if len(got.Groups) != 2 || got.Groups[0].SourceID != "core" || got.Groups[1].SourceID != "native" {
		t.Fatalf("buildBOM() groups = %#v", got.Groups)
	}
}

func TestValidateManifestAcceptsCompleteManifest(t *testing.T) {
	t.Parallel()
	m := manifest{
		Schema: manifestSchema,
		BaselineImport: baselineImport{
			Commit:       "commit",
			SourceID:     "core",
			Relationship: "derived",
			Overrides:    []baselineOverride{{Pattern: "proto/**", SourceID: "proto", Relationship: "derived"}},
		},
		Sources: []source{
			{ID: "native", Repository: "https://example.test/native", Revision: "SELF", LicenseExpression: "LicenseRef-Pending"},
			{ID: "core", Repository: "https://example.test/core", Revision: "abc", LicenseExpression: "AGPL-3.0-only"},
			{ID: "proto", Repository: "https://example.test/core", Revision: "abc", LicenseExpression: "Apache-2.0"},
			{ID: "ideas", Repository: "https://example.test/ideas", Revision: "def", LicenseExpression: "MIT"},
		},
		PathRules: []pathRule{{
			ID:           "native-rule",
			Patterns:     []string{"backend/v1/**"},
			SourceID:     "native",
			Relationship: "native",
			InspiredBy:   []string{"ideas"},
			Reason:       "test rule",
		}},
		DefaultClassification: classification{SourceID: "native", Relationship: "native", RuleID: "default"},
	}
	if err := validateManifest(m); err != nil {
		t.Fatalf("validateManifest(): %v", err)
	}
}

func TestSplitNUL(t *testing.T) {
	t.Parallel()
	got := splitNUL("one\x00two\x00")
	if len(got) != 2 || got[0] != "one" || got[1] != "two" {
		t.Fatalf("splitNUL() = %#v", got)
	}
}

func TestNormalizeLineEndings(t *testing.T) {
	t.Parallel()
	want := []byte("one\ntwo\n")
	for _, input := range [][]byte{want, []byte("one\r\ntwo\r\n")} {
		if got := normalizeLineEndings(input); string(got) != string(want) {
			t.Fatalf("normalizeLineEndings() = %q, want %q", got, want)
		}
	}
}
