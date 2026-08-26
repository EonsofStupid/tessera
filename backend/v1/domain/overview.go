package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"
)

const (
	OverviewSchemaVersion uint32 = 1
	OverviewServiceID            = "nomen"
)

type OverviewReadinessStatus string

const (
	OverviewReady    OverviewReadinessStatus = "ready"
	OverviewDegraded OverviewReadinessStatus = "degraded"
)

type OverviewLensStatus string

const (
	OverviewLensReady     OverviewLensStatus = "ready"
	OverviewLensAttention OverviewLensStatus = "attention"
	OverviewLensQuiet     OverviewLensStatus = "quiet"
)

type OverviewPartyProtocol string

const (
	OverviewProtocolOIDC   OverviewPartyProtocol = "OIDC"
	OverviewProtocolOAuth2 OverviewPartyProtocol = "OAuth2"
	OverviewProtocolSAML   OverviewPartyProtocol = "SAML"
	OverviewProtocolLDAP   OverviewPartyProtocol = "LDAP"
)

type OverviewPartyStatus string

const (
	OverviewPartyReady     OverviewPartyStatus = "ready"
	OverviewPartyAttention OverviewPartyStatus = "attention"
	OverviewPartyDraft     OverviewPartyStatus = "draft"
)

type OverviewActivityResult string

const (
	OverviewActivityAllowed OverviewActivityResult = "allowed"
	OverviewActivityRefused OverviewActivityResult = "refused"
	OverviewActivityChanged OverviewActivityResult = "changed"
)

type Overview struct {
	SchemaVersion    uint32             `json:"schema_version"`
	ServiceID        string             `json:"service_id"`
	ResourceRevision string             `json:"resource_revision"`
	ObservedAt       time.Time          `json:"observed_at"`
	Readiness        OverviewReadiness  `json:"readiness"`
	Lenses           []OverviewLens     `json:"lenses"`
	Federation       OverviewFederation `json:"federation"`
	Activity         []OverviewActivity `json:"activity"`
}

type OverviewReadiness struct {
	Status         OverviewReadinessStatus `json:"status"`
	Issuer         string                  `json:"issuer"`
	SigningKeys    uint32                  `json:"signing_keys"`
	Flows          uint64                  `json:"flows"`
	PolicyRevision string                  `json:"policy_revision"`
	Reasons        []string                `json:"reasons"`
}

type OverviewLens struct {
	ID     string             `json:"id"`
	Label  string             `json:"label"`
	Value  uint64             `json:"value"`
	Unit   string             `json:"unit"`
	Detail string             `json:"detail"`
	Status OverviewLensStatus `json:"status"`
}

type OverviewFederationParty struct {
	ID       string                `json:"id"`
	Name     string                `json:"name"`
	Protocol OverviewPartyProtocol `json:"protocol"`
	Status   OverviewPartyStatus   `json:"status"`
	Detail   string                `json:"detail"`
}

type OverviewFederation struct {
	Upstreams []OverviewFederationParty `json:"upstreams"`
	Clients   []OverviewFederationParty `json:"clients"`
}

type OverviewActivity struct {
	ID     string                 `json:"id"`
	At     time.Time              `json:"at"`
	Actor  string                 `json:"actor"`
	Action string                 `json:"action"`
	Target string                 `json:"target"`
	Result OverviewActivityResult `json:"result"`
}

// OverviewFacts are Nomen-owned source facts. Host-product inventory and
// billing deliberately cannot enter this assembler.
type OverviewFacts struct {
	WorkspaceAttachments uint64
	AgentSeats           uint64
	HumanSeats           uint64
	Flows                uint64
	PolicyRevisions      []string
}

func BuildOverview(facts OverviewFacts, issuer string, signingKeys uint32, observedAt time.Time) Overview {
	policyRevisions := compactSorted(facts.PolicyRevisions)
	policyRevision := digest(policyRevisions)
	reasons := make([]string, 0, 2)
	if signingKeys == 0 {
		reasons = append(reasons, "no_active_asymmetric_signing_key")
	}
	if facts.Flows == 0 {
		reasons = append(reasons, "no_nomen_flow_configured")
	}
	status := OverviewReady
	if len(reasons) != 0 {
		status = OverviewDegraded
	}

	overview := Overview{
		SchemaVersion: OverviewSchemaVersion,
		ServiceID:     OverviewServiceID,
		ObservedAt:    observedAt.UTC(),
		Readiness: OverviewReadiness{
			Status:         status,
			Issuer:         issuer,
			SigningKeys:    signingKeys,
			Flows:          facts.Flows,
			PolicyRevision: policyRevision,
			Reasons:        reasons,
		},
		Lenses: []OverviewLens{
			lens("infrastructure", "Infrastructure", facts.WorkspaceAttachments, "attachments", "Workspace identity attachments managed by Nomen."),
			lens("ai", "AI", facts.AgentSeats, "agent seats", "Agent seats with explicit identity and scope."),
			lens("customers", "Customers", facts.HumanSeats, "human seats", "Human seats managed by Nomen."),
		},
		Federation: OverviewFederation{
			Upstreams: []OverviewFederationParty{},
			Clients:   []OverviewFederationParty{},
		},
		Activity: []OverviewActivity{},
	}
	overview.ResourceRevision = digest(struct {
		Issuer               string
		SigningKeys          uint32
		WorkspaceAttachments uint64
		AgentSeats           uint64
		HumanSeats           uint64
		Flows                uint64
		Policies             []string
	}{issuer, signingKeys, facts.WorkspaceAttachments, facts.AgentSeats, facts.HumanSeats, facts.Flows, policyRevisions})
	return overview
}

