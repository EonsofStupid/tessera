package command

import (
	"context"

	"golang.org/x/text/language"

	"github.com/shippinAI/nomen/internal/api/authz"
	"github.com/shippinAI/nomen/internal/command/preparation"
	"github.com/shippinAI/nomen/internal/domain"
	"github.com/shippinAI/nomen/internal/eventstore"
	"github.com/shippinAI/nomen/internal/i18n"
	"github.com/shippinAI/nomen/internal/repository/restrictions"
	"github.com/shippinAI/nomen/internal/zerrors"
)

type SetRestrictions struct {
	DisallowPublicOrgRegistration *bool
	AllowedLanguages              []language.Tag
}

func (s *SetRestrictions) Validate(defaultLanguage language.Tag) error {
	if s == nil || (s.DisallowPublicOrgRegistration == nil && s.AllowedLanguages == nil) {
		return zerrors.ThrowInvalidArgument(nil, "COMMAND-oASwj", "Errors.Restrictions.NoneSpecified")
	}
	if s.AllowedLanguages != nil {
		if err := domain.LanguagesHaveDuplicates(s.AllowedLanguages); err != nil {
			return err
		}
		if err := domain.LanguagesAreSupported(i18n.SupportedLanguages(), s.AllowedLanguages...); err != nil {
			return err
		}
		if err := domain.LanguageIsAllowed(false, s.AllowedLanguages, defaultLanguage); err != nil {
			return zerrors.ThrowPreconditionFailedf(err, "COMMAND-L0m2u", "Errors.Restrictions.DefaultLanguageMustBeAllowed")
		}
	}
	return nil
}

// SetInstanceRestrictions creates new restrictions or updates existing restrictions.
func (c *Commands) SetInstanceRestrictions(
	ctx context.Context,
	setRestrictions *SetRestrictions,
) (*domain.ObjectDetails, error) {
	instanceId := authz.GetInstance(ctx).InstanceID()
	wm, err := c.getRestrictionsWriteModel(ctx, instanceId, instanceId)
	if err != nil {
		return nil, err
	}
	aggregateId := wm.AggregateID
	if aggregateId == "" {
		aggregateId, err = c.idGenerator.Next()
		if err != nil {
			return nil, err
		}
	}
	setCmd, err := c.SetRestrictionsCommand(restrictions.NewAggregate(aggregateId, instanceId, instanceId), wm, setRestrictions)()
	if err != nil {
		return nil, err
	}
	cmds, err := setCmd(ctx, nil)
	if err != nil {
		return nil, err
	}
	if len(cmds) > 0 {
		events, err := c.eventstore.Push(ctx, cmds...)
		if err != nil {
			return nil, err
		}
		err = AppendAndReduce(wm, events...)
		if err != nil {
			return nil, err
		}
	}
	return writeModelToObjectDetails(&wm.WriteModel), nil
}

func (c *Commands) getRestrictionsWriteModel(ctx context.Context, instanceId, resourceOwner string) (*restrictionsWriteModel, error) {
	wm := newRestrictionsWriteModel(instanceId, resourceOwner)
	return wm, c.eventstore.FilterToQueryReducer(ctx, wm)
}

func (c *Commands) SetRestrictionsCommand(a *restrictions.Aggregate, wm *restrictionsWriteModel, setRestrictions *SetRestrictions) preparation.Validation {
	return func() (preparation.CreateCommands, error) {
		return func(ctx context.Context, _ preparation.FilterToQueryReducer) ([]eventstore.Command, error) {
			if err := setRestrictions.Validate(authz.GetInstance(ctx).DefaultLanguage()); err != nil {
				return nil, err
			}
			changes := wm.NewChanges(setRestrictions)
			if len(changes) == 0 {
				return nil, nil
			}
			return []eventstore.Command{restrictions.NewSetEvent(
				eventstore.NewBaseEventForPush(
					ctx,
					&a.Aggregate,
					restrictions.SetEventType,
				),
				changes...,
			)}, nil
		}, nil
	}
}
