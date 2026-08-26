package domain

import "fmt"

// DeploymentState is the server-owned operational state of one Nomen
// deployment. It does not describe a user, session, federation provider or
// commercial plan.
type DeploymentState string

const (
	DeploymentAbsent           DeploymentState = "absent"
	DeploymentPreparing        DeploymentState = "preparing"
	DeploymentInitializing     DeploymentState = "initializing"
	DeploymentNeedsOwner       DeploymentState = "needs_owner"
	DeploymentReady            DeploymentState = "ready"
	DeploymentDegraded         DeploymentState = "degraded"
	DeploymentMaintenance      DeploymentState = "maintenance"
	DeploymentRecoveryRequired DeploymentState = "recovery_required"
	DeploymentRetired          DeploymentState = "retired"
)

var deploymentStateOrder = []DeploymentState{
	DeploymentAbsent,
	DeploymentPreparing,
	DeploymentInitializing,
	DeploymentNeedsOwner,
	DeploymentReady,
	DeploymentDegraded,
	DeploymentMaintenance,
	DeploymentRecoveryRequired,
	DeploymentRetired,
}

var deploymentTransitions = map[DeploymentState]map[DeploymentState]struct{}{
	DeploymentAbsent: {
		DeploymentPreparing: {},
	},
	DeploymentPreparing: {
		DeploymentAbsent:       {},
		DeploymentInitializing: {},
	},
	DeploymentInitializing: {
		DeploymentNeedsOwner:       {},
		DeploymentDegraded:         {},
		DeploymentRecoveryRequired: {},
	},
	DeploymentNeedsOwner: {
		DeploymentReady:            {},
		DeploymentDegraded:         {},
		DeploymentRecoveryRequired: {},
		DeploymentRetired:          {},
	},
	DeploymentReady: {
		DeploymentDegraded:         {},
		DeploymentMaintenance:      {},
		DeploymentRecoveryRequired: {},
		DeploymentRetired:          {},
	},
	DeploymentDegraded: {
		DeploymentReady:            {},
		DeploymentMaintenance:      {},
		DeploymentRecoveryRequired: {},
		DeploymentRetired:          {},
	},
	DeploymentMaintenance: {
		DeploymentReady:            {},
		DeploymentDegraded:         {},
		DeploymentRecoveryRequired: {},
		DeploymentRetired:          {},
	},
	DeploymentRecoveryRequired: {
		DeploymentMaintenance: {},
		DeploymentRetired:     {},
	},
	DeploymentRetired: {
		DeploymentAbsent: {},
	},
}

// DeploymentTransitionRefusal is a stable machine-readable reason. Transport
// adapters map it to their typed error contract; callers do not branch on the
// diagnostic string.
type DeploymentTransitionRefusal string

const (
	DeploymentRefusalUnknownCurrentState  DeploymentTransitionRefusal = "unknown_current_state"
	DeploymentRefusalUnknownTargetState   DeploymentTransitionRefusal = "unknown_target_state"
	DeploymentRefusalTransitionNotAllowed DeploymentTransitionRefusal = "transition_not_allowed"
)

// DeploymentTransitionError describes a refused lifecycle transition.
type DeploymentTransitionError struct {
	Reason  DeploymentTransitionRefusal
	Current DeploymentState
	Target  DeploymentState
}

func (e *DeploymentTransitionError) Error() string {
	switch e.Reason {
	case DeploymentRefusalUnknownCurrentState:
		return fmt.Sprintf("deployment lifecycle: unknown current state %q", e.Current)
	case DeploymentRefusalUnknownTargetState:
		return fmt.Sprintf("deployment lifecycle: unknown target state %q", e.Target)
	default:
		return fmt.Sprintf("deployment lifecycle: transition from %q to %q is not allowed", e.Current, e.Target)
	}
}

// DeploymentStates returns the complete state vocabulary in canonical order.
func DeploymentStates() []DeploymentState {
	return append([]DeploymentState(nil), deploymentStateOrder...)
}

// Valid reports whether state belongs to the deployment lifecycle contract.
func (s DeploymentState) Valid() bool {
	_, ok := deploymentTransitions[s]
	return ok
}

// ValidateDeploymentTransition accepts a declared graph edge or an idempotent
// transition to the same known state. It refuses unknown values before testing
// graph membership so corrupt stored state is distinct from a bad target.
func ValidateDeploymentTransition(current, target DeploymentState) error {
	if !current.Valid() {
		return &DeploymentTransitionError{
			Reason:  DeploymentRefusalUnknownCurrentState,
			Current: current,
			Target:  target,
		}
	}
	if !target.Valid() {
		return &DeploymentTransitionError{
			Reason:  DeploymentRefusalUnknownTargetState,
			Current: current,
			Target:  target,
		}
	}
	if current == target {
		return nil
	}
	if _, ok := deploymentTransitions[current][target]; ok {
		return nil
	}
	return &DeploymentTransitionError{
		Reason:  DeploymentRefusalTransitionNotAllowed,
		Current: current,
		Target:  target,
	}
}

// AllowedDeploymentTransitions returns the current state first (the
// idempotent achieved target), followed by state-changing targets in canonical
// order.
func AllowedDeploymentTransitions(current DeploymentState) ([]DeploymentState, error) {
	if !current.Valid() {
		return nil, &DeploymentTransitionError{
			Reason:  DeploymentRefusalUnknownCurrentState,
			Current: current,
		}
	}
	allowed := []DeploymentState{current}
	for _, target := range deploymentStateOrder {
		if target == current {
			continue
		}
		if _, ok := deploymentTransitions[current][target]; ok {
			allowed = append(allowed, target)
		}
	}
	return allowed, nil
}
