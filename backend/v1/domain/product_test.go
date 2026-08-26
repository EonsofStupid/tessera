package domain

import "testing"

func TestProductVersionIsFrozenAlpha(t *testing.T) {
	t.Parallel()
	if ProductVersion != "1.0.0-alpha" {
		t.Fatalf("product version is frozen at 1.0.0-alpha until Vault, Mesh, and G2 ship; got %q", ProductVersion)
	}
}
