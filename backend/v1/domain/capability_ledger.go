package domain

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"sync"
)

const CapabilityLedgerSchema = "nomen.capability-ledger.v1"

var (
	//go:embed capability-ledger.v1.json
	capabilityLedgerJSON []byte
	capabilityIDPattern  = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	capabilityLedgerOnce sync.Once
	capabilityLedgerData CapabilityLedger
	capabilityLedgerErr  error
)

type CapabilityLedger struct {
	Schema          string                   `json:"schema"`
	ProgramRevision string                   `json:"program_revision"`
	Families        []CapabilityLedgerFamily `json:"families"`
}

type CapabilityLedgerFamily struct {
	ID                  string                  `json:"id"`
	ParityScope         string                  `json:"parity_scope"`
	TargetGate          string                  `json:"target_gate"`
	EvidenceOwner       string                  `json:"evidence_owner"`
	CurrentStatus       CapabilityStatus        `json:"current_status"`
	RequiredComponents  []ComponentRole         `json:"required_components"`
	CertificationTracks []string                `json:"certification_tracks"`
	Capabilities        []CapabilityLedgerEntry `json:"capabilities"`
}

type CapabilityLedgerEntry struct {
	ID                  string           `json:"id"`
	Family              string           `json:"-"`
	ParityScope         string           `json:"-"`
	TargetGate          string           `json:"target_gate,omitempty"`
	EvidenceOwner       string           `json:"evidence_owner,omitempty"`
	CurrentStatus       CapabilityStatus `json:"current_status,omitempty"`
	RequiredComponents  []ComponentRole  `json:"required_components,omitempty"`
	CertificationTracks []string         `json:"certification_tracks,omitempty"`
	DependsOn           []string         `json:"depends_on,omitempty"`
}

// LoadCapabilityLedger returns a detached, fully resolved ledger. Family
// defaults are applied to every entry before validation so downstream callers
// never need to infer policy.
func LoadCapabilityLedger() (CapabilityLedger, error) {
	capabilityLedgerOnce.Do(func() {
		if err := json.Unmarshal(capabilityLedgerJSON, &capabilityLedgerData); err != nil {
			capabilityLedgerErr = fmt.Errorf("decode embedded capability ledger: %w", err)
			return
		}
		resolveCapabilityLedger(&capabilityLedgerData)
		capabilityLedgerErr = capabilityLedgerData.Validate()
	})
	if capabilityLedgerErr != nil {
		return CapabilityLedger{}, capabilityLedgerErr
	}
	return cloneCapabilityLedger(capabilityLedgerData), nil
}

func cloneCapabilityLedger(source CapabilityLedger) CapabilityLedger {
	cloned := source
	cloned.Families = make([]CapabilityLedgerFamily, len(source.Families))
	for familyIndex, family := range source.Families {
		cloned.Families[familyIndex] = family
		cloned.Families[familyIndex].RequiredComponents = slices.Clone(family.RequiredComponents)
		cloned.Families[familyIndex].CertificationTracks = slices.Clone(family.CertificationTracks)
		cloned.Families[familyIndex].Capabilities = make([]CapabilityLedgerEntry, len(family.Capabilities))
		for entryIndex, entry := range family.Capabilities {
			cloned.Families[familyIndex].Capabilities[entryIndex] = entry
			cloned.Families[familyIndex].Capabilities[entryIndex].RequiredComponents = slices.Clone(entry.RequiredComponents)
			cloned.Families[familyIndex].Capabilities[entryIndex].CertificationTracks = slices.Clone(entry.CertificationTracks)
			cloned.Families[familyIndex].Capabilities[entryIndex].DependsOn = slices.Clone(entry.DependsOn)
		}
	}
	return cloned
}

func (l CapabilityLedger) Entries() []CapabilityLedgerEntry {
	var entries []CapabilityLedgerEntry
	for _, family := range l.Families {
		entries = append(entries, family.Capabilities...)
	}
	return entries
}

