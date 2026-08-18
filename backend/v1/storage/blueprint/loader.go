// Package blueprint reads blueprint files into domain types.
//
// The filesystem is an adapter like any other: this package knows about
// directories, extensions and YAML, and the domain knows none of it. Parsing
// goes through sigs.k8s.io/yaml — YAML read *as JSON* — which is the same
// decision the reference syntax made: a blueprint the panel emits as JSON and
// a blueprint a person writes as YAML are one format, not two.
package blueprint

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"sigs.k8s.io/yaml"

	"github.com/EonsofStupid/tessera/backend/v1/domain"
)

// File is one loaded blueprint and where it came from — every error a caller
// reports should name the file, because the fix is an edit to it.
type File struct {
	Path      string
	Blueprint *domain.Blueprint
}

// Load reads every *.yaml / *.yml directly in dir, in lexicographic order.
//
// Lexicographic, and not recursive, on purpose. Apply order is the files'
// order, so the names ARE the sequence — `10-accounts.yaml` before
// `20-seats.yaml` — and a glance at `ls` is a glance at the plan. Recursing
// would make order depend on walk semantics nobody reviews.
//
// Every file is validated here, all findings collected before any error
// returns: a directory with three broken files costs one round trip, the same
// contract Validate gives a single blueprint.
func Load(dir string) ([]File, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("blueprint: %w", err)
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		switch strings.ToLower(filepath.Ext(e.Name())) {
		case ".yaml", ".yml":
			names = append(names, e.Name())
		}
	}
	if len(names) == 0 {
		// An empty apply is more likely a wrong --dir than an empty estate,
		// and succeeding at nothing would report "converged" about a mistake.
		return nil, fmt.Errorf("blueprint: no *.yaml files in %s — is the directory right?", dir)
	}
	slices.Sort(names)

	var (
		files []File
		errs  []string
	)
	for _, name := range names {
		path := filepath.Join(dir, name)
		raw, err := os.ReadFile(path)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		var bp domain.Blueprint
		// Strict at the file level for the same reason attrs decode strictly:
		// a typo'd top-level `entrys:` would otherwise load as an empty
		// blueprint that applies clean and declares nothing.
		if err := yaml.UnmarshalStrict(raw, &bp); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", path, err))
			continue
		}
		if err := bp.Validate(); err != nil {
			errs = append(errs, fmt.Sprintf("%s:\n%v", path, err))
			continue
		}
		files = append(files, File{Path: path, Blueprint: &bp})
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf("blueprint:\n%s", strings.Join(errs, "\n"))
	}
	return files, nil
}
