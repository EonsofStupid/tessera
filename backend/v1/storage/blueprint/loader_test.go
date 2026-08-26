package blueprint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

const valid = `schema: nomen.blueprint.v1
name: %s
entries:
  - model: nomen/seat
    identifiers: {member: mem_1}
    attrs: {account: acc_1}
`

func TestLoad_LexicographicOrderIsTheSequence(t *testing.T) {
	dir := t.TempDir()
	// Written out of order; the names are the plan.
	write(t, dir, "20-seats.yaml", strings.ReplaceAll(valid, "%s", "second"))
	write(t, dir, "10-accounts.yml", strings.ReplaceAll(valid, "%s", "first"))
	write(t, dir, "README.md", "not a blueprint")

	files, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("loaded %d files, want the 2 yaml ones", len(files))
	}
	if files[0].Blueprint.Name != "first" || files[1].Blueprint.Name != "second" {
		t.Fatalf("order = %s, %s", files[0].Blueprint.Name, files[1].Blueprint.Name)
	}
}

func TestLoad_EmptyDirIsAWrongDirNotSuccess(t *testing.T) {
	_, err := Load(t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "no *.yaml files") {
		t.Fatalf("err = %v — succeeding at nothing reports converged about a mistake", err)
	}
}

func TestLoad_ErrorsNameTheFile(t *testing.T) {
	dir := t.TempDir()
	// A typo'd top level would load as an empty blueprint without strictness.
	write(t, dir, "a.yaml", "schema: nomen.blueprint.v1\nentrys: []\n")
	// And a structurally invalid one fails Validate.
	write(t, dir, "b.yaml", "schema: wrong\nentries: []\n")

	_, err := Load(dir)
	if err == nil {
		t.Fatal("accepted")
	}
	for _, frag := range []string{"a.yaml", `unknown field "entrys"`, "b.yaml", `schema is "wrong"`} {
		if !strings.Contains(err.Error(), frag) {
			t.Errorf("error lacks %q:\n%v", frag, err)
		}
	}
}
