package productboundary_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve product-boundary test path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}

func readRepositoryFile(t *testing.T, root, name string) string {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join(root, name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(contents)
}

func normalizeMarkdown(contents string) string {
	return strings.Join(strings.Fields(contents), " ")
}

func TestStandaloneProductBoundaryIsAuthoritative(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	required := map[string][]string{
		"AGENTS.md": {
			"Tessera is a standalone identity and access management product",
			"docs/02-standalone-product-contract.md",
		},
		"README.md": {
			"Tessera is a standalone identity and access management platform",
			"Tessera product law lives in this repository",
		},
		"docs/00-charter.md": {
			"Tessera first; managed operation second; host-product integration third.",
		},
		"docs/02-standalone-product-contract.md": {
			"without installing or contacting Shippin",
			"The first managed release supports two explicit isolation profiles",
			"A transactional outbox projects redacted",
			"The Shippin adapter and embedded shell are deliberately outside this gate.",
			"semantic events and shared human/AI action grammar",
			"two control planes never become simultaneous sources of truth",
		},
		"docs/18-operator-interaction-contract.md": {
			"one operator model for humans, automation and dedicated AI specialists",
			"pixels and DOM recordings are never the source of truth",
			"AI may draft and explain. It cannot silently approve its own high-impact plan",
		},
		"docs/19-tenancy-and-authority-contract.md": {
			"Every new Tessera-owned PostgreSQL table",
			"Community mode uses PostgreSQL row-level security",
			"Tessera never merges two simultaneous writers by last-write-wins",
		},
		"web/package.json": {
			"@tessera/ui",
			"@tanstack/react-router",
		},
	}

	for name, phrases := range required {
		name, phrases := name, phrases
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			contents := normalizeMarkdown(readRepositoryFile(t, root, name))
			for _, phrase := range phrases {
				if !strings.Contains(contents, phrase) {
					t.Errorf("%s must contain %q", name, phrase)
				}
			}
		})
	}
}

func TestPrimaryProductDocsDoNotDeclareShippinOwnership(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	files := []string{
		"AGENTS.md",
		"README.md",
		"docs/00-charter.md",
		"docs/02-standalone-product-contract.md",
		"docs/03-roadmap.md",
		"docs/18-operator-interaction-contract.md",
		"docs/19-tenancy-and-authority-contract.md",
	}
	forbidden := []string{
		"Identity and authorization for the Shippin umbrella",
		"the token contract is the product boundary",
		"Product surface: Tessera inside the persistent Shippin member shell",
		"Tessera does not ship a second customer shell",
	}

	for _, name := range files {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			contents := normalizeMarkdown(readRepositoryFile(t, root, name))
			for _, phrase := range forbidden {
				if strings.Contains(contents, phrase) {
					t.Errorf("%s contains forbidden product-boundary declaration %q", name, phrase)
				}
			}
		})
	}
}