func (l CapabilityLedger) Validate() error {
	if l.Schema != CapabilityLedgerSchema || strings.TrimSpace(l.ProgramRevision) == "" || len(l.Families) == 0 {
		return fmt.Errorf("capability ledger identity is incomplete")
	}
	ids := make(map[string]struct{})
	for _, family := range l.Families {
		if !capabilityIDPattern.MatchString(family.ID) || (family.ParityScope != "union" && family.ParityScope != "supporting") || len(family.Capabilities) == 0 {
			return fmt.Errorf("capability family %q is incomplete", family.ID)
		}
		for _, entry := range family.Capabilities {
			if !capabilityIDPattern.MatchString(entry.ID) || entry.Family != family.ID || entry.ParityScope != family.ParityScope {
				return fmt.Errorf("capability %q has incomplete identity", entry.ID)
			}
			if _, duplicate := ids[entry.ID]; duplicate {
				return fmt.Errorf("duplicate capability id %q", entry.ID)
			}
			ids[entry.ID] = struct{}{}
			if !validCapabilityGate(entry.TargetGate) || strings.TrimSpace(entry.EvidenceOwner) == "" || !entry.CurrentStatus.Valid() || len(entry.RequiredComponents) == 0 || len(entry.CertificationTracks) == 0 {
				return fmt.Errorf("capability %s has incomplete delivery policy", entry.ID)
			}
			if err := validateUniqueComponents(entry.ID, entry.RequiredComponents); err != nil {
				return err
			}
			if err := validateUniqueStrings(entry.ID+" certification_tracks", entry.CertificationTracks); err != nil {
				return err
			}
			if err := validateUniqueStrings(entry.ID+" depends_on", entry.DependsOn); err != nil {
				return err
			}
		}
	}
	for _, entry := range l.Entries() {
		for _, dependency := range entry.DependsOn {
			if dependency == entry.ID {
				return fmt.Errorf("capability %s depends on itself", entry.ID)
			}
			if _, exists := ids[dependency]; !exists {
				return fmt.Errorf("capability %s depends on unknown capability %s", entry.ID, dependency)
			}
		}
	}
	return nil
}

func resolveCapabilityLedger(ledger *CapabilityLedger) {
	for familyIndex := range ledger.Families {
		family := &ledger.Families[familyIndex]
		for entryIndex := range family.Capabilities {
			entry := &family.Capabilities[entryIndex]
			entry.Family = family.ID
			entry.ParityScope = family.ParityScope
			if entry.TargetGate == "" {
				entry.TargetGate = family.TargetGate
			}
			if entry.EvidenceOwner == "" {
				entry.EvidenceOwner = family.EvidenceOwner
			}
			if entry.CurrentStatus == "" {
				entry.CurrentStatus = family.CurrentStatus
			}
			if len(entry.RequiredComponents) == 0 {
				entry.RequiredComponents = slices.Clone(family.RequiredComponents)
			}
			if len(entry.CertificationTracks) == 0 {
				entry.CertificationTracks = slices.Clone(family.CertificationTracks)
			}
		}
	}
}

func validCapabilityGate(gate string) bool {
	return len(gate) == 2 && gate[0] == 'G' && gate[1] >= '0' && gate[1] <= '6'
}

func validateUniqueComponents(id string, components []ComponentRole) error {
	seen := make(map[ComponentRole]struct{}, len(components))
	for _, component := range components {
		if !component.Valid() {
			return fmt.Errorf("capability %s requires unknown component %q", id, component)
		}
		if _, duplicate := seen[component]; duplicate {
			return fmt.Errorf("capability %s repeats required component %s", id, component)
		}
		seen[component] = struct{}{}
	}
	return nil
}

func validateUniqueStrings(field string, values []string) error {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if !capabilityIDPattern.MatchString(value) {
			return fmt.Errorf("%s contains invalid value %q", field, value)
		}
		if _, duplicate := seen[value]; duplicate {
			return fmt.Errorf("%s repeats %q", field, value)
		}
		seen[value] = struct{}{}
	}
	return nil
}
