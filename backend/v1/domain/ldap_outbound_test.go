package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func validLDAPOutboundConnector() LDAPOutboundConnector {
	return LDAPOutboundConnector{
		ID: "ldap-01", AccountID: "account-01", WorkspaceID: "workspace-01", ResourceRevision: "revision-01",
		Name: "Company directory", Profile: LDAPProfileOpenLDAP,
		Endpoints: []string{"ldap://ldap.example.test:389", "ldaps://ldap.example.test:636"},
		BindDN:    "cn=reader,dc=example,dc=test", BindSecretReferenceID: "secret-ref-01",
		UserBaseDN: "ou=people,dc=example,dc=test", GroupBaseDN: "ou=groups,dc=example,dc=test",
		UserObjectClasses: []string{"inetOrgPerson"}, LoginAttributes: []string{"uid", "mail"},
		Attributes:    LDAPAttributeMapping{Subject: "entryUUID", Username: "uid", FirstName: "givenName", LastName: "sn", Email: "mail"},
		Groups:        LDAPGroupMapping{Name: "cn", Member: "member", ObjectClasses: []string{"groupOfNames"}, Traversal: LDAPGroupsNested, MaxDepth: 5},
		DisabledUsers: LDAPDisabledUserPolicy{Attribute: "employeeType", Values: []string{"disabled"}},
		Effects:       LDAPOutboundEffects{Authenticate: true}, ConnectTimeout: time.Second, SearchTimeout: 2 * time.Second, ResultLimit: 100,
	}
}

func ldapRefusal(t *testing.T, err error) *LDAPOutboundError {
	t.Helper()
	var refusal *LDAPOutboundError
	if !errors.As(err, &refusal) {
		t.Fatalf("error = %T %v, want LDAP refusal", err, err)
	}
	return refusal
}

func TestLDAPOutboundConnectorValidation(t *testing.T) {
	t.Parallel()
	connector := validLDAPOutboundConnector()
	if err := connector.Validate(); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		mutate func(*LDAPOutboundConnector)
	}{
		{"plaintext", func(value *LDAPOutboundConnector) { value.Endpoints = []string{"http://ldap.example.test:389"} }},
		{"endpoint credential", func(value *LDAPOutboundConnector) {
			value.Endpoints = []string{"ldaps://user:pass@ldap.example.test:636"}
		}},
		{"unbounded timeout", func(value *LDAPOutboundConnector) { value.ConnectTimeout = 31 * time.Second }},
		{"deprovision alone", func(value *LDAPOutboundConnector) { value.Effects = LDAPOutboundEffects{Deprovision: true} }},
		{"unknown attribute", func(value *LDAPOutboundConnector) { value.LoginAttributes = []string{"uid)(objectClass=*"} }},
		{"invalid descriptor start", func(value *LDAPOutboundConnector) { value.LoginAttributes = []string{"-uid"} }},
		{"invalid bind dn", func(value *LDAPOutboundConnector) { value.BindDN = "cn=reader,dc" }},
		{"invalid group base dn", func(value *LDAPOutboundConnector) { value.GroupBaseDN = "ou=groups,dc" }},
		{"invalid group class", func(value *LDAPOutboundConnector) { value.Groups.ObjectClasses = []string{"group)(objectClass=*"} }},
		{"unbounded nesting", func(value *LDAPOutboundConnector) { value.Groups.MaxDepth = 11 }},
		{"lifecycle without paging", func(value *LDAPOutboundConnector) { value.Effects.Import = true }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			changed := connector
			test.mutate(&changed)
			if got := ldapRefusal(t, changed.Validate()).Reason; got != LDAPRefusalInvalidConnector {
				t.Fatalf("reason = %s", got)
			}
		})
	}
}

func TestLDAPUserSearchEscapesLogin(t *testing.T) {
	t.Parallel()
	connector := validLDAPOutboundConnector()
	filter := connector.UserSearchFilter("alice)(uid=*)")
	if strings.Contains(filter, "alice)(uid=*)") {
		t.Fatalf("login was not escaped: %s", filter)
	}
	for _, want := range []string{"(objectClass=inetOrgPerson)", `(uid=alice\29\28uid=\2a\29)`, `(mail=alice\29\28uid=\2a\29)`} {
		if !strings.Contains(filter, want) {
			t.Errorf("filter %q missing %q", filter, want)
		}
	}
}

func TestLDAPDisabledPolicySupportsValuesAndADBitMask(t *testing.T) {
	t.Parallel()
	if !((LDAPDisabledUserPolicy{Attribute: "employeeType", Values: []string{"disabled"}}).Disabled("DISABLED")) {
		t.Fatal("value policy did not disable user")
	}
	if !((LDAPDisabledUserPolicy{Attribute: "userAccountControl", BitMask: 2}).Disabled("514")) {
		t.Fatal("AD account-disabled bit was not detected")
	}
	if (LDAPDisabledUserPolicy{Attribute: "userAccountControl", BitMask: 2}).Disabled("512") {
		t.Fatal("enabled AD account was disabled")
	}
}
