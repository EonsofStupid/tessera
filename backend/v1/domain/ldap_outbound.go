package domain

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	ldaplib "github.com/go-ldap/ldap/v3"
)

type LDAPDirectoryProfile string

const (
	LDAPProfileGeneric         LDAPDirectoryProfile = "generic"
	LDAPProfileOpenLDAP        LDAPDirectoryProfile = "openldap"
	LDAPProfileActiveDirectory LDAPDirectoryProfile = "active_directory"
)

func (p LDAPDirectoryProfile) Valid() bool {
	return p == LDAPProfileGeneric || p == LDAPProfileOpenLDAP || p == LDAPProfileActiveDirectory
}

type LDAPGroupTraversal string

const (
	LDAPGroupsDirect LDAPGroupTraversal = "direct"
	LDAPGroupsNested LDAPGroupTraversal = "nested"
)

type LDAPOutboundEffects struct {
	Authenticate bool `json:"authenticate"`
	Import       bool `json:"import"`
	Reconcile    bool `json:"reconcile"`
	Deprovision  bool `json:"deprovision"`
}

type LDAPAttributeMapping struct {
	Subject     string `json:"subject"`
	Username    string `json:"username"`
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	Email       string `json:"email,omitempty"`
	Phone       string `json:"phone,omitempty"`
}

func (m LDAPAttributeMapping) Requested() []string {
	values := []string{m.Subject, m.Username, m.FirstName, m.LastName, m.DisplayName, m.Email, m.Phone}
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, value)
	}
	return out
}

type LDAPGroupMapping struct {
	Name          string             `json:"name"`
	Member        string             `json:"member"`
	ObjectClasses []string           `json:"object_classes"`
	Traversal     LDAPGroupTraversal `json:"traversal"`
	MaxDepth      uint32             `json:"max_depth"`
}

type LDAPDisabledUserPolicy struct {
	Attribute string   `json:"attribute,omitempty"`
	Values    []string `json:"values,omitempty"`
	BitMask   uint32   `json:"bit_mask,omitempty"`
}

func (p LDAPDisabledUserPolicy) Disabled(raw string) bool {
	if p.Attribute == "" || raw == "" {
		return false
	}
	for _, value := range p.Values {
		if strings.EqualFold(strings.TrimSpace(raw), strings.TrimSpace(value)) {
			return true
		}
	}
	if p.BitMask != 0 {
		value, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 32)
		return err == nil && uint32(value)&p.BitMask != 0
	}
	return false
}

type LDAPOutboundConnector struct {
	ID                    string                 `json:"connector_id"`
	AccountID             string                 `json:"account_id"`
	WorkspaceID           string                 `json:"workspace_id"`
	ResourceRevision      string                 `json:"resource_revision"`
	Name                  string                 `json:"name"`
	Profile               LDAPDirectoryProfile   `json:"profile"`
	Endpoints             []string               `json:"endpoints"`
	BindDN                string                 `json:"bind_dn"`
	BindSecretReferenceID string                 `json:"bind_secret_reference_id"`
	UserBaseDN            string                 `json:"user_base_dn"`
	GroupBaseDN           string                 `json:"group_base_dn,omitempty"`
	UserObjectClasses     []string               `json:"user_object_classes"`
	LoginAttributes       []string               `json:"login_attributes"`
	Attributes            LDAPAttributeMapping   `json:"attributes"`
	Groups                LDAPGroupMapping       `json:"groups"`
	DisabledUsers         LDAPDisabledUserPolicy `json:"disabled_users"`
	Effects               LDAPOutboundEffects    `json:"effects"`
	ConnectTimeout        time.Duration          `json:"connect_timeout"`
	SearchTimeout         time.Duration          `json:"search_timeout"`
	ResultLimit           uint32                 `json:"result_limit"`
	TrustAnchorsPEM       string                 `json:"trust_anchors_pem,omitempty"`
}

