package domain

import "strings"

const (
	EditionPublic     = "public"
	EditionEnterprise = "enterprise"

	ReasonEditionPublicWithheld = "edition_public_withheld"
	ReasonDemoCapExceeded       = "demo_cap_exceeded"

	EntitlementDemoInstance     = "nomen.demo.instance"
	EntitlementDemoOrganization = "nomen.demo.organization"
	EntitlementDemoUser         = "nomen.demo.user"

	DemoCapInstances     = 1
	DemoCapOrganizations = 1
	DemoCapUsers         = 25
)

// NormalizeEdition maps config to public or enterprise. Empty is public.
func NormalizeEdition(edition string) string {
	if strings.EqualFold(strings.TrimSpace(edition), EditionEnterprise) {
		return EditionEnterprise
	}
	return EditionPublic
}

func RedisAllowed(edition string) bool {
	return NormalizeEdition(edition) == EditionEnterprise
}

func PublicWithheldCapability(id string) bool {
	switch id {
	case CapabilityIDVaultixSecretCustody, CapabilityIDAnalyticsOLAP, "high_availability":
		return true
	default:
		return false
	}
}

func PublicWithheldComponent(role ComponentRole) bool {
	return role == ComponentVaultix || role == ComponentZuul || role == ComponentClickHouse
}

// EditionPolicy is evaluated at runtime. Hosted demo caps are not applied to
// community self-host public.
type EditionPolicy struct {
	Edition  string
	DemoCaps bool
}

func (p EditionPolicy) Normalized() EditionPolicy {
	p.Edition = NormalizeEdition(p.Edition)
	if p.Edition != EditionPublic {
		p.DemoCaps = false
	}
	return p
}

func (p EditionPolicy) AllowNewInstance(current uint64) *ManagementError {
	return p.denyIfDemoCap(current, DemoCapInstances, EntitlementDemoInstance, "This hosted demo allows one instance.")
}

func (p EditionPolicy) AllowNewOrganization(current uint64) *ManagementError {
	return p.denyIfDemoCap(current, DemoCapOrganizations, EntitlementDemoOrganization, "This hosted demo allows one organization.")
}

func (p EditionPolicy) AllowNewUser(current uint64) *ManagementError {
	return p.denyIfDemoCap(current, DemoCapUsers, EntitlementDemoUser, "This hosted demo allows 25 users.")
}

func (p EditionPolicy) denyIfDemoCap(current, max uint64, entitlement, message string) *ManagementError {
	p = p.Normalized()
	if !p.DemoCaps {
		return nil
	}
	if current < max {
		return nil
	}
	err := ManagementError{
		Type:    ManagementErrorEntitlementRequired,
		Reason:  ReasonDemoCapExceeded,
		Message: message,
		Remedy: ManagementRemedy{
			Kind:  ManagementRemedyRequestEntitlement,
			Label: defaultRemedyLabel(ManagementRemedyRequestEntitlement),
		},
		Retry:              RetryOperatorAction,
		MissingEntitlement: entitlement,
	}
	return &err
}
