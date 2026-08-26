package productlanguage

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProductLicenseIsAngryVibesAndShippin(t *testing.T) {
	t.Parallel()
	root := "../.."
	license, err := os.ReadFile(filepath.Join(root, "LICENSE"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(license)
	for _, need := range []string{"AngryVibes LLC", "shippin.ai", "absorbs no third-party product license"} {
		if !strings.Contains(text, need) {
			t.Errorf("LICENSE missing %q", need)
		}
	}
	for _, banned := range []string{"GNU Affero", "Apache License", "ZITADEL AG", "Authentik Security"} {
		if strings.Contains(text, banned) {
			t.Errorf("LICENSE absorbs third-party product license text %q", banned)
		}
	}
	notice, err := os.ReadFile(filepath.Join(root, "NOTICE"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(notice), "AngryVibes LLC") || !strings.Contains(string(notice), "shippin.ai") {
		t.Fatal("NOTICE must name AngryVibes LLC and shippin.ai")
	}
}

func TestRuntimeTreesDoNotShipThirdPartyProductLicenses(t *testing.T) {
	t.Parallel()
	err := fs.WalkDir(os.DirFS("../.."), ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			name := entry.Name()
			if name == ".git" || name == "node_modules" || name == "upstream" || name == ".artifacts" || name == ".pgdata" || name == ".chat" {
				return fs.SkipDir
			}
			return nil
		}
		base := strings.ToLower(entry.Name())
		if base != "license" && base != "licensing.md" && base != "copying" {
			return nil
		}
		if path == "LICENSE" || path == "NOTICE" {
			return nil
		}
		t.Errorf("%s ships a third-party product license; Nomen absorbs none (AngryVibes LLC and shippin.ai only)", path)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