func (c LDAPOutboundConnector) Validate() error {
	switch {
	case !safeIdentifier(c.ID):
		return ldapOutboundError(LDAPRefusalInvalidConnector, "connector_id", "a safe stable connector id is required")
	case !safeIdentifier(c.AccountID) || !safeIdentifier(c.WorkspaceID):
		return ldapOutboundError(LDAPRefusalInvalidConnector, "scope", "account and workspace boundaries are required")
	case !safeIdentifier(c.ResourceRevision):
		return ldapOutboundError(LDAPRefusalInvalidConnector, "resource_revision", "a safe resource revision is required")
	case strings.TrimSpace(c.Name) == "" || len(c.Name) > 200:
		return ldapOutboundError(LDAPRefusalInvalidConnector, "name", "a display name is required")
	case !c.Profile.Valid():
		return ldapOutboundError(LDAPRefusalInvalidConnector, "profile", "a supported directory profile is required")
	case len(c.Endpoints) == 0 || len(c.Endpoints) > 8:
		return ldapOutboundError(LDAPRefusalInvalidConnector, "endpoints", "one to eight TLS endpoints are required")
	case strings.TrimSpace(c.BindDN) == "":
		return ldapOutboundError(LDAPRefusalInvalidConnector, "bind_dn", "a service bind DN is required")
	case !validLDAPDN(c.BindDN):
		return ldapOutboundError(LDAPRefusalInvalidConnector, "bind_dn", "the service bind DN is invalid")
	case !safeIdentifier(c.BindSecretReferenceID):
		return ldapOutboundError(LDAPRefusalInvalidConnector, "bind_secret_reference_id", "a Tessera secret reference id is required")
	case strings.TrimSpace(c.UserBaseDN) == "":
		return ldapOutboundError(LDAPRefusalInvalidConnector, "user_base_dn", "a tenant-scoped user base DN is required")
	case !validLDAPDN(c.UserBaseDN):
		return ldapOutboundError(LDAPRefusalInvalidConnector, "user_base_dn", "the tenant-scoped user base DN is invalid")
	case len(c.UserObjectClasses) == 0 || len(c.LoginAttributes) == 0:
		return ldapOutboundError(LDAPRefusalInvalidConnector, "user_search", "object classes and login attributes are required")
	case !validLDAPAttribute(c.Attributes.Subject) || !validLDAPAttribute(c.Attributes.Username):
		return ldapOutboundError(LDAPRefusalInvalidConnector, "attributes", "subject and username mappings are required")
	case !c.Effects.Authenticate && !c.Effects.Import && !c.Effects.Reconcile && !c.Effects.Deprovision:
		return ldapOutboundError(LDAPRefusalInvalidConnector, "effects", "at least one explicit lifecycle effect is required")
	case c.Effects.Deprovision && !c.Effects.Reconcile:
		return ldapOutboundError(LDAPRefusalInvalidConnector, "effects.deprovision", "deprovision requires reconcile")
	case c.ConnectTimeout < time.Second || c.ConnectTimeout > 30*time.Second:
		return ldapOutboundError(LDAPRefusalInvalidConnector, "connect_timeout", "connect timeout must be between one and thirty seconds")
	case c.SearchTimeout < time.Second || c.SearchTimeout > 30*time.Second:
		return ldapOutboundError(LDAPRefusalInvalidConnector, "search_timeout", "search timeout must be between one and thirty seconds")
	case c.ResultLimit == 0 || c.ResultLimit > 1000:
		return ldapOutboundError(LDAPRefusalInvalidConnector, "result_limit", "result limit must be between one and one thousand")
	}
	for _, endpoint := range c.Endpoints {
		if !validLDAPTLSEndpoint(endpoint) {
			return ldapOutboundError(LDAPRefusalInvalidConnector, "endpoints", "endpoints must be ldap StartTLS or ldaps URLs without credentials, paths, query or fragment")
		}
	}
	for _, attribute := range append(append([]string{}, c.LoginAttributes...), c.Attributes.Requested()...) {
		if !validLDAPAttribute(attribute) {
			return ldapOutboundError(LDAPRefusalInvalidConnector, "attributes", "LDAP attribute names must use the schema identifier grammar")
		}
	}
	for _, class := range c.UserObjectClasses {
		if !validLDAPAttribute(class) {
			return ldapOutboundError(LDAPRefusalInvalidConnector, "user_object_classes", "object class names must use the schema identifier grammar")
		}
	}
	if c.DisabledUsers.Attribute != "" && !validLDAPAttribute(c.DisabledUsers.Attribute) {
		return ldapOutboundError(LDAPRefusalInvalidConnector, "disabled_users.attribute", "disabled-user attribute is invalid")
	}
	if c.GroupBaseDN != "" {
		if !validLDAPDN(c.GroupBaseDN) {
			return ldapOutboundError(LDAPRefusalInvalidConnector, "group_base_dn", "the tenant-scoped group base DN is invalid")
		}
		if !validLDAPAttribute(c.Groups.Name) || !validLDAPAttribute(c.Groups.Member) || len(c.Groups.ObjectClasses) == 0 {
			return ldapOutboundError(LDAPRefusalInvalidConnector, "groups", "group name, member and object classes are required")
		}
		if c.Groups.Traversal != LDAPGroupsDirect && c.Groups.Traversal != LDAPGroupsNested {
			return ldapOutboundError(LDAPRefusalInvalidConnector, "groups.traversal", "direct or nested traversal is required")
		}
		if c.Groups.Traversal == LDAPGroupsNested && (c.Groups.MaxDepth == 0 || c.Groups.MaxDepth > 10) {
			return ldapOutboundError(LDAPRefusalInvalidConnector, "groups.max_depth", "nested depth must be between one and ten")
		}
		for _, class := range c.Groups.ObjectClasses {
			if !validLDAPAttribute(class) {
				return ldapOutboundError(LDAPRefusalInvalidConnector, "groups.object_classes", "group object class names must use the schema identifier grammar")
			}
		}
	}
	return nil
}

