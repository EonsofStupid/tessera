package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

const manifestSchema = "tessera.source-provenance-manifest.v1"
const bomSchema = "tessera.source-bom.v1"

type source struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Repository        string `json:"repository"`
	Revision          string `json:"revision"`
	RevisionStatus    string `json:"revision_status"`
	LicenseExpression string `json:"license_expression"`
	LicenseReference  string `json:"license_reference"`
}

type classification struct {
	SourceID     string   `json:"source_id"`
	Relationship string   `json:"relationship"`
	RuleID       string   `json:"rule_id,omitempty"`
	InspiredBy   []string `json:"inspired_by,omitempty"`
}

type pathRule struct {
	ID           string   `json:"id"`
	Patterns     []string `json:"patterns"`
	SourceID     string   `json:"source_id"`
	Relationship string   `json:"relationship"`
	InspiredBy   []string `json:"inspired_by,omitempty"`
	Reason       string   `json:"reason"`
}

type baselineOverride struct {
	Pattern      string `json:"pattern"`
	SourceID     string `json:"source_id"`
	Relationship string `json:"relationship"`
}

type baselineImport struct {
	Commit       string             `json:"commit"`
	SourceID     string             `json:"source_id"`
	Relationship string             `json:"relationship"`
	Overrides    []baselineOverride `json:"overrides"`
}

type manifest struct {
	Schema                string         `json:"schema"`
	BaselineImport        baselineImport `json:"baseline_import"`
	Sources               []source       `json:"sources"`
	PathRules             []pathRule     `json:"path_rules"`
	ReviewedPathPatterns  []string       `json:"reviewed_path_patterns"`
	DefaultClassification classification `json:"default_classification"`
}

type bomGroup struct {
	RuleID            string   `json:"rule_id"`
	SourceID          string   `json:"source_id"`
	Relationship      string   `json:"relationship"`
	LicenseExpression string   `json:"license_expression"`
	InspiredBy        []string `json:"inspired_by,omitempty"`
	Paths             []string `json:"paths"`
}

type sourceSummary struct {
	SourceID string `json:"source_id"`
	Paths    int    `json:"paths"`
}

type bom struct {
	Schema               string          `json:"schema"`
	Manifest             string          `json:"manifest"`
	BaselineImportCommit string          `json:"baseline_import_commit"`
	PathCount            int             `json:"path_count"`
	ClassificationSHA256 string          `json:"classification_sha256"`
	Sources              []source        `json:"sources"`
	Summary              []sourceSummary `json:"summary"`
	Groups               []bomGroup      `json:"groups"`
}

type classifier struct {
	manifest manifest
	sources  map[string]source
	imported map[string]struct{}
}

func main() {
	var manifestPath string
	var outputPath string
	var check bool
	flag.StringVar(&manifestPath, "manifest", "provenance/source-manifest.json", "source manifest path relative to the repository root")
	flag.StringVar(&outputPath, "output", "provenance/source-bom.json", "generated BOM path relative to the repository root")
	flag.BoolVar(&check, "check", false, "verify that the checked-in BOM is current")
	flag.Parse()

	if err := run(manifestPath, outputPath, check); err != nil {
		fmt.Fprintln(os.Stderr, "source inventory:", err)
		os.Exit(1)
	}
}

