// Package seat is the operator's view of who occupies what.
//
// It exists because Tessera is run *for* customers rather than by them: nobody
// on the far side of this ever sees a seat, so the people who do need a way to
// read and correct one that does not involve a SQL client. Blueprints will be
// how seats are normally authored — declared in reviewed files rather than typed
// — and they drive the same repository this does, so the two cannot disagree
// about what a seat is.
package seat

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	old_logging "github.com/zitadel/logging" //nolint:staticcheck

	seatdomain "github.com/EonsofStupid/tessera/backend/v1/domain"
	tesseramigration "github.com/EonsofStupid/tessera/backend/v1/storage/migration"
	seatstorage "github.com/EonsofStupid/tessera/backend/v1/storage/seat"
	v3_postgres "github.com/EonsofStupid/tessera/backend/v3/storage/database/dialect/postgres"
	"github.com/EonsofStupid/tessera/internal/database"
)

type config struct {
	Database database.Config
	Log      *old_logging.Config
}

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "seat",
		Short: "Inspect and set seats",
		Long: "Seats are normally authored as blueprints. These commands are for " +
			"reading one and for correcting one when a person has to.",
	}
	cmd.AddCommand(newSetCmd(), newShowCmd(), newListCmd())
	return cmd
}

// open connects and makes sure our schema is current, so a fresh box does not
// fail with a missing table when the operator is already mid-incident.
func open(ctx context.Context) (*seatstorage.Repository, func(), error) {
	cfg := new(config)
	if err := viper.Unmarshal(cfg, viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		database.DecodeHook(false),
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.TextUnmarshallerHookFunc(),
	))); err != nil {
		return nil, nil, fmt.Errorf("unable to read config: %w", err)
	}
	db, err := database.Connect(cfg.Database, false)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot connect: %w", err)
	}
	if err := tesseramigration.Migrate(ctx, db.Pool); err != nil {
		return nil, nil, fmt.Errorf("cannot migrate the tessera schema: %w", err)
	}
	return seatstorage.NewRepository(v3_postgres.PGxPool(db.Pool)), func() { _ = db.Close() }, nil
}

func newSetCmd() *cobra.Command {
	var (
		instance, member, account, occupant, basis, policyVersion string
		workspaces, scopes                                        []string
	)
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Create or replace a seat",
		Long: "The whole seat, not a patch — the flags you leave out are set to their " +
			"defaults rather than left as they were, so running this twice lands in the " +
			"same place. Workspaces are replaced, so dropping one revokes it.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if instance == "" || member == "" {
				return errors.New("--instance and --member are required")
			}
			repo, closeFn, err := open(cmd.Context())
			if err != nil {
				return err
			}
			defer closeFn()

			s := &seatdomain.Seat{
				MemberID:      member,
				AccountID:     account,
				Occupant:      seatdomain.ParseOccupant(occupant),
				Basis:         seatdomain.ParseBasis(basis),
				Workspaces:    workspaces,
				Scopes:        scopes,
				PolicyVersion: policyVersion,
			}
			if err := repo.SetSeat(cmd.Context(), instance, s); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "seat %s: %s/%s in %s\n", member, s.Occupant, s.Basis, strings.Join(workspaces, " "))
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVar(&instance, "instance", "", "instance id (the tenant root)")
	f.StringVar(&member, "member", "", "member id — the token's subject")
	f.StringVar(&account, "account", "", "account id — the tenant")
	f.StringVar(&occupant, "occupant", "agent", "human | agent")
	f.StringVar(&basis, "basis", "unknown", "subscription | usage | local | unknown")
	f.StringSliceVar(&workspaces, "workspaces", nil, "workspaces this seat may occupy")
	f.StringSliceVar(&scopes, "scopes", nil, "entitlement scopes")
	f.StringVar(&policyVersion, "policy-version", "", "the policy the scopes cite")
	return cmd
}

func newShowCmd() *cobra.Command {
	var instance, member string
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show one seat",
		RunE: func(cmd *cobra.Command, _ []string) error {
			repo, closeFn, err := open(cmd.Context())
			if err != nil {
				return err
			}
			defer closeFn()
			s, err := repo.SeatByMember(cmd.Context(), instance, member)
			if err != nil {
				return err
			}
			if len(s.Workspaces) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "%s occupies nothing — no seat is recorded\n", member)
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "member     %s\naccount    %s\noccupant   %s\nbasis      %s\nworkspaces %s\nscopes     %s\npolicy     %s\n",
				s.MemberID, s.AccountID, s.Occupant, s.Basis,
				strings.Join(s.Workspaces, " "), strings.Join(s.Scopes, " "), s.PolicyVersion)
			return nil
		},
	}
	cmd.Flags().StringVar(&instance, "instance", "", "instance id")
	cmd.Flags().StringVar(&member, "member", "", "member id")
	return cmd
}

func newListCmd() *cobra.Command {
	var instance, workspace string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the seats occupying a workspace",
		RunE: func(cmd *cobra.Command, _ []string) error {
			repo, closeFn, err := open(cmd.Context())
			if err != nil {
				return err
			}
			defer closeFn()
			seats, err := repo.SeatsInWorkspace(cmd.Context(), instance, workspace)
			if err != nil {
				return err
			}
			if len(seats) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "nobody occupies %s\n", workspace)
				return nil
			}
			for _, s := range seats {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s/%s\t%s\n", s.MemberID, s.Occupant, s.Basis, strings.Join(s.Scopes, ","))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&instance, "instance", "", "instance id")
	cmd.Flags().StringVar(&workspace, "workspace", "", "workspace id")
	return cmd
}
