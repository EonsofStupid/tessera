//go:build ldapconformance

package ldapoutbound

import (
	"context"
	"crypto/rand"
	"crypto/sha1" // OpenLDAP's portable {SSHA} fixture format requires SHA-1.
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	ldaplib "github.com/go-ldap/ldap/v3"

	"github.com/EonsofStupid/tessera/backend/v1/custody/fake"
	"github.com/EonsofStupid/tessera/backend/v1/domain"
)

func TestOpenLDAPConformance(t *testing.T) {
	adminPassword := requiredEnv(t, "TESSERA_LDAP_ADMIN_PASSWORD")
	userPassword := requiredEnv(t, "TESSERA_LDAP_USER_PASSWORD")
	caPEM, err := os.ReadFile(requiredEnv(t, "TESSERA_LDAP_CA_FILE"))
	if err != nil {
		t.Fatal(err)
	}
	seedDirectory(t, adminPassword, userPassword, caPEM)

	now := time.Now().UTC()
	custody := fake.New(func() time.Time { return now })
	t.Cleanup(custody.Close)
	_, err = custody.Enroll(context.Background(), domain.EnrollSecretRequest{
		ReferenceID: "openldap-bind", AccountID: "account-01", WorkspaceID: "workspace-01",
		Purpose: domain.SecretPurposeLDAPBind, OperationID: "ldap-enroll-01",
	}, strings.NewReader(adminPassword))
	if err != nil {
		t.Fatal(err)
	}
	client := New(custody)

	for _, endpoint := range []string{"ldap://localhost:1389", "ldaps://localhost:1636"} {
		t.Run(endpoint, func(t *testing.T) {
			connector := conformanceConnector(endpoint, string(caPEM))
			user, err := client.Authenticate(context.Background(), connector, "alice", []byte(userPassword), "ldap-auth-01")
			if err != nil {
				t.Fatal(err)
			}
			if user.Username != "alice" || user.Email != "alice@example.test" || user.Disabled || strings.Join(user.Groups, ",") != "admins,developers" {
				t.Fatalf("mapped user = %#v", user)
			}
			preview, err := client.Preview(context.Background(), connector, "alice", "ldap-preview-01")
			if err != nil || preview.Username != "alice" {
				t.Fatalf("preview = %#v, %v", preview, err)
			}
		})
	}

	t.Run("failover", func(t *testing.T) {
		connector := conformanceConnector("ldap://localhost:1389", string(caPEM))
		connector.Endpoints = []string{"ldap://127.0.0.1:1390", connector.Endpoints[0]}
		if _, err := client.Authenticate(context.Background(), connector, "alice", []byte(userPassword), "ldap-failover-01"); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("wrong password", func(t *testing.T) {
		_, err := client.Authenticate(context.Background(), conformanceConnector("ldaps://localhost:1636", string(caPEM)), "alice", []byte("wrong"), "ldap-wrong-01")
		assertLDAPReason(t, err, domain.LDAPRefusalInvalidCredential)
	})
	t.Run("disabled", func(t *testing.T) {
		connector := conformanceConnector("ldaps://localhost:1636", string(caPEM))
		_, err := client.Authenticate(context.Background(), connector, "disabled", []byte(userPassword), "ldap-disabled-01")
		assertLDAPReason(t, err, domain.LDAPRefusalUserDisabled)
		preview, err := client.Preview(context.Background(), connector, "disabled", "ldap-disabled-preview-01")
		if err != nil || !preview.Disabled {
			t.Fatalf("disabled preview = %#v, %v", preview, err)
		}
	})
	t.Run("required mapping value", func(t *testing.T) {
		connector := conformanceConnector("ldaps://localhost:1636", string(caPEM))
		connector.Attributes.Subject = "description"
		_, err := client.Preview(context.Background(), connector, "alice", "ldap-mapping-01")
		assertLDAPReason(t, err, domain.LDAPRefusalInvalidMapping)
	})
	t.Run("aggregate group limit", func(t *testing.T) {
		connector := conformanceConnector("ldaps://localhost:1636", string(caPEM))
		connector.ResultLimit = 1
		_, err := client.Preview(context.Background(), connector, "alice", "ldap-group-limit-01")
		assertLDAPReason(t, err, domain.LDAPRefusalResultLimit)
	})
	t.Run("paged lifecycle snapshot", func(t *testing.T) {
		connector := conformanceConnector("ldaps://localhost:1636", string(caPEM))
		connector.Effects = domain.LDAPOutboundEffects{Authenticate: true, Import: true, Reconcile: true, Deprovision: true}
		connector.LifecyclePageSize = 1
		connector.MaxSyncUsers = 10
		users, err := client.Snapshot(context.Background(), connector, "ldap-snapshot-01")
		if err != nil {
			t.Fatal(err)
		}
		if len(users) != 2 || users[0].Username != "alice" || users[0].Disabled || users[1].Username != "disabled" || !users[1].Disabled {
			t.Fatalf("lifecycle snapshot = %#v", users)
		}
		target := newAtomicMemoryTarget()
		plan, err := client.PlanLifecycle(context.Background(), connector, target, "ldap-plan-01", now)
		if err != nil || len(plan.Actions) != 2 {
			t.Fatalf("lifecycle plan = %#v, %v", plan, err)
		}
		applied, err := client.ApplyLifecycle(context.Background(), connector, target, plan, now.Add(time.Minute))
		if err != nil || applied.Created != 2 || applied.Revision != "r2" {
			t.Fatalf("lifecycle apply = %#v, %v", applied, err)
		}
		settled, err := client.PlanLifecycle(context.Background(), connector, target, "ldap-plan-settled-01", now.Add(2*time.Minute))
		if err != nil || len(settled.Actions) != 0 {
			t.Fatalf("settled lifecycle plan = %#v, %v", settled, err)
		}
		connector.MaxSyncUsers = 1
		_, err = client.Snapshot(context.Background(), connector, "ldap-snapshot-limit-01")
		assertLDAPReason(t, err, domain.LDAPRefusalResultLimit)
	})
	t.Run("removed and tenant isolation", func(t *testing.T) {
		connector := conformanceConnector("ldaps://localhost:1636", string(caPEM))
		for _, login := range []string{"removed", "bob"} {
			_, err := client.Authenticate(context.Background(), connector, login, []byte(userPassword), "ldap-missing-01")
			assertLDAPReason(t, err, domain.LDAPRefusalUserNotFound)
		}
	})
	t.Run("filter injection", func(t *testing.T) {
		_, err := client.Authenticate(context.Background(), conformanceConnector("ldaps://localhost:1636", string(caPEM)), "alice)(uid=*)", []byte(userPassword), "ldap-injection-01")
		assertLDAPReason(t, err, domain.LDAPRefusalUserNotFound)
	})
	t.Run("wrong trust root", func(t *testing.T) {
		connector := conformanceConnector("ldaps://localhost:1636", "-----BEGIN CERTIFICATE-----\ninvalid\n-----END CERTIFICATE-----")
		_, err := client.Authenticate(context.Background(), connector, "alice", []byte(userPassword), "ldap-tls-01")
		assertLDAPReason(t, err, domain.LDAPRefusalTLS)
	})
}

func conformanceConnector(endpoint, ca string) domain.LDAPOutboundConnector {
	return domain.LDAPOutboundConnector{
		ID: "openldap-01", AccountID: "account-01", WorkspaceID: "workspace-01", ResourceRevision: "revision-01", Name: "OpenLDAP conformance",
		Profile: domain.LDAPProfileOpenLDAP, Endpoints: []string{endpoint}, BindDN: "cn=admin,dc=example,dc=test", BindSecretReferenceID: "openldap-bind",
		UserBaseDN: "ou=people,dc=example,dc=test", GroupBaseDN: "ou=groups,dc=example,dc=test",
		UserObjectClasses: []string{"inetOrgPerson"}, LoginAttributes: []string{"uid", "mail"},
		Attributes:    domain.LDAPAttributeMapping{Subject: "entryUUID", Username: "uid", FirstName: "givenName", LastName: "sn", DisplayName: "cn", Email: "mail"},
		Groups:        domain.LDAPGroupMapping{Name: "cn", Member: "member", ObjectClasses: []string{"groupOfNames"}, Traversal: domain.LDAPGroupsNested, MaxDepth: 5},
		DisabledUsers: domain.LDAPDisabledUserPolicy{Attribute: "employeeType", Values: []string{"disabled"}},
		Effects:       domain.LDAPOutboundEffects{Authenticate: true}, ConnectTimeout: time.Second, SearchTimeout: 3 * time.Second, ResultLimit: 100, TrustAnchorsPEM: ca,
	}
}

func seedDirectory(t *testing.T, adminPassword, userPassword string, caPEM []byte) {
	t.Helper()
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(caPEM) {
		t.Fatal("invalid test CA")
	}
	conn, err := ldaplib.DialURL("ldaps://localhost:1636", ldaplib.DialWithTLSConfig(&tls.Config{MinVersion: tls.VersionTLS12, RootCAs: roots, ServerName: "localhost"}))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := conn.Bind("cn=admin,dc=example,dc=test", adminPassword); err != nil {
		t.Fatal(err)
	}
	entries := []struct {
		dn    string
		attrs map[string][]string
	}{
		{"ou=people,dc=example,dc=test", map[string][]string{"objectClass": {"top", "organizationalUnit"}, "ou": {"people"}}},
		{"ou=other,dc=example,dc=test", map[string][]string{"objectClass": {"top", "organizationalUnit"}, "ou": {"other"}}},
		{"ou=groups,dc=example,dc=test", map[string][]string{"objectClass": {"top", "organizationalUnit"}, "ou": {"groups"}}},
		{"uid=alice,ou=people,dc=example,dc=test", userAttrs(t, "alice", "Alice", userPassword, "active")},
		{"uid=disabled,ou=people,dc=example,dc=test", userAttrs(t, "disabled", "Disabled", userPassword, "disabled")},
		{"uid=bob,ou=other,dc=example,dc=test", userAttrs(t, "bob", "Bob", userPassword, "active")},
		{"cn=developers,ou=groups,dc=example,dc=test", map[string][]string{"objectClass": {"top", "groupOfNames"}, "cn": {"developers"}, "member": {"uid=alice,ou=people,dc=example,dc=test"}}},
		{"cn=admins,ou=groups,dc=example,dc=test", map[string][]string{"objectClass": {"top", "groupOfNames"}, "cn": {"admins"}, "member": {"cn=developers,ou=groups,dc=example,dc=test"}}},
	}
	for _, entry := range entries {
		request := ldaplib.NewAddRequest(entry.dn, nil)
		for name, values := range entry.attrs {
			request.Attribute(name, values)
		}
		if err := conn.Add(request); err != nil && !ldaplib.IsErrorWithCode(err, ldaplib.LDAPResultEntryAlreadyExists) {
			t.Fatalf("add %s: %v", entry.dn, err)
		}
	}
}

func userAttrs(t *testing.T, uid, name, password, state string) map[string][]string {
	t.Helper()
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		t.Fatal(err)
	}
	digest := sha1.New()
	_, _ = digest.Write([]byte(password))
	_, _ = digest.Write(salt)
	encoded := append(digest.Sum(nil), salt...)
	passwordHash := "{SSHA}" + base64.StdEncoding.EncodeToString(encoded)
	clear(salt)
	clear(encoded)
	return map[string][]string{"objectClass": {"top", "person", "organizationalPerson", "inetOrgPerson"}, "uid": {uid}, "cn": {name}, "sn": {name}, "givenName": {name}, "mail": {uid + "@example.test"}, "employeeType": {state}, "userPassword": {passwordHash}}
}

func assertLDAPReason(t *testing.T, err error, want domain.LDAPOutboundRefusal) {
	t.Helper()
	var refusal *domain.LDAPOutboundError
	if !errors.As(err, &refusal) || refusal.Reason != want {
		t.Fatalf("error = %T %v, want %s", err, err, want)
	}
}

func requiredEnv(t *testing.T, name string) string {
	t.Helper()
	value := os.Getenv(name)
	if value == "" {
		t.Fatalf("%s is required", name)
	}
	return value
}
