// Package ldapoutbound implements Tessera's provider-neutral outbound LDAP
// boundary. The inherited login provider remains a compatibility detail; this
// package owns TLS, custody, mapping and safe refusal behavior.
package ldapoutbound

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"net/url"
	"sort"
	"strings"
	"time"

	ldaplib "github.com/go-ldap/ldap/v3"

	"github.com/EonsofStupid/tessera/backend/v1/domain"
)

type Client struct {
	custody domain.SecretCustody
}

func New(custody domain.SecretCustody) *Client {
	return &Client{custody: custody}
}

func (c *Client) Authenticate(ctx context.Context, connector domain.LDAPOutboundConnector, login string, password []byte, operationID string) (domain.LDAPOutboundUser, error) {
	if err := connector.Validate(); err != nil {
		return domain.LDAPOutboundUser{}, err
	}
	if !connector.Effects.Authenticate {
		return domain.LDAPOutboundUser{}, safeError(domain.LDAPRefusalInvalidConnector, "effects.authenticate", "authentication is not enabled for this connector")
	}
	if strings.TrimSpace(login) == "" || len(password) == 0 {
		return domain.LDAPOutboundUser{}, safeError(domain.LDAPRefusalInvalidCredential, "credentials", "the supplied credentials were not accepted")
	}
	workingPassword := append([]byte(nil), password...)
	defer clear(workingPassword)

	var user domain.LDAPOutboundUser
	_, err := c.custody.Use(ctx, domain.UseSecretRequest{
		ReferenceID: connector.BindSecretReferenceID,
		AccountID:   connector.AccountID,
		WorkspaceID: connector.WorkspaceID,
		Purpose:     domain.SecretPurposeLDAPBind,
		OperationID: operationID,
	}, func(bindSecret []byte) error {
		var useErr error
		user, useErr = authenticateWithBind(ctx, connector, login, workingPassword, bindSecret)
		return useErr
	})
	if err != nil {
		var ldapErr *domain.LDAPOutboundError
		if errors.As(err, &ldapErr) {
			return domain.LDAPOutboundUser{}, ldapErr
		}
		return domain.LDAPOutboundUser{}, err
	}
	return user, nil
}

// Preview performs the same bound, tenant-scoped lookup without a user bind.
// The caller decides whether to display the returned sample; this package does
// not persist it.
func (c *Client) Preview(ctx context.Context, connector domain.LDAPOutboundConnector, login, operationID string) (domain.LDAPOutboundUser, error) {
	if err := connector.Validate(); err != nil {
		return domain.LDAPOutboundUser{}, err
	}
	if strings.TrimSpace(login) == "" {
		return domain.LDAPOutboundUser{}, safeError(domain.LDAPRefusalUserNotFound, "login", "a sample login is required")
	}
	var user domain.LDAPOutboundUser
	_, err := c.custody.Use(ctx, domain.UseSecretRequest{
		ReferenceID: connector.BindSecretReferenceID,
		AccountID:   connector.AccountID,
		WorkspaceID: connector.WorkspaceID,
		Purpose:     domain.SecretPurposeLDAPBind,
		OperationID: operationID,
	}, func(bindSecret []byte) error {
		conn, err := connectAndBind(ctx, connector, bindSecret)
		if err != nil {
			return err
		}
		defer conn.Close()
		entry, err := findUser(conn, connector, login)
		if err != nil {
			return err
		}
		user, err = mapUser(conn, connector, entry)
		return err
	})
	if err != nil {
		var ldapErr *domain.LDAPOutboundError
		if errors.As(err, &ldapErr) {
			return domain.LDAPOutboundUser{}, ldapErr
		}
		return domain.LDAPOutboundUser{}, err
	}
	return user, nil
}

// Snapshot returns only a complete, bounded lifecycle view. Any paging,
// mapping, group, context or limit failure returns no users so a caller cannot
// mistake a partial directory response for upstream deletion.
func (c *Client) Snapshot(ctx context.Context, connector domain.LDAPOutboundConnector, operationID string) ([]domain.LDAPOutboundUser, error) {
	if err := connector.Validate(); err != nil {
		return nil, err
	}
	if !connector.Effects.Import && !connector.Effects.Reconcile {
		return nil, safeError(domain.LDAPRefusalInvalidConnector, "effects", "import or reconcile must be enabled for lifecycle snapshots")
	}
	var users []domain.LDAPOutboundUser
	_, err := c.custody.Use(ctx, domain.UseSecretRequest{
		ReferenceID: connector.BindSecretReferenceID,
		AccountID:   connector.AccountID,
		WorkspaceID: connector.WorkspaceID,
		Purpose:     domain.SecretPurposeLDAPBind,
		OperationID: operationID,
	}, func(bindSecret []byte) error {
		var snapshotErr error
		users, snapshotErr = snapshotWithBind(ctx, connector, bindSecret)
		return snapshotErr
	})
	if err != nil {
		var ldapErr *domain.LDAPOutboundError
		if errors.As(err, &ldapErr) {
			return nil, ldapErr
		}
		return nil, err
	}
	return users, nil
}

