package domain

import (
	"os"
	"strings"
	"testing"
)

func TestSecretReferenceWireContractHasNoValueField(t *testing.T) {
	t.Parallel()
	contents, err := os.ReadFile("../../../proto/tessera/management/v1/secret_reference.proto")
	if err != nil {
		t.Fatal(err)
	}
	text := string(contents)
	for _, required := range []string{
		"message SecretReference",
		"string reference_id",
		"SecretPurpose purpose",
		"string provider_reference",
		"SecretCustodyState state",
		"string custody_audit_id",
	} {
		if !strings.Contains(text, required) {
			t.Errorf("wire contract missing %q", required)
		}
	}
	for _, forbidden := range []string{"string value", "bytes value", "password", "private_key", "credential", "map<"} {
		if strings.Contains(strings.ToLower(text), forbidden) {
			t.Errorf("wire contract contains forbidden value-bearing shape %q", forbidden)
		}
	}
}
