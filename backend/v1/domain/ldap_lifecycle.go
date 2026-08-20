package domain

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
	"time"
)

const LDAPLifecyclePlanTTL = 10 * time.Minute

type LDAPLifecycleActionType string

const (
	LDAPLifecycleCreate     LDAPLifecycleActionType = "create"
	LDAPLifecycleUpdate     LDAPLifecycleActionType = "update"
	LDAPLifecycleReactivate LDAPLifecycleActionType = "reactivate"
	LDAPLifecycleSuspend    LDAPLifecycleActionType = "suspend"
)

type LDAPManagedUser struct {
	TargetUserID string           `json:"target_user_id"`
	AccountID    string           `json:"account_id"`
	WorkspaceID  string           `json:"workspace_id"`
	ConnectorID  string           `json:"connector_id"`
	User         LDAPOutboundUser `json:"user"`
	Suspended    bool             `json:"suspended"`
}

type LDAPLifecycleAction struct {
	Type         LDAPLifecycleActionType `json:"type"`
	TargetUserID string                  `json:"target_user_id,omitempty"`
	User         LDAPOutboundUser        `json:"user"`
}

type LDAPLifecyclePlan struct {
	PlanHash               string                `json:"plan_hash"`
	ConnectorID            string                `json:"connector_id"`
	ConnectorRevision      string                `json:"connector_revision"`
	AccountID              string                `json:"account_id"`
	WorkspaceID            string                `json:"workspace_id"`
	ExpectedTargetRevision string                `json:"expected_target_revision"`
	GeneratedAt            time.Time             `json:"generated_at"`
	ExpiresAt              time.Time             `json:"expires_at"`
	Actions                []LDAPLifecycleAction `json:"actions"`
}

type LDAPTargetSnapshot struct {
	Revision string            `json:"revision"`
	Users    []LDAPManagedUser `json:"users"`
}

type LDAPLifecycleApplyResult struct {
	Revision    string `json:"revision"`
	Created     uint32 `json:"created"`
	Updated     uint32 `json:"updated"`
	Reactivated uint32 `json:"reactivated"`
	Suspended   uint32 `json:"suspended"`
	Replayed    bool   `json:"replayed"`
}

// LDAPLifecycleTarget owns the atomic user write boundary. Snapshot must return
// the complete tenant username namespace plus connector ownership metadata.
// Apply must compare ExpectedTargetRevision and make PlanHash idempotent.
type LDAPLifecycleTarget interface {
	Snapshot(context.Context, string, string, string) (LDAPTargetSnapshot, error)
	Apply(context.Context, LDAPLifecyclePlan) (LDAPLifecycleApplyResult, error)
}