func (c *Client) PlanLifecycle(ctx context.Context, connector domain.LDAPOutboundConnector, target domain.LDAPLifecycleTarget, operationID string, now time.Time) (domain.LDAPLifecyclePlan, error) {
	if err := connector.Validate(); err != nil {
		return domain.LDAPLifecyclePlan{}, err
	}
	if target == nil {
		return domain.LDAPLifecyclePlan{}, safeError(domain.LDAPRefusalInvalidConnector, "target", "a lifecycle target is required")
	}
	targetSnapshot, err := target.Snapshot(ctx, connector.AccountID, connector.WorkspaceID, connector.ID)
	if err != nil {
		return domain.LDAPLifecyclePlan{}, safeTargetError(err)
	}
	directoryUsers, err := c.Snapshot(ctx, connector, operationID)
	if err != nil {
		return domain.LDAPLifecyclePlan{}, err
	}
	return domain.BuildLDAPLifecyclePlan(connector, targetSnapshot, directoryUsers, now)
}

func (c *Client) ApplyLifecycle(ctx context.Context, connector domain.LDAPOutboundConnector, target domain.LDAPLifecycleTarget, plan domain.LDAPLifecyclePlan, now time.Time) (domain.LDAPLifecycleApplyResult, error) {
	if target == nil {
		return domain.LDAPLifecycleApplyResult{}, safeError(domain.LDAPRefusalInvalidConnector, "target", "a lifecycle target is required")
	}
	if err := connector.Validate(); err != nil {
		return domain.LDAPLifecycleApplyResult{}, err
	}
	if err := plan.Validate(connector, now); err != nil {
		return domain.LDAPLifecycleApplyResult{}, err
	}
	result, err := target.Apply(ctx, plan)
	if err != nil {
		return domain.LDAPLifecycleApplyResult{}, safeTargetError(err)
	}
	return result, nil
}

func authenticateWithBind(ctx context.Context, connector domain.LDAPOutboundConnector, login string, password, bindSecret []byte) (domain.LDAPOutboundUser, error) {
	conn, err := connectAndBind(ctx, connector, bindSecret)
	if err != nil {
		return domain.LDAPOutboundUser{}, err
	}
	defer conn.Close()
	entry, err := findUser(conn, connector, login)
	if err != nil {
		return domain.LDAPOutboundUser{}, err
	}
	if connector.DisabledUsers.Disabled(entry.GetAttributeValue(connector.DisabledUsers.Attribute)) {
		return domain.LDAPOutboundUser{}, safeError(domain.LDAPRefusalUserDisabled, "login", "the directory account is disabled")
	}
	user, err := mapUser(conn, connector, entry)
	if err != nil {
		return domain.LDAPOutboundUser{}, err
	}
	if err := ctx.Err(); err != nil {
		return domain.LDAPOutboundUser{}, safeError(domain.LDAPRefusalUnavailable, "context", "the directory operation was canceled")
	}
	if err := conn.Bind(entry.DN, string(password)); err != nil {
		return domain.LDAPOutboundUser{}, safeError(domain.LDAPRefusalInvalidCredential, "credentials", "the supplied credentials were not accepted")
	}
	return user, nil
}

