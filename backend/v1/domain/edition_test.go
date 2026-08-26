package domain

import "testing"

func TestNormalizeEdition(t *testing.T) {
	t.Parallel()
	if got := NormalizeEdition(""); got != EditionPublic {
		t.Fatalf("empty: %s", got)
	}
	if got := NormalizeEdition("ENTERPRISE"); got != EditionEnterprise {
		t.Fatalf("enterprise: %s", got)
	}
	if RedisAllowed(EditionPublic) {
		t.Fatal("public must refuse Redis")
	}
	if !RedisAllowed(EditionEnterprise) {
		t.Fatal("enterprise may enable Redis")
	}
}

func TestDemoCapsAllowFirstThenRefuse(t *testing.T) {
	t.Parallel()
	policy := EditionPolicy{Edition: EditionPublic, DemoCaps: true}
	if err := policy.AllowNewUser(24); err != nil {
		t.Fatalf("24 users still under cap: %v", err)
	}
	denied := policy.AllowNewUser(25)
	if denied == nil {
		t.Fatal("25 users must refuse the 26th")
	}
	if denied.Type != ManagementErrorEntitlementRequired || denied.Reason != ReasonDemoCapExceeded {
		t.Fatalf("wrong error %#v", denied)
	}
	if denied.MissingEntitlement != EntitlementDemoUser {
		t.Fatalf("entitlement %s", denied.MissingEntitlement)
	}
	if err := denied.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := policy.AllowNewOrganization(1); err == nil {
		t.Fatal("second org must be refused")
	}
	if err := policy.AllowNewInstance(1); err == nil {
		t.Fatal("second instance must be refused")
	}
}

func TestDemoCapsIgnoredOnEnterpriseAndSelfHostPublic(t *testing.T) {
	t.Parallel()
	if err := (EditionPolicy{Edition: EditionPublic, DemoCaps: false}).AllowNewUser(100); err != nil {
		t.Fatal("self-host public has no user cap")
	}
	if err := (EditionPolicy{Edition: EditionEnterprise, DemoCaps: true}).AllowNewUser(100); err != nil {
		t.Fatal("enterprise drops hosted demo caps")
	}
}