func run(manifestPath, outputPath string, check bool) error {
	root, err := gitOutput("rev-parse", "--show-toplevel")
	if err != nil {
		return err
	}
	root = strings.TrimSpace(root)
	manifestFile := filepath.Join(root, filepath.FromSlash(manifestPath))
	outputFile := filepath.Join(root, filepath.FromSlash(outputPath))

	manifestBytes, err := os.ReadFile(manifestFile)
	if err != nil {
		return fmt.Errorf("read manifest: %w", err)
	}
	var m manifest
	if err := json.Unmarshal(manifestBytes, &m); err != nil {
		return fmt.Errorf("decode manifest: %w", err)
	}
	if err := validateManifest(m); err != nil {
		return err
	}

	importedOutput, err := gitOutput("diff-tree", "--no-commit-id", "--name-only", "--diff-filter=A", "-r", m.BaselineImport.Commit)
	if err != nil {
		return fmt.Errorf("read baseline import %s (full Git history is required): %w", m.BaselineImport.Commit, err)
	}
	imported := make(map[string]struct{})
	for _, name := range strings.Fields(importedOutput) {
		imported[name] = struct{}{}
	}

	trackedOutput, err := gitOutput("ls-files", "-z")
	if err != nil {
		return err
	}
	tracked := splitNUL(trackedOutput)
	sort.Strings(tracked)

	sourceByID := make(map[string]source, len(m.Sources))
	for _, item := range m.Sources {
		sourceByID[item.ID] = item
	}
	c := classifier{manifest: m, sources: sourceByID, imported: imported}
	generated, err := buildBOM(c, tracked, manifestPath)
	if err != nil {
		return err
	}
	wanted, err := json.MarshalIndent(generated, "", "  ")
	if err != nil {
		return fmt.Errorf("encode BOM: %w", err)
	}
	wanted = append(wanted, '\n')

	if check {
		actual, err := os.ReadFile(outputFile)
		if err != nil {
			return fmt.Errorf("read checked-in BOM: %w", err)
		}
		if !bytes.Equal(normalizeLineEndings(actual), wanted) {
			return errors.New("source-bom.json is stale; run `go run ./dev/source-inventory` and commit the result")
		}
		fmt.Printf("source provenance verified: %d tracked paths, sha256 %s\n", generated.PathCount, generated.ClassificationSHA256)
		return nil
	}

	if err := os.WriteFile(outputFile, wanted, 0o644); err != nil {
		return fmt.Errorf("write BOM: %w", err)
	}
	fmt.Printf("wrote %s: %d tracked paths, sha256 %s\n", outputPath, generated.PathCount, generated.ClassificationSHA256)
	return nil
}

func normalizeLineEndings(value []byte) []byte {
	return bytes.ReplaceAll(value, []byte("\r\n"), []byte("\n"))
}

func validateManifest(m manifest) error {
	if m.Schema != manifestSchema {
		return fmt.Errorf("unsupported manifest schema %q", m.Schema)
	}
	if m.BaselineImport.Commit == "" || m.BaselineImport.SourceID == "" || m.BaselineImport.Relationship == "" {
		return errors.New("baseline import commit, source_id and relationship are required")
	}
	if len(m.Sources) == 0 {
		return errors.New("at least one source is required")
	}
	sources := make(map[string]struct{}, len(m.Sources))
	for _, item := range m.Sources {
		if item.ID == "" || item.Repository == "" || item.Revision == "" || item.LicenseExpression == "" {
			return fmt.Errorf("source %q is incomplete", item.ID)
		}
		if _, exists := sources[item.ID]; exists {
			return fmt.Errorf("duplicate source id %q", item.ID)
		}
		sources[item.ID] = struct{}{}
	}
	checkSource := func(id, context string) error {
		if _, ok := sources[id]; !ok {
			return fmt.Errorf("%s references unknown source %q", context, id)
		}
		return nil
	}
	if err := checkSource(m.BaselineImport.SourceID, "baseline import"); err != nil {
		return err
	}
	for _, override := range m.BaselineImport.Overrides {
		if override.Pattern == "" || override.Relationship == "" {
			return errors.New("baseline override pattern and relationship are required")
		}
		if err := checkSource(override.SourceID, "baseline override"); err != nil {
			return err
		}
	}
	ruleIDs := make(map[string]struct{}, len(m.PathRules))
	for _, rule := range m.PathRules {
		if rule.ID == "" || len(rule.Patterns) == 0 || rule.Relationship == "" || rule.Reason == "" {
			return fmt.Errorf("path rule %q is incomplete", rule.ID)
		}
		if _, exists := ruleIDs[rule.ID]; exists {
			return fmt.Errorf("duplicate path rule id %q", rule.ID)
		}
		ruleIDs[rule.ID] = struct{}{}
		if err := checkSource(rule.SourceID, "path rule "+rule.ID); err != nil {
			return err
		}
		for _, inspiredBy := range rule.InspiredBy {
			if err := checkSource(inspiredBy, "path rule "+rule.ID); err != nil {
				return err
			}
		}
	}
	if m.DefaultClassification.RuleID == "" || m.DefaultClassification.Relationship == "" {
		return errors.New("default classification rule_id and relationship are required")
	}
	return checkSource(m.DefaultClassification.SourceID, "default classification")
}

