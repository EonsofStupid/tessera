package build

import "testing"

func TestVersionDefaultsToFrozenAlpha(t *testing.T) {
	t.Parallel()
	if Version() != "1.0.0-alpha" {
		t.Fatalf("got %q", Version())
	}
}
