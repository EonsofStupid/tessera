package conformance_test

import (
	"encoding/json"
	"os"
	"sort"
	"strings"
	"testing"

	capabilitycontract "github.com/EonsofStupid/tessera/contracts/capabilities"
)

type registry struct {
	SchemaVersion string  `json:"schemaVersion"`
	Suites        []suite `json:"suites"`
}

type suite struct {
	CapabilityID  string     `json:"capabilityId"`
	ConformanceID string     `json:"conformanceId"`
	Status        string     `json:"status"`
	Profiles      []string   `json:"profiles"`
	Cases         []testCase `json:"cases"`
}

type testCase struct {
	ID          string `json:"id"`
	Class       string `json:"class"`
	Profile     string `json:"profile"`
	Requirement string `json:"requirement"`
}

type workspaceManifest struct {
	Capabilities []struct {
		ID            string   `json:"id"`
		ConformanceID string   `json:"conformanceId"`
		Profiles      []string `json:"profiles"`
	} `json:"capabilities"`
}

func TestSuiteRegistryCoversMandatoryCapabilities(t *testing.T) {
	value := loadRegistry(t)
	want := capabilitycontract.Mandatory()
	got := make([]string, 0, len(value.Suites))
	for _, item := range value.Suites {
		got = append(got, item.CapabilityID)
	}
	sort.Strings(got)
	sort.Strings(want)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("suite capabilities = %v, want %v", got, want)
	}
}

func TestEverySuiteHasAdversarialAndFailureCoverage(t *testing.T) {
	value := loadRegistry(t)
	for _, item := range value.Suites {
		t.Run(item.CapabilityID, func(t *testing.T) {
			if item.Status != "planned" && item.Status != "implemented" && item.Status != "verified" {
				t.Fatalf("unknown status %q", item.Status)
			}
			if item.ConformanceID == "" || len(item.Profiles) == 0 || len(item.Cases) < 4 {
				t.Fatal("suite is incomplete")
			}
			classes := make(map[string]bool)
			caseIDs := make(map[string]struct{})
			for _, test := range item.Cases {
				if test.ID == "" || test.Profile == "" || len(test.Requirement) < 20 {
					t.Fatalf("incomplete case: %#v", test)
				}
				if _, duplicate := caseIDs[test.ID]; duplicate {
					t.Fatalf("duplicate case id %q", test.ID)
				}
				caseIDs[test.ID] = struct{}{}
				classes[test.Class] = true
			}
			for _, required := range []string{"positive", "negative", "failure", "recovery"} {
				if !classes[required] {
					t.Errorf("missing %s coverage", required)
				}
			}
		})
	}
}

func TestVisualFlowSuiteIncludesAccessibility(t *testing.T) {
	value := loadRegistry(t)
	for _, item := range value.Suites {
		if item.CapabilityID != capabilitycontract.VisualFlowEngine {
			continue
		}
		for _, test := range item.Cases {
			if test.Class == "accessibility" {
				return
			}
		}
	}
	t.Fatal("visual flow suite has no accessibility case")
}

func TestSuitesMatchWorkspaceConformanceContracts(t *testing.T) {
	registry := loadRegistry(t)
	data, err := os.ReadFile("../dev/workspace/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var workspace workspaceManifest
	if err := json.Unmarshal(data, &workspace); err != nil {
		t.Fatal(err)
	}
	contracts := make(map[string]suite, len(registry.Suites))
	for _, item := range registry.Suites {
		contracts[item.CapabilityID] = item
	}
	for _, capability := range workspace.Capabilities {
		item, ok := contracts[capability.ID]
		if !ok {
			t.Errorf("workspace capability %s has no suite", capability.ID)
			continue
		}
		if item.ConformanceID != capability.ConformanceID || strings.Join(item.Profiles, ",") != strings.Join(capability.Profiles, ",") {
			t.Errorf("suite %s differs from workspace contract", capability.ID)
		}
	}
}

func loadRegistry(t *testing.T) registry {
	t.Helper()
	data, err := os.ReadFile("suites.json")
	if err != nil {
		t.Fatal(err)
	}
	var value registry
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatal(err)
	}
	if value.SchemaVersion != "tessera.conformance-suites.v1" {
		t.Fatalf("schemaVersion = %q", value.SchemaVersion)
	}
	return value
}