func BuildLDAPLifecyclePlan(connector LDAPOutboundConnector, snapshot LDAPTargetSnapshot, source []LDAPOutboundUser, now time.Time) (LDAPLifecyclePlan, error) {
	if err := connector.Validate(); err != nil {
		return LDAPLifecyclePlan{}, err
	}
	if !connector.Effects.Import && !connector.Effects.Reconcile {
		return LDAPLifecyclePlan{}, ldapOutboundError(LDAPRefusalInvalidConnector, "effects", "import or reconcile must be enabled for lifecycle planning")
	}
	if len(source) > int(connector.MaxSyncUsers) {
		return LDAPLifecyclePlan{}, ldapOutboundError(LDAPRefusalResultLimit, "max_sync_users", "the complete directory snapshot exceeded its configured user limit")
	}

	sourceBySubject := make(map[string]LDAPOutboundUser, len(source))
	usernames := make(map[string]string, len(source))
	for _, user := range source {
		normalizeLDAPUser(&user)
		if user.Subject == "" || user.Username == "" {
			return LDAPLifecyclePlan{}, ldapOutboundError(LDAPRefusalInvalidMapping, "source", "a lifecycle source user is missing its subject or username")
		}
		if _, exists := sourceBySubject[user.Subject]; exists {
			return LDAPLifecyclePlan{}, ldapOutboundError(LDAPRefusalSyncConflict, "subject", "the directory snapshot contains a duplicate subject")
		}
		usernameKey := strings.ToLower(user.Username)
		if subject, exists := usernames[usernameKey]; exists && subject != user.Subject {
			return LDAPLifecyclePlan{}, ldapOutboundError(LDAPRefusalSyncConflict, "username", "the directory snapshot contains a duplicate username")
		}
		sourceBySubject[user.Subject] = user
		usernames[usernameKey] = user.Subject
	}

	existingBySubject := make(map[string]LDAPManagedUser, len(snapshot.Users))
	targetUsernames := make(map[string]string, len(snapshot.Users))
	for _, managed := range snapshot.Users {
		if managed.AccountID != connector.AccountID || managed.WorkspaceID != connector.WorkspaceID {
			return LDAPLifecyclePlan{}, ldapOutboundError(LDAPRefusalSyncConflict, "scope", "the target snapshot crossed its requested tenant boundary")
		}
		if managed.TargetUserID == "" || managed.User.Username == "" {
			return LDAPLifecyclePlan{}, ldapOutboundError(LDAPRefusalSyncConflict, "target", "a target user has no stable identity")
		}
		normalizeLDAPUser(&managed.User)
		usernameKey := strings.ToLower(managed.User.Username)
		if prior, exists := targetUsernames[usernameKey]; exists && prior != managed.TargetUserID {
			return LDAPLifecyclePlan{}, ldapOutboundError(LDAPRefusalSyncConflict, "target.username", "the target contains a duplicate username")
		}
		targetUsernames[usernameKey] = managed.TargetUserID
		if managed.ConnectorID != connector.ID {
			continue
		}
		if managed.User.Subject == "" {
			return LDAPLifecyclePlan{}, ldapOutboundError(LDAPRefusalSyncConflict, "target", "a connector-owned target user has no stable subject")
		}
		if _, exists := existingBySubject[managed.User.Subject]; exists {
			return LDAPLifecyclePlan{}, ldapOutboundError(LDAPRefusalSyncConflict, "target", "the target contains a duplicate connector subject")
		}
		existingBySubject[managed.User.Subject] = managed
	}
	for subject, user := range sourceBySubject {
		allowedTargetID := ""
		if managed, exists := existingBySubject[subject]; exists {
			allowedTargetID = managed.TargetUserID
		}
		if owner, exists := targetUsernames[strings.ToLower(user.Username)]; exists && owner != allowedTargetID {
			return LDAPLifecyclePlan{}, ldapOutboundError(LDAPRefusalSyncConflict, "username", "a directory username collides with another target identity")
		}
	}

	actions := make([]LDAPLifecycleAction, 0)
	for subject, user := range sourceBySubject {
		managed, exists := existingBySubject[subject]
		if !exists {
			if connector.Effects.Import {
				actions = append(actions, LDAPLifecycleAction{Type: LDAPLifecycleCreate, User: user})
			}
			continue
		}
		if connector.Effects.Reconcile && !sameLDAPUser(managed.User, user) {
			actions = append(actions, LDAPLifecycleAction{Type: LDAPLifecycleUpdate, TargetUserID: managed.TargetUserID, User: user})
		}
		if !connector.Effects.Reconcile {
			continue
		}
		if user.Disabled && !managed.Suspended {
			actions = append(actions, LDAPLifecycleAction{Type: LDAPLifecycleSuspend, TargetUserID: managed.TargetUserID, User: user})
		} else if !user.Disabled && managed.Suspended {
			actions = append(actions, LDAPLifecycleAction{Type: LDAPLifecycleReactivate, TargetUserID: managed.TargetUserID, User: user})
		}
	}
	if connector.Effects.Deprovision {
		for subject, managed := range existingBySubject {
			if _, present := sourceBySubject[subject]; !present && !managed.Suspended {
				actions = append(actions, LDAPLifecycleAction{Type: LDAPLifecycleSuspend, TargetUserID: managed.TargetUserID, User: managed.User})
			}
		}
	}
	sort.Slice(actions, func(i, j int) bool {
		if actions[i].User.Subject != actions[j].User.Subject {
			return actions[i].User.Subject < actions[j].User.Subject
		}
		return lifecycleActionOrder(actions[i].Type) < lifecycleActionOrder(actions[j].Type)
	})

	generated := now.UTC().Truncate(time.Second)
	plan := LDAPLifecyclePlan{
		ConnectorID: connector.ID, ConnectorRevision: connector.ResourceRevision,
		AccountID: connector.AccountID, WorkspaceID: connector.WorkspaceID,
		ExpectedTargetRevision: snapshot.Revision, GeneratedAt: generated,
		ExpiresAt: generated.Add(LDAPLifecyclePlanTTL), Actions: actions,
	}
	plan.PlanHash = planHash(plan)
	return plan, nil
}