func (o Overview) Validate() error {
	if o.SchemaVersion != OverviewSchemaVersion || o.ServiceID != OverviewServiceID {
		return fmt.Errorf("unsupported overview identity %d/%q", o.SchemaVersion, o.ServiceID)
	}
	if !validOverviewDigest(o.ResourceRevision) || !validOverviewDigest(o.Readiness.PolicyRevision) {
		return fmt.Errorf("overview revisions must be lowercase sha256 digests")
	}
	if o.ObservedAt.IsZero() {
		return fmt.Errorf("overview observed_at is required")
	}
	issuer, err := url.Parse(o.Readiness.Issuer)
	if err != nil || issuer.Host == "" || (issuer.Scheme != "https" && issuer.Scheme != "http") {
		return fmt.Errorf("overview issuer must be an absolute http(s) URL")
	}
	if o.Readiness.Status != OverviewReady && o.Readiness.Status != OverviewDegraded {
		return fmt.Errorf("unknown overview readiness status %q", o.Readiness.Status)
	}
	if o.Readiness.Status == OverviewReady && (o.Readiness.SigningKeys == 0 || o.Readiness.Flows == 0 || len(o.Readiness.Reasons) != 0) {
		return fmt.Errorf("ready overview requires signing keys, flows and no degraded reasons")
	}
	if o.Readiness.Status == OverviewDegraded && len(o.Readiness.Reasons) == 0 {
		return fmt.Errorf("degraded overview requires a reason")
	}
	if err := validateOverviewLenses(o.Lenses); err != nil {
		return err
	}
	for _, party := range append(slices.Clone(o.Federation.Upstreams), o.Federation.Clients...) {
		if strings.TrimSpace(party.ID) == "" || strings.TrimSpace(party.Name) == "" || strings.TrimSpace(party.Detail) == "" {
			return fmt.Errorf("federation party has incomplete identity")
		}
		if !slices.Contains([]OverviewPartyProtocol{OverviewProtocolOIDC, OverviewProtocolOAuth2, OverviewProtocolSAML, OverviewProtocolLDAP}, party.Protocol) {
			return fmt.Errorf("unknown federation protocol %q", party.Protocol)
		}
		if !slices.Contains([]OverviewPartyStatus{OverviewPartyReady, OverviewPartyAttention, OverviewPartyDraft}, party.Status) {
			return fmt.Errorf("unknown federation status %q", party.Status)
		}
	}
	for _, activity := range o.Activity {
		if strings.TrimSpace(activity.ID) == "" || activity.At.IsZero() || strings.TrimSpace(activity.Actor) == "" || strings.TrimSpace(activity.Action) == "" || strings.TrimSpace(activity.Target) == "" {
			return fmt.Errorf("overview activity has incomplete identity")
		}
		if !slices.Contains([]OverviewActivityResult{OverviewActivityAllowed, OverviewActivityRefused, OverviewActivityChanged}, activity.Result) {
			return fmt.Errorf("unknown overview activity result %q", activity.Result)
		}
	}
	return nil
}

func lens(id, label string, value uint64, unit, detail string) OverviewLens {
	status := OverviewLensReady
	if value == 0 {
		status = OverviewLensQuiet
	}
	return OverviewLens{ID: id, Label: label, Value: value, Unit: unit, Detail: detail, Status: status}
}

func validateOverviewLenses(lenses []OverviewLens) error {
	if len(lenses) != 3 {
		return fmt.Errorf("overview requires exactly three identity lenses")
	}
	want := map[string]struct{}{"infrastructure": {}, "ai": {}, "customers": {}}
	for _, lens := range lenses {
		if _, ok := want[lens.ID]; !ok {
			return fmt.Errorf("unknown or duplicate overview lens %q", lens.ID)
		}
		delete(want, lens.ID)
		if strings.TrimSpace(lens.Label) == "" || strings.TrimSpace(lens.Unit) == "" || strings.TrimSpace(lens.Detail) == "" {
			return fmt.Errorf("overview lens %s is incomplete", lens.ID)
		}
		if !slices.Contains([]OverviewLensStatus{OverviewLensReady, OverviewLensAttention, OverviewLensQuiet}, lens.Status) {
			return fmt.Errorf("unknown overview lens status %q", lens.Status)
		}
	}
	return nil
}

func compactSorted(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" && !slices.Contains(result, value) {
			result = append(result, value)
		}
	}
	slices.Sort(result)
	return result
}

func digest(value any) string {
	encoded, _ := json.Marshal(value)
	sum := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func validOverviewDigest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != len("sha256:")+sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil && value == strings.ToLower(value)
}
