package console

import (
	"encoding/json"
	"testing"
)

func TestCreateEnvironmentJSONIncludesEdition(t *testing.T) {
	t.Parallel()
	raw, err := createEnvironmentJSON("https://nomen.example.test", "https://nomen.example.test", "client", "", "", "", "", false, "public", false)
	if err != nil {
		t.Fatal(err)
	}
	var env map[string]any
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatal(err)
	}
	if env["edition"] != "public" {
		t.Fatalf("edition %v", env["edition"])
	}
	if env["version"] != "1.0.0-alpha" {
		t.Fatalf("version %v", env["version"])
	}
	if _, ok := env["demo_caps"]; ok {
		t.Fatal("false demo_caps must omit")
	}
	raw, err = createEnvironmentJSON("", "", "", "", "", "", "", false, "", true)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatal(err)
	}
	if env["edition"] != "public" {
		t.Fatalf("default edition %v", env["edition"])
	}
	if env["demo_caps"] != true {
		t.Fatal("demo_caps")
	}
}