func (p LDAPLifecyclePlan) Validate(connector LDAPOutboundConnector, now time.Time) error {
	if p.ConnectorID != connector.ID || p.ConnectorRevision != connector.ResourceRevision || p.AccountID != connector.AccountID || p.WorkspaceID != connector.WorkspaceID {
		return ldapOutboundError(LDAPRefusalInvalidPlan, "scope", "the lifecycle plan does not match the connector revision and tenant")
	}
	if p.PlanHash == "" || p.PlanHash != planHash(p) {
		return ldapOutboundError(LDAPRefusalInvalidPlan, "plan_hash", "the lifecycle plan hash is invalid")
	}
	if !now.UTC().Before(p.ExpiresAt) || p.ExpiresAt.Sub(p.GeneratedAt) != LDAPLifecyclePlanTTL {
		return ldapOutboundError(LDAPRefusalInvalidPlan, "expires_at", "the lifecycle plan is expired or has an invalid lifetime")
	}
	for _, action := range p.Actions {
		if action.User.Subject == "" || action.User.Username == "" {
			return ldapOutboundError(LDAPRefusalInvalidPlan, "actions", "a lifecycle action is missing its stable identity")
		}
		switch action.Type {
		case LDAPLifecycleCreate:
			if action.TargetUserID != "" {
				return ldapOutboundError(LDAPRefusalInvalidPlan, "actions", "a create action cannot name an existing target user")
			}
		case LDAPLifecycleUpdate, LDAPLifecycleReactivate, LDAPLifecycleSuspend:
			if action.TargetUserID == "" {
				return ldapOutboundError(LDAPRefusalInvalidPlan, "actions", "a target user is required for this lifecycle action")
			}
		default:
			return ldapOutboundError(LDAPRefusalInvalidPlan, "actions", "the lifecycle plan contains an unknown action")
		}
	}
	return nil
}

func planHash(plan LDAPLifecyclePlan) string {
	plan.PlanHash = ""
	encoded, _ := json.Marshal(plan)
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}

func normalizeLDAPUser(user *LDAPOutboundUser) {
	user.Subject = strings.TrimSpace(user.Subject)
	user.Username = strings.TrimSpace(user.Username)
	groups := make([]string, 0, len(user.Groups))
	seen := make(map[string]struct{}, len(user.Groups))
	for _, group := range user.Groups {
		group = strings.TrimSpace(group)
		if group == "" {
			continue
		}
		key := strings.ToLower(group)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		groups = append(groups, group)
	}
	sort.Slice(groups, func(i, j int) bool { return strings.ToLower(groups[i]) < strings.ToLower(groups[j]) })
	user.Groups = groups
}

func sameLDAPUser(left, right LDAPOutboundUser) bool {
	normalizeLDAPUser(&left)
	normalizeLDAPUser(&right)
	return left.DN == right.DN && left.Subject == right.Subject && left.Username == right.Username && left.FirstName == right.FirstName && left.LastName == right.LastName && left.DisplayName == right.DisplayName && left.Email == right.Email && left.Phone == right.Phone && left.Disabled == right.Disabled && strings.Join(left.Groups, "\x00") == strings.Join(right.Groups, "\x00")
}

func lifecycleActionOrder(action LDAPLifecycleActionType) int {
	switch action {
	case LDAPLifecycleCreate:
		return 1
	case LDAPLifecycleUpdate:
		return 2
	case LDAPLifecycleReactivate:
		return 3
	case LDAPLifecycleSuspend:
		return 4
	default:
		return 99
	}
}
