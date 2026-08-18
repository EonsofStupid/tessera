package domain

import (
	"errors"
	"slices"
	"testing"
)

func TestValidateDeploymentTransitionMatrix(t *testing.T) {
	t.Parallel()
	allowedChanges := map[DeploymentState][]DeploymentState{
		DeploymentAbsent:           {DeploymentPreparing},
		DeploymentPreparing:        {DeploymentAbsent, DeploymentInitializing},
		DeploymentInitializing:     {DeploymentNeedsOwner, DeploymentDegraded, DeploymentRecoveryRequired},
		DeploymentNeedsOwner:       {DeploymentReady, DeploymentDegraded, DeploymentRecoveryRequired, DeploymentRetired},
		DeploymentReady:            {DeploymentDegraded, DeploymentMaintenance, DeploymentRecoveryRequired, DeploymentRetired},
		DeploymentDegraded:         {DeploymentReady, DeploymentMaintenance, DeploymentRecoveryRequired, DeploymentRetired},
		DeploymentMaintenance:      {DeploymentReady, DeploymentDegraded, DeploymentRecoveryRequired, DeploymentRetired},
		DeploymentRecoveryRequired: {DeploymentMaintenance, DeploymentRetired},
		DeploymentRetired:          {DeploymentAbsent},
	}

	for _, current := range DeploymentStates() {
		current := current
		for _, target := range DeploymentStates() {
			target := target
			t.Run(string(current)+"_to_"+string(target), func(t *testing.T) {
				t.Parallel()
				err := ValidateDeploymentTransition(current, target)
				wantAllowed := current == target || slices.Contains(allowedChanges[current], target)
				if wantAllowed {
					if err != nil {
						t.Fatalf("ValidateDeploymentTransition(%q, %q): %v", current, target, err)
					}
					return
				}
				var refusal *DeploymentTransitionError
				if !errors.As(err, &refusal) {
					t.Fatalf("ValidateDeploymentTransition(%q, %q) error = %T %v, want *DeploymentTransitionError", current, target, err, err)
				}
				if refusal.Reason != DeploymentRefusalTransitionNotAllowed || refusal.Current != current || refusal.Target != target {
					t.Fatalf("refusal = %#v, want transition_not_allowed from %q to %q", refusal, current, target)
				}
			})
		}
	}
}

func TestValidateDeploymentTransitionUnknownStates(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		current    DeploymentState
		target     DeploymentState
		wantReason DeploymentTransitionRefusal
	}{
		{name: "unknown current wins", current: "mystery", target: "also-mystery", wantReason: DeploymentRefusalUnknownCurrentState},
		{name: "unknown target", current: DeploymentReady, target: "mystery", wantReason: DeploymentRefusalUnknownTargetState},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			var refusal *DeploymentTransitionError
			if err := ValidateDeploymentTransition(test.current, test.target); !errors.As(err, &refusal) {
				t.Fatalf("ValidateDeploymentTransition(%q, %q) error = %T %v, want *DeploymentTransitionError", test.current, test.target, err, err)
			}
			if refusal.Reason != test.wantReason || refusal.Current != test.current || refusal.Target != test.target {
				t.Fatalf("refusal = %#v, want reason %q from %q to %q", refusal, test.wantReason, test.current, test.target)
			}
		})
	}
}

func TestAllowedDeploymentTransitionsAgreesWithValidation(t *testing.T) {
	t.Parallel()
	for _, current := range DeploymentStates() {
		current := current
		t.Run(string(current), func(t *testing.T) {
			t.Parallel()
			allowed, err := AllowedDeploymentTransitions(current)
			if err != nil {
				t.Fatalf("AllowedDeploymentTransitions(%q): %v", current, err)
			}
			if len(allowed) == 0 || allowed[0] != current {
				t.Fatalf("AllowedDeploymentTransitions(%q) = %v, want current state first", current, allowed)
			}
			for _, target := range DeploymentStates() {
				wantAllowed := ValidateDeploymentTransition(current, target) == nil
				if gotAllowed := slices.Contains(allowed, target); gotAllowed != wantAllowed {
					t.Fatalf("AllowedDeploymentTransitions(%q) contains %q = %v, validation allowed = %v", current, target, gotAllowed, wantAllowed)
				}
			}
		})
	}
}

func TestAllowedDeploymentTransitionsRejectsUnknownCurrentState(t *testing.T) {
	t.Parallel()
	allowed, err := AllowedDeploymentTransitions("mystery")
	if allowed != nil {
		t.Fatalf("AllowedDeploymentTransitions(mystery) = %v, want nil", allowed)
	}
	var refusal *DeploymentTransitionError
	if !errors.As(err, &refusal) || refusal.Reason != DeploymentRefusalUnknownCurrentState {
		t.Fatalf("AllowedDeploymentTransitions(mystery) error = %T %v, want unknown_current_state", err, err)
	}
}

func TestDeploymentStatesReturnsCopy(t *testing.T) {
	t.Parallel()
	states := DeploymentStates()
	states[0] = "changed"
	if DeploymentStates()[0] != DeploymentAbsent {
		t.Fatal("DeploymentStates returned mutable package state")
	}
}
