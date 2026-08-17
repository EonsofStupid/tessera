package domain

import "strings"

// SeatAudienceScope is how a caller asks for a seat token, and it is the only
// way a workspace ever reaches `aud`.
//
// Zitadel builds an audience out of project and client ids, and RFC 8693's
// exchange refuses any requested audience that was not already in the subject
// token — deliberately, because letting a caller name its own audience is
// privilege escalation. So a workspace audience cannot be *asked for* at
// exchange time; it has to originate at the first mint, from a scope, exactly
// the way `urn:zitadel:iam:org:project:id:{id}:aud` already does.
//
//	urn:shippin:audience:automaton:ws-0001  →  aud: ["automaton:ws-0001"]
//
// The value is the audience entry verbatim rather than a bare workspace id, so
// one token names one consumer in one workspace and the contract's `aud` rule
// stays a single comparison for the verifier.
//
// **This scope grants nothing on its own.** It adds an entry to `aud` and
// nothing more; whether the member may occupy that workspace is decided at mint
// time against their stored entitlement, and a token for a workspace they do
// not occupy is refused rather than issued smaller. Naming a scope is a
// request, never a permission — which is why the check does not live here.
const SeatAudienceScope = "urn:shippin:audience:"

// AddSeatAudienceScopeToAudience adds the audience entries named by
// [SeatAudienceScope] scopes.
//
// Deliberately not folded into [AddAudScopeToAudience]: that function also
// builds the *role* audience in `userinfo.go`, where a workspace entry would be
// read as a project whose roles should be asserted. Two callers wanting
// different things from one list is how an audience quietly grows entries
// nobody asked for.
func AddSeatAudienceScopeToAudience(audience, scopes []string) []string {
	for _, scope := range scopes {
		entry, found := strings.CutPrefix(scope, SeatAudienceScope)
		if !found {
			continue
		}
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		audience = addProjectID(audience, entry) // de-duplicating append
	}
	return audience
}
