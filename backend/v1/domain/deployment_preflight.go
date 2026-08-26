package domain

import (
	"fmt"
	"net"
	"net/url"
	"slices"
	"strings"
	"time"
)

const DeploymentPreflightSchemaVersion uint32 = 1

type PreflightStatus string

const (
	PreflightReady          PreflightStatus = "ready"
	PreflightActionRequired PreflightStatus = "action_required"
	PreflightBlocked        PreflightStatus = "blocked"
)

type PreflightCheckStatus string

const (
	PreflightCheckPassed  PreflightCheckStatus = "passed"
	PreflightCheckWarning PreflightCheckStatus = "warning"
	PreflightCheckFailed  PreflightCheckStatus = "failed"
)

const (
	PreflightCheckDatabase             = "database"
	PreflightCheckIssuer               = "issuer"
	PreflightCheckTLS                  = "tls"
	PreflightCheckAsymmetricSigning    = "asymmetric_signing"
	PreflightCheckNotificationDelivery = "notification_delivery"
)

var preflightCheckOrder = []string{
	PreflightCheckDatabase,
	PreflightCheckIssuer,
	PreflightCheckTLS,
	PreflightCheckAsymmetricSigning,
	PreflightCheckNotificationDelivery,
}

type DeploymentPreflight struct {
	SchemaVersion    uint32           `json:"schema_version"`
	ResourceRevision string           `json:"resource_revision"`
	ObservedAt       time.Time        `json:"observed_at"`
	Status           PreflightStatus  `json:"status"`
	Issuer           string           `json:"issuer"`
	Checks           []PreflightCheck `json:"checks"`
}

type PreflightCheck struct {
	ID            string               `json:"id"`
	Status        PreflightCheckStatus `json:"status"`
	Required      bool                 `json:"required"`
	Summary       string               `json:"summary"`
	Reason        string               `json:"reason,omitempty"`
	Remediation   string               `json:"remediation,omitempty"`
	DiagnosticRef string               `json:"diagnostic_ref,omitempty"`
}

type DeploymentPreflightFacts struct {
	DatabaseAvailable          bool
	DatabaseProbeAvailable     bool
	SigningKeys                uint32
	SigningProbeAvailable      bool
	NotificationConfigured     bool
	NotificationProbeAvailable bool
}

func BuildDeploymentPreflight(facts DeploymentPreflightFacts, issuer string, observedAt time.Time) DeploymentPreflight {
	checks := []PreflightCheck{
		buildDatabaseCheck(facts),
		buildIssuerCheck(issuer),
		buildTLSCheck(issuer),
		buildSigningCheck(facts),
		buildNotificationCheck(facts),
	}
	result := DeploymentPreflight{
		SchemaVersion: DeploymentPreflightSchemaVersion,
		ObservedAt:    observedAt.UTC(),
		Issuer:        issuer,
		Status:        preflightStatus(checks),
		Checks:        checks,
	}
	result.ResourceRevision = digest(struct {
		Issuer string
		Checks []PreflightCheck
	}{issuer, checks})
	return result
}

func (p DeploymentPreflight) Validate() error {
	if p.SchemaVersion != DeploymentPreflightSchemaVersion || !validOverviewDigest(p.ResourceRevision) || p.ObservedAt.IsZero() {
		return fmt.Errorf("deployment preflight identity is incomplete")
	}
	if len(p.Checks) != len(preflightCheckOrder) {
		return fmt.Errorf("deployment preflight requires exactly %d checks", len(preflightCheckOrder))
	}
	for index, check := range p.Checks {
		if check.ID != preflightCheckOrder[index] {
			return fmt.Errorf("deployment preflight check %d must be %s", index, preflightCheckOrder[index])
		}
		if !slices.Contains([]PreflightCheckStatus{PreflightCheckPassed, PreflightCheckWarning, PreflightCheckFailed}, check.Status) || strings.TrimSpace(check.Summary) == "" {
			return fmt.Errorf("deployment preflight check %s is incomplete", check.ID)
		}
		if check.Status != PreflightCheckPassed && (strings.TrimSpace(check.Reason) == "" || strings.TrimSpace(check.Remediation) == "") {
			return fmt.Errorf("deployment preflight check %s requires reason and remediation", check.ID)
		}
		if check.DiagnosticRef != "" && !safeDiagnosticReference(check.DiagnosticRef) {
			return fmt.Errorf("deployment preflight check %s has unsafe diagnostic reference", check.ID)
		}
		if check.Required && check.Status == PreflightCheckWarning {
			return fmt.Errorf("required deployment preflight check %s cannot warn", check.ID)
		}
	}
	want := preflightStatus(p.Checks)
	if p.Status != want {
		return fmt.Errorf("deployment preflight status %s does not match checks (%s)", p.Status, want)
	}
	return nil
}

func preflightStatus(checks []PreflightCheck) PreflightStatus {
	status := PreflightReady
	for _, check := range checks {
		if check.Required && check.Status == PreflightCheckFailed {
			return PreflightBlocked
		}
		if check.Status == PreflightCheckWarning || (!check.Required && check.Status == PreflightCheckFailed) {
			status = PreflightActionRequired
		}
	}
	return status
}

