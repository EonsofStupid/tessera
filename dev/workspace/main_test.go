package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

func TestManifestMatchesCapabilityContract(t *testing.T) {
	value, err := loadManifest("manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	got := make([]string, 0, len(value.Capabilities))
	for _, item := range value.Capabilities {
		got = append(got, item.ID)
	}
	want := expectedCapabilities()
	sort.Strings(got)
	sort.Strings(want)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("capabilities = %v, want %v", got, want)
	}
}

func TestInitStateGeneratesProtectedStableMasterKey(t *testing.T) {
	root := t.TempDir()
	created, err := initState(root, ".artifacts/workspace")
	if err != nil {
		t.Fatal(err)
	}
	if !created {
		t.Fatal("first init did not create the master key")
	}
	path := filepath.Join(root, ".artifacts", "workspace", "secrets", "tessera-masterkey")
	first, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != 32 {
		t.Fatalf("master key length = %d, want 32", len(first))
	}
	if runtime.GOOS != "windows" {
		stateInfo, err := os.Stat(filepath.Join(root, ".artifacts", "workspace"))
		if err != nil {
			t.Fatal(err)
		}
		if got := stateInfo.Mode().Perm(); got != 0o700 {
			t.Fatalf("workspace state mode = %04o, want 0700", got)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if got := info.Mode().Perm(); got != 0o600 {
			t.Fatalf("master key mode = %04o, want 0600", got)
		}
	}
	created, err = initState(root, ".artifacts/workspace")
	if err != nil {
		t.Fatal(err)
	}
	if created {
		t.Fatal("second init replaced the master key")
	}
	second, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("second init changed the master key")
	}
}

func TestReferenceMatchesSourceProvenance(t *testing.T) {
	value, err := loadManifest("manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join("..", "..", "provenance", "source-manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var provenance struct {
		Sources []struct {
			ID         string `json:"id"`
			Repository string `json:"repository"`
			Revision   string `json:"revision"`
		} `json:"sources"`
	}
	if err := json.Unmarshal(data, &provenance); err != nil {
		t.Fatal(err)
	}
	reference, ok := referenceByID(value.References, "authentik-concepts")
	if !ok {
		t.Fatal("authentik-concepts workspace reference is missing")
	}
	for _, source := range provenance.Sources {
		if source.ID == reference.ID {
			if strings.TrimSuffix(reference.Repository, ".git") != strings.TrimSuffix(source.Repository, ".git") || reference.Revision != source.Revision {
				t.Fatalf("workspace reference = %s@%s, provenance = %s@%s", reference.Repository, reference.Revision, source.Repository, source.Revision)
			}
			return
		}
	}
	t.Fatal("authentik-concepts provenance source is missing")
}

func TestValidateManifestRejectsUnsafeReferenceDestination(t *testing.T) {
	value, err := loadManifest("manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	value.References[0].Destination = "../authentik"
	if err := validateManifest(value); err == nil || !strings.Contains(err.Error(), "destination") {
		t.Fatalf("validateManifest error = %v, want unsafe destination refusal", err)
	}
}

func TestValidateManifestRejectsUnscopedState(t *testing.T) {
	value, err := loadManifest("manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	value.StateRoot = ".artifacts"
	if err := validateManifest(value); err == nil || !strings.Contains(err.Error(), "stateRoot") {
		t.Fatalf("validateManifest error = %v, want stateRoot refusal", err)
	}
}

func TestUnknownProfileFailsClosed(t *testing.T) {
	value, err := loadManifest("manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	report := doctor(t.TempDir(), value, "future")
	if report.Ready || len(report.Checks) != 1 || report.Checks[0].Status != "fail" {
		t.Fatalf("unexpected report: %#v", report)
	}
}

func TestRepositoryTracksNoForbiddenSecretShapedFiles(t *testing.T) {
	root, err := gitOutput("rev-parse", "--show-toplevel")
	if err != nil {
		t.Fatal(err)
	}
	if err := checkNoForbiddenTrackedSecretFiles(root); err != nil {
		t.Fatal(err)
	}
}
