package command

import (
	"context"

	"github.com/shippinAI/nomen/internal/api/authz"
	"github.com/shippinAI/nomen/internal/command/preparation"
	"github.com/shippinAI/nomen/internal/domain"
	"github.com/shippinAI/nomen/internal/eventstore"
	"github.com/shippinAI/nomen/internal/feature"
	"github.com/shippinAI/nomen/internal/repository/feature/feature_v2"
	"github.com/shippinAI/nomen/internal/zerrors"
)

type InstanceFeatures struct {
	LoginDefaultOrg                *bool
	UserSchema                     *bool
	ImprovedPerformance            []feature.ImprovedPerformanceType
	DebugOIDCParentError           *bool
	OIDCSingleV1SessionTermination *bool
	LoginV2                        *feature.LoginV2
	PermissionCheckV2              *bool
	ManagementConsoleUseV2UserApi  *bool `mapstructure:"ConsoleUseV2UserApi"` // for backwards compatibility we need to change this back to the old config name
	EnableRelationalTables         *bool
}

func (m *InstanceFeatures) isEmpty() bool {
	return m == nil || (m.LoginDefaultOrg == nil &&
		m.UserSchema == nil &&
		// nil check to allow unset improvements
		m.ImprovedPerformance == nil &&
		m.DebugOIDCParentError == nil &&
		m.OIDCSingleV1SessionTermination == nil &&
		m.LoginV2 == nil &&
		m.PermissionCheckV2 == nil &&
		m.ManagementConsoleUseV2UserApi == nil &&
		m.EnableRelationalTables == nil)
}

func (c *Commands) SetInstanceFeatures(ctx context.Context, f *InstanceFeatures) (*domain.ObjectDetails, error) {
	if f.isEmpty() {
		return nil, zerrors.ThrowInvalidArgument(nil, "COMMAND-Vigh1", "Errors.NoChangesFound")
	}
	wm := NewInstanceFeaturesWriteModel(authz.GetInstance(ctx).InstanceID())
	if err := c.eventstore.FilterToQueryReducer(ctx, wm); err != nil {
		return nil, err
	}
	commands := wm.setCommands(ctx, f)
	if len(commands) == 0 {
		return writeModelToObjectDetails(wm.WriteModel), nil
	}
	events, err := c.eventstore.Push(ctx, commands...)
	if err != nil {
		return nil, err
	}
	return pushedEventsToObjectDetails(events), nil
}

func prepareSetFeatures(instanceID string, f *InstanceFeatures) preparation.Validation {
	return func() (preparation.CreateCommands, error) {
		return func(ctx context.Context, _ preparation.FilterToQueryReducer) ([]eventstore.Command, error) {
			wm := NewInstanceFeaturesWriteModel(instanceID)
			return wm.setCommands(ctx, f), nil
		}, nil
	}
}

func (c *Commands) ResetInstanceFeatures(ctx context.Context) (*domain.ObjectDetails, error) {
	instanceID := authz.GetInstance(ctx).InstanceID()
	wm := NewInstanceFeaturesWriteModel(instanceID)
	if err := c.eventstore.FilterToQueryReducer(ctx, wm); err != nil {
		return nil, err
	}
	if wm.isEmpty() {
		return writeModelToObjectDetails(wm.WriteModel), nil
	}
	aggregate := feature_v2.NewAggregate(instanceID, instanceID)
	events, err := c.eventstore.Push(ctx, feature_v2.NewResetEvent(ctx, aggregate, feature_v2.InstanceResetEventType))
	if err != nil {
		return nil, err
	}
	return pushedEventsToObjectDetails(events), nil
}