func buildDatabaseCheck(facts DeploymentPreflightFacts) PreflightCheck {
	if !facts.DatabaseProbeAvailable {
		return failedPreflightCheck(PreflightCheckDatabase, true, "PostgreSQL readiness could not be verified.", "database_probe_unavailable", "Restore the configured PostgreSQL connection and run preflight again.", "preflight.database.probe")
	}
	if !facts.DatabaseAvailable {
		return failedPreflightCheck(PreflightCheckDatabase, true, "PostgreSQL is not reachable.", "database_unavailable", "Restore PostgreSQL connectivity and run preflight again.", "preflight.database.unavailable")
	}
	return passedPreflightCheck(PreflightCheckDatabase, true, "PostgreSQL answered the Nomen health probe.")
}

func buildIssuerCheck(raw string) PreflightCheck {
	issuer, ok := validPreflightIssuer(raw)
	if !ok {
		return failedPreflightCheck(PreflightCheckIssuer, true, "The public issuer is invalid.", "issuer_invalid", "Configure one absolute HTTP or HTTPS issuer without user information, query, or fragment.", "preflight.issuer.invalid")
	}
	if net.ParseIP(issuer.Hostname()) != nil {
		return failedPreflightCheck(PreflightCheckIssuer, true, "The public issuer uses an IP literal that browsers reject for WebAuthn.", "webauthn_rp_id_invalid", "Configure a stable DNS hostname and certificate for the public issuer.", "preflight.issuer.webauthn_rp_id")
	}
	return passedPreflightCheck(PreflightCheckIssuer, true, "Public issuer is absolute and structurally valid: "+issuer.Scheme+"://"+issuer.Host)
}

func buildTLSCheck(raw string) PreflightCheck {
	issuer, ok := validPreflightIssuer(raw)
	if !ok {
		return failedPreflightCheck(PreflightCheckTLS, true, "TLS policy cannot be evaluated for an invalid issuer.", "issuer_invalid", "Correct the public issuer before evaluating TLS.", "preflight.tls.issuer")
	}
	if issuer.Scheme == "https" {
		return passedPreflightCheck(PreflightCheckTLS, true, "The public issuer requires HTTPS.")
	}
	if isLoopbackHost(issuer.Hostname()) {
		return PreflightCheck{ID: PreflightCheckTLS, Status: PreflightCheckWarning, Required: false, Summary: "Loopback development issuer uses HTTP.", Reason: "loopback_http", Remediation: "Terminate TLS and configure the stable HTTPS issuer before production evidence is collected.", DiagnosticRef: "preflight.tls.loopback"}
	}
	return failedPreflightCheck(PreflightCheckTLS, true, "The public issuer does not require HTTPS.", "https_required", "Terminate TLS at the deployment edge and configure an HTTPS issuer.", "preflight.tls.required")
}

func buildSigningCheck(facts DeploymentPreflightFacts) PreflightCheck {
	if !facts.SigningProbeAvailable {
		return failedPreflightCheck(PreflightCheckAsymmetricSigning, true, "Asymmetric signing readiness could not be verified.", "signing_probe_unavailable", "Restore signing-key access and run preflight again.", "preflight.signing.probe")
	}
	if facts.SigningKeys == 0 {
		return failedPreflightCheck(PreflightCheckAsymmetricSigning, true, "No active asymmetric signing key is available.", "signing_key_missing", "Create or restore an active asymmetric web signing key.", "preflight.signing.missing")
	}
	return passedPreflightCheck(PreflightCheckAsymmetricSigning, true, fmt.Sprintf("%d active asymmetric signing key(s) are available.", facts.SigningKeys))
}

func buildNotificationCheck(facts DeploymentPreflightFacts) PreflightCheck {
	if !facts.NotificationProbeAvailable {
		return PreflightCheck{ID: PreflightCheckNotificationDelivery, Status: PreflightCheckWarning, Required: false, Summary: "Notification delivery could not be verified.", Reason: "notification_probe_unavailable", Remediation: "Configure and verify an email delivery provider before inviting members.", DiagnosticRef: "preflight.notification.probe"}
	}
	if !facts.NotificationConfigured {
		return PreflightCheck{ID: PreflightCheckNotificationDelivery, Status: PreflightCheckWarning, Required: false, Summary: "No active email delivery provider is configured.", Reason: "notification_provider_missing", Remediation: "Configure and test email delivery before inviting members.", DiagnosticRef: "preflight.notification.missing"}
	}
	return passedPreflightCheck(PreflightCheckNotificationDelivery, false, "An active email delivery provider is configured.")
}

func passedPreflightCheck(id string, required bool, summary string) PreflightCheck {
	return PreflightCheck{ID: id, Status: PreflightCheckPassed, Required: required, Summary: summary}
}

func failedPreflightCheck(id string, required bool, summary, reason, remediation, diagnostic string) PreflightCheck {
	return PreflightCheck{ID: id, Status: PreflightCheckFailed, Required: required, Summary: summary, Reason: reason, Remediation: remediation, DiagnosticRef: diagnostic}
}

func validPreflightIssuer(raw string) (*url.URL, bool) {
	issuer, err := url.Parse(raw)
	if err != nil || issuer.Host == "" || issuer.User != nil || issuer.RawQuery != "" || issuer.Fragment != "" || (issuer.Scheme != "http" && issuer.Scheme != "https") {
		return nil, false
	}
	return issuer, true
}

func isLoopbackHost(host string) bool {
	return strings.EqualFold(host, "localhost") || strings.HasSuffix(strings.ToLower(host), ".localhost")
}

func safeDiagnosticReference(value string) bool {
	if len(value) > 128 || value == "" {
		return false
	}
	for _, char := range value {
		if (char < 'a' || char > 'z') && (char < '0' || char > '9') && char != '.' && char != '_' && char != '-' {
			return false
		}
	}
	return true
}