func snapshotWithBind(ctx context.Context, connector domain.LDAPOutboundConnector, bindSecret []byte) ([]domain.LDAPOutboundUser, error) {
	conn, err := connectAndBind(ctx, connector, bindSecret)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	filter := "(&" + objectClassFilter(connector.UserObjectClasses) + ")"
	paging := ldaplib.NewControlPaging(connector.LifecyclePageSize)
	users := make([]domain.LDAPOutboundUser, 0)
	for {
		if err := ctx.Err(); err != nil {
			return nil, safeError(domain.LDAPRefusalUnavailable, "context", "the directory lifecycle snapshot was canceled")
		}
		request := ldaplib.NewSearchRequest(
			connector.UserBaseDN,
			ldaplib.ScopeWholeSubtree,
			ldaplib.NeverDerefAliases,
			0,
			int(connector.SearchTimeout/time.Second),
			false,
			filter,
			connector.RequestedUserAttributes(),
			[]ldaplib.Control{paging},
		)
		result, err := conn.Search(request)
		if err != nil {
			return nil, safeError(domain.LDAPRefusalSearch, "lifecycle_search", "the directory lifecycle search failed")
		}
		if len(result.Entries) > int(connector.LifecyclePageSize) {
			return nil, safeError(domain.LDAPRefusalResultLimit, "lifecycle_page_size", "the directory returned an oversized lifecycle page")
		}
		for _, entry := range result.Entries {
			if uint32(len(users)) >= connector.MaxSyncUsers {
				return nil, safeError(domain.LDAPRefusalResultLimit, "max_sync_users", "the complete directory snapshot exceeded its configured user limit")
			}
			user, err := mapUser(conn, connector, entry)
			if err != nil {
				return nil, err
			}
			users = append(users, user)
		}
		control, _ := ldaplib.FindControl(result.Controls, ldaplib.ControlTypePaging).(*ldaplib.ControlPaging)
		if control == nil || len(control.Cookie) == 0 {
			break
		}
		if uint32(len(users)) >= connector.MaxSyncUsers {
			return nil, safeError(domain.LDAPRefusalResultLimit, "max_sync_users", "the complete directory snapshot exceeded its configured user limit")
		}
		paging.SetCookie(control.Cookie)
	}
	sort.Slice(users, func(i, j int) bool {
		if users[i].Subject != users[j].Subject {
			return users[i].Subject < users[j].Subject
		}
		return users[i].Username < users[j].Username
	})
	return users, nil
}

func connectAndBind(ctx context.Context, connector domain.LDAPOutboundConnector, bindSecret []byte) (*ldaplib.Conn, error) {
	var sawTLSFailure bool
	for _, endpoint := range connector.Endpoints {
		if err := ctx.Err(); err != nil {
			return nil, safeError(domain.LDAPRefusalUnavailable, "context", "the directory operation was canceled")
		}
		conn, tlsFailure, err := dial(connector, endpoint)
		if err != nil {
			sawTLSFailure = sawTLSFailure || tlsFailure
			continue
		}
		if err := conn.Bind(connector.BindDN, string(bindSecret)); err != nil {
			conn.Close()
			return nil, safeError(domain.LDAPRefusalBind, "bind_dn", "the directory service bind was refused")
		}
		return conn, nil
	}
	if sawTLSFailure {
		return nil, safeError(domain.LDAPRefusalTLS, "endpoints", "no endpoint passed certificate and hostname validation")
	}
	return nil, safeError(domain.LDAPRefusalUnavailable, "endpoints", "no configured directory endpoint was reachable")
}

func dial(connector domain.LDAPOutboundConnector, endpoint string) (*ldaplib.Conn, bool, error) {
	parsed, _ := url.Parse(endpoint) // connector validation already proved it
	roots, err := x509.SystemCertPool()
	if err != nil || roots == nil {
		roots = x509.NewCertPool()
	}
	if connector.TrustAnchorsPEM != "" && !roots.AppendCertsFromPEM([]byte(connector.TrustAnchorsPEM)) {
		return nil, true, errors.New("invalid trust anchors")
	}
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		RootCAs:    roots,
		ServerName: parsed.Hostname(),
	}
	dialer := &net.Dialer{Timeout: connector.ConnectTimeout}
	opts := []ldaplib.DialOpt{ldaplib.DialWithDialer(dialer)}
	if parsed.Scheme == "ldaps" {
		opts = append(opts, ldaplib.DialWithTLSConfig(tlsConfig))
	}
	conn, err := ldaplib.DialURL(endpoint, opts...)
	if err != nil {
		return nil, isTLSFailure(err), err
	}
	conn.SetTimeout(connector.SearchTimeout)
	if parsed.Scheme == "ldap" {
		if err := conn.StartTLS(tlsConfig); err != nil {
			conn.Close()
			return nil, true, err
		}
	}
	return conn, false, nil
}

func findUser(conn *ldaplib.Conn, connector domain.LDAPOutboundConnector, login string) (*ldaplib.Entry, error) {
	request := ldaplib.NewSearchRequest(
		connector.UserBaseDN,
		ldaplib.ScopeWholeSubtree,
		ldaplib.NeverDerefAliases,
		int(connector.ResultLimit),
		int(connector.SearchTimeout/time.Second),
		false,
		connector.UserSearchFilter(login),
		connector.RequestedUserAttributes(),
		nil,
	)
	result, err := conn.Search(request)
	if err != nil {
		return nil, safeError(domain.LDAPRefusalSearch, "user_search", "the directory user search failed")
	}
	switch len(result.Entries) {
	case 0:
		return nil, safeError(domain.LDAPRefusalUserNotFound, "login", "the directory user was not found")
	case 1:
		return result.Entries[0], nil
	default:
		return nil, safeError(domain.LDAPRefusalUserAmbiguous, "login", "the login matched more than one directory user")
	}
}

