package productlanguage

import (
	"io/fs"
	"os"
	"strings"
	"testing"
)

func TestTrackedProductFilesDoNotContainOperatorCredentials(t *testing.T) {
	t.Parallel()
	forbidden := []string{
		"jesse@nomen.sh",
		"founder.password",
		"NOMEN_FIRSTINSTANCE_ORG_HUMAN_PASSWORD=",
		"Password: Password1!",
	}
	err := fs.WalkDir(os.DirFS("../.."), ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			name := entry.Name()
			if name == ".git" || name == "node_modules" || name == "upstream" || name == ".claude" || name == ".artifacts" || name == ".pgdata" || name == ".chat" || name == "static" {
				return fs.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, "_test.go") || strings.Contains(path, "/e2e/") || strings.HasSuffix(path, "credentials_test.go") || strings.HasSuffix(path, ".jsonl") || strings.HasSuffix(path, "pretooluse.py") {
			return nil
		}
		data, err := os.ReadFile("../../" + path)
		if err != nil {
			return nil
		}
		text := strings.ToLower(string(data))
		for _, needle := range forbidden {
			if strings.Contains(text, strings.ToLower(needle)) && path != "dev/product-language/credentials_test.go" {
				t.Errorf("%s contains hardcoded credential material %q", path, needle)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