func buildBOM(c classifier, tracked []string, manifestPath string) (bom, error) {
	type groupKey struct {
		ruleID       string
		sourceID     string
		relationship string
		inspiredBy   string
	}
	groups := make(map[groupKey][]string)
	summary := make(map[string]int)
	hash := sha256.New()

	for _, name := range tracked {
		result, err := c.classify(name)
		if err != nil {
			return bom{}, err
		}
		inspired := append([]string(nil), result.InspiredBy...)
		sort.Strings(inspired)
		key := groupKey{
			ruleID:       result.RuleID,
			sourceID:     result.SourceID,
			relationship: result.Relationship,
			inspiredBy:   strings.Join(inspired, ","),
		}
		groups[key] = append(groups[key], name)
		summary[result.SourceID]++
		license := c.sources[result.SourceID].LicenseExpression
		fmt.Fprintf(hash, "%s\x00%s\x00%s\x00%s\x00%s\x00%s\n", name, result.SourceID, result.Relationship, result.RuleID, license, key.inspiredBy)
	}

	keys := make([]groupKey, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		left := keys[i].sourceID + "\x00" + keys[i].relationship + "\x00" + keys[i].ruleID + "\x00" + keys[i].inspiredBy
		right := keys[j].sourceID + "\x00" + keys[j].relationship + "\x00" + keys[j].ruleID + "\x00" + keys[j].inspiredBy
		return left < right
	})
	bomGroups := make([]bomGroup, 0, len(keys))
	for _, key := range keys {
		paths := groups[key]
		sort.Strings(paths)
		var inspiredBy []string
		if key.inspiredBy != "" {
			inspiredBy = strings.Split(key.inspiredBy, ",")
		}
		bomGroups = append(bomGroups, bomGroup{
			RuleID:            key.ruleID,
			SourceID:          key.sourceID,
			Relationship:      key.relationship,
			LicenseExpression: c.sources[key.sourceID].LicenseExpression,
			InspiredBy:        inspiredBy,
			Paths:             paths,
		})
	}

	sourceIDs := make([]string, 0, len(summary))
	for sourceID := range summary {
		sourceIDs = append(sourceIDs, sourceID)
	}
	sort.Strings(sourceIDs)
	summaries := make([]sourceSummary, 0, len(sourceIDs))
	for _, sourceID := range sourceIDs {
		summaries = append(summaries, sourceSummary{SourceID: sourceID, Paths: summary[sourceID]})
	}

	return bom{
		Schema:               bomSchema,
		Manifest:             filepath.ToSlash(manifestPath),
		BaselineImportCommit: c.manifest.BaselineImport.Commit,
		PathCount:            len(tracked),
		ClassificationSHA256: hex.EncodeToString(hash.Sum(nil)),
		Sources:              append([]source(nil), c.manifest.Sources...),
		Summary:              summaries,
		Groups:               bomGroups,
	}, nil
}

func (c classifier) classify(name string) (classification, error) {
	for _, rule := range c.manifest.PathRules {
		for _, pattern := range rule.Patterns {
			if matches(pattern, name) {
				return classification{
					SourceID:     rule.SourceID,
					Relationship: rule.Relationship,
					RuleID:       rule.ID,
					InspiredBy:   append([]string(nil), rule.InspiredBy...),
				}, nil
			}
		}
	}
	if _, ok := c.imported[name]; ok {
		for _, override := range c.manifest.BaselineImport.Overrides {
			if matches(override.Pattern, name) {
				return classification{
					SourceID:     override.SourceID,
					Relationship: override.Relationship,
					RuleID:       "baseline-import-override",
				}, nil
			}
		}
		return classification{
			SourceID:     c.manifest.BaselineImport.SourceID,
			Relationship: c.manifest.BaselineImport.Relationship,
			RuleID:       "baseline-import",
		}, nil
	}
	for _, pattern := range c.manifest.ReviewedPathPatterns {
		if matches(pattern, name) {
			return classification{}, fmt.Errorf("tracked path %q is under provenance review but has no explicit rule", name)
		}
	}
	return c.manifest.DefaultClassification, nil
}

func matches(pattern, name string) bool {
	if strings.HasSuffix(pattern, "/**") {
		prefix := strings.TrimSuffix(pattern, "/**")
		return name == prefix || strings.HasPrefix(name, prefix+"/")
	}
	matched, err := path.Match(pattern, name)
	return err == nil && matched
}

func splitNUL(value string) []string {
	parts := strings.Split(value, "\x00")
	if len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}

func gitOutput(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(string(exitError.Stderr)))
		}
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return string(output), nil
}
