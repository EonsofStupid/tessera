// Package capabilities contains dependency-free identifiers shared by Tessera
// domain code, workspace tooling, conformance drivers, and generated clients.
package capabilities

const (
	LDAPOutbound         = "ldap_outbound"
	LDAPInbound          = "ldap_inbound"
	ForwardAuth          = "forward_auth"
	IdentityAwareProxy   = "identity_aware_proxy"
	VisualFlowEngine     = "visual_flow_engine"
	VaultixSecretCustody = "vaultix_secret_custody"
)

// Mandatory returns a new slice so callers may sort it without changing
// process-global state.
func Mandatory() []string {
	return []string{
		LDAPOutbound,
		LDAPInbound,
		ForwardAuth,
		IdentityAwareProxy,
		VisualFlowEngine,
		VaultixSecretCustody,
	}
}