func (c LDAPOutboundConnector) UserSearchFilter(login string) string {
	classes := make([]string, 0, len(c.UserObjectClasses))
	for _, class := range c.UserObjectClasses {
		classes = append(classes, "(objectClass="+ldaplib.EscapeFilter(class)+")")
	}
	attributes := make([]string, 0, len(c.LoginAttributes))
	for _, attribute := range c.LoginAttributes {
		attributes = append(attributes, "("+attribute+"="+ldaplib.EscapeFilter(login)+")")
	}
	return "(&" + strings.Join(classes, "") + "(|" + strings.Join(attributes, "") + "))"
}

func (c LDAPOutboundConnector) RequestedUserAttributes() []string {
	values := c.Attributes.Requested()
	if c.DisabledUsers.Attribute != "" {
		values = append(values, c.DisabledUsers.Attribute)
	}
	return values
}

type LDAPOutboundUser struct {
	DN          string   `json:"dn"`
	Subject     string   `json:"subject"`
	Username    string   `json:"username"`
	FirstName   string   `json:"first_name,omitempty"`
	LastName    string   `json:"last_name,omitempty"`
	DisplayName string   `json:"display_name,omitempty"`
	Email       string   `json:"email,omitempty"`
	Phone       string   `json:"phone,omitempty"`
	Groups      []string `json:"groups"`
	Disabled    bool     `json:"disabled"`
}

type LDAPOutboundRefusal string

const (
	LDAPRefusalInvalidConnector  LDAPOutboundRefusal = "invalid_connector"
	LDAPRefusalUnavailable       LDAPOutboundRefusal = "directory_unavailable"
	LDAPRefusalTLS               LDAPOutboundRefusal = "tls_validation_failed"
	LDAPRefusalBind              LDAPOutboundRefusal = "service_bind_failed"
	LDAPRefusalUserNotFound      LDAPOutboundRefusal = "user_not_found"
	LDAPRefusalUserAmbiguous     LDAPOutboundRefusal = "user_ambiguous"
	LDAPRefusalInvalidCredential LDAPOutboundRefusal = "invalid_credentials"
	LDAPRefusalUserDisabled      LDAPOutboundRefusal = "user_disabled"
	LDAPRefusalSearch            LDAPOutboundRefusal = "directory_search_failed"
	LDAPRefusalInvalidMapping    LDAPOutboundRefusal = "invalid_mapping"
	LDAPRefusalResultLimit       LDAPOutboundRefusal = "result_limit_exceeded"
)

type LDAPOutboundError struct {
	Reason LDAPOutboundRefusal `json:"reason"`
	Field  string              `json:"field,omitempty"`
	Detail string              `json:"detail"`
}

func (e *LDAPOutboundError) Error() string {
	return fmt.Sprintf("ldap outbound: %s: %s", e.Reason, e.Detail)
}

func (e *LDAPOutboundError) custodySafe() {}

func ldapOutboundError(reason LDAPOutboundRefusal, field, detail string) error {
	return &LDAPOutboundError{Reason: reason, Field: field, Detail: detail}
}

func validLDAPTLSEndpoint(value string) bool {
	parsed, err := url.Parse(value)
	if err != nil || (parsed.Scheme != "ldap" && parsed.Scheme != "ldaps") || parsed.Hostname() == "" || parsed.Port() == "" {
		return false
	}
	return parsed.User == nil && (parsed.Path == "" || parsed.Path == "/") && parsed.RawQuery == "" && parsed.Fragment == ""
}

func validLDAPDN(value string) bool {
	_, err := ldaplib.ParseDN(value)
	return err == nil
}

func validLDAPAttribute(value string) bool {
	if value == "" || len(value) > 128 {
		return false
	}
	for index, char := range value {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (index > 0 && ((char >= '0' && char <= '9') || char == '-')) {
			continue
		}
		return false
	}
	return true
}