func mapUser(conn *ldaplib.Conn, connector domain.LDAPOutboundConnector, entry *ldaplib.Entry) (domain.LDAPOutboundUser, error) {
	groups, err := resolveGroups(conn, connector, entry.DN)
	if err != nil {
		return domain.LDAPOutboundUser{}, err
	}
	mapping := connector.Attributes
	user := domain.LDAPOutboundUser{
		DN:          entry.DN,
		Subject:     entry.GetAttributeValue(mapping.Subject),
		Username:    entry.GetAttributeValue(mapping.Username),
		FirstName:   entry.GetAttributeValue(mapping.FirstName),
		LastName:    entry.GetAttributeValue(mapping.LastName),
		DisplayName: entry.GetAttributeValue(mapping.DisplayName),
		Email:       entry.GetAttributeValue(mapping.Email),
		Phone:       entry.GetAttributeValue(mapping.Phone),
		Groups:      groups,
		Disabled:    connector.DisabledUsers.Disabled(entry.GetAttributeValue(connector.DisabledUsers.Attribute)),
	}
	if strings.TrimSpace(user.Subject) == "" || strings.TrimSpace(user.Username) == "" {
		return domain.LDAPOutboundUser{}, safeError(domain.LDAPRefusalInvalidMapping, "attributes", "the directory entry is missing a required identity mapping")
	}
	return user, nil
}

func resolveGroups(conn *ldaplib.Conn, connector domain.LDAPOutboundConnector, userDN string) ([]string, error) {
	if connector.GroupBaseDN == "" {
		return []string{}, nil
	}
	queue := []string{userDN}
	seenDN := map[string]struct{}{strings.ToLower(userDN): {}}
	groups := make(map[string]string)
	var discovered uint32
	maxDepth := uint32(1)
	if connector.Groups.Traversal == domain.LDAPGroupsNested {
		maxDepth = connector.Groups.MaxDepth
	}
	for depth := uint32(0); depth < maxDepth && len(queue) > 0; depth++ {
		current := queue
		queue = nil
		for _, memberDN := range current {
			filter := "(&" + objectClassFilter(connector.Groups.ObjectClasses) + "(" + connector.Groups.Member + "=" + ldaplib.EscapeFilter(memberDN) + "))"
			request := ldaplib.NewSearchRequest(connector.GroupBaseDN, ldaplib.ScopeWholeSubtree, ldaplib.NeverDerefAliases, int(connector.ResultLimit), int(connector.SearchTimeout/time.Second), false, filter, []string{connector.Groups.Name}, nil)
			result, err := conn.Search(request)
			if err != nil {
				return nil, safeError(domain.LDAPRefusalSearch, "group_search", "the directory group search failed")
			}
			for _, entry := range result.Entries {
				name := entry.GetAttributeValue(connector.Groups.Name)
				if strings.TrimSpace(name) == "" {
					return nil, safeError(domain.LDAPRefusalInvalidMapping, "groups.name", "a directory group is missing its mapped name")
				}
				key := strings.ToLower(entry.DN)
				if _, exists := seenDN[key]; exists {
					continue
				}
				if discovered >= connector.ResultLimit {
					return nil, safeError(domain.LDAPRefusalResultLimit, "groups", "the directory group result limit was exceeded")
				}
				discovered++
				seenDN[key] = struct{}{}
				groups[strings.ToLower(name)] = name
				queue = append(queue, entry.DN)
			}
		}
	}
	out := make([]string, 0, len(groups))
	for _, group := range groups {
		out = append(out, group)
	}
	sort.Strings(out)
	return out, nil
}

func objectClassFilter(classes []string) string {
	parts := make([]string, 0, len(classes))
	for _, class := range classes {
		parts = append(parts, "(objectClass="+ldaplib.EscapeFilter(class)+")")
	}
	return strings.Join(parts, "")
}

func isTLSFailure(err error) bool {
	var unknownAuthority x509.UnknownAuthorityError
	var hostname x509.HostnameError
	var certificateInvalid x509.CertificateInvalidError
	return errors.As(err, &unknownAuthority) || errors.As(err, &hostname) || errors.As(err, &certificateInvalid) || strings.Contains(strings.ToLower(err.Error()), "tls") || strings.Contains(strings.ToLower(err.Error()), "certificate")
}

func safeError(reason domain.LDAPOutboundRefusal, field, detail string) error {
	return &domain.LDAPOutboundError{Reason: reason, Field: field, Detail: detail}
}

func safeTargetError(err error) error {
	var refusal *domain.LDAPOutboundError
	if errors.As(err, &refusal) {
		return refusal
	}
	return safeError(domain.LDAPRefusalUnavailable, "target", "the lifecycle target operation failed")
}
