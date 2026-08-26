// Package blueprint is how declared state reaches a database.
//
//	nomen blueprint validate --dir ./blueprints
//	nomen blueprint apply    --dir ./blueprints --instance <id>
//
// validate needs no database and no configuration — it is the CI check and the
// editor's loop. apply is one transaction per file: a file that fails on its
// last entry applies none of itself, proven in the storage tests against a
// real Postgres.
package blueprint

import (
	"context"
	"errors"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	old_logging "github.com/shippinAI/nomen/logging" //nolint:staticcheck

	seatdomain "github.com/shippinAI/nomen/backend/v1/domain"
	blueprintstorage "github.com/shippinAI/nomen/backend/v1/storage/blueprint"
	flowstorage "github.com/shippinAI/nomen/backend/v1/storage/flow"
	nomenmigration "github.com/shippinAI/nomen/backend/v1/storage/migration"
	seatstorage "github.com/shippinAI/nomen/backend/v1/storage/seat"
	v3db "github.com/shippinAI/nomen/backend/v3/storage/database"
	v3_postgres "github.com/shippinAI/nomen/backend/v3/storage/database/dialect/postgres"
	"github.com/shippinAI/nomen/internal/database"
)

// engine builds the one registry there is. A new model means a new applier
// registered here, and nowhere else has an opinion.
func engine() *seatdomain.BlueprintEngine {
	return seatdomain.NewBlueprintEngine(
		seatstorage.NewApplier(),
		flowstorage.NewApplier(),
	)
}

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "blueprint",
		Short: "Validate and apply declared identity state",
	}
	cmd.AddCommand(newValidateCmd(), newApplyCmd())
	return cmd
}

func newValidateCmd() *cobra.Command {
	var dir string
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Parse and check a blueprint directory — no database, no writes",
		RunE: func(cmd *cobra.Command, _ []string) error {
			files, err := blueprintstorage.Load(dir)
			if err != nil {
				return err
			}
			// Load validated structure; the engine adds the one question
			// structure cannot answer — does every model have an applier.
			eng := engine()
			var errs []error
			for _, f := range files {
				if err := eng.Check(f.Blueprint); err != nil {
					errs = append(errs, fmt.Errorf("%s:\n%w", f.Path, err))
				}
			}
			if err := errors.Join(errs...); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%d file(s) valid\n", len(files))
			return nil
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "./blueprints", "directory of *.yaml blueprints")
	return cmd
}

type config struct {
	Database database.Config
	Log      *old_logging.Config
}

func newApplyCmd() *cobra.Command {
	var dir, instance string
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Make a blueprint directory true for one instance",
		Long: "Files apply in name order, each in its own transaction — a failing file " +
			"applies none of itself, and files before it stay applied. Re-applying a " +
			"converged directory reports every entry unchanged.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if instance == "" {
				return errors.New("--instance is required — a blueprint declares state for one tenant at a time")
			}
			files, err := blueprintstorage.Load(dir)
			if err != nil {
				return err
			}

			cfg := new(config)
			if err := viper.Unmarshal(cfg, viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
				database.DecodeHook(false),
				mapstructure.StringToTimeDurationHookFunc(),
				mapstructure.TextUnmarshallerHookFunc(),
			))); err != nil {
				return fmt.Errorf("unable to read config: %w", err)
			}
			db, err := database.Connect(cfg.Database, false)
			if err != nil {
				return fmt.Errorf("cannot connect: %w", err)
			}
			defer func() { _ = db.Close() }()
			if err := nomenmigration.Migrate(cmd.Context(), db.Pool); err != nil {
				return fmt.Errorf("cannot migrate the nomen schema: %w", err)
			}

			return Apply(cmd.Context(), v3_postgres.PGxPool(db.Pool), instance, files, func(format string, a ...any) {
				fmt.Fprintf(cmd.OutOrStdout(), format, a...)
			})
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "./blueprints", "directory of *.yaml blueprints")
	cmd.Flags().StringVar(&instance, "instance", "", "instance id to apply to")
	return cmd
}

// Apply runs loaded files in order and narrates one line per file. Exported
// because startup apply (3.4b) is the same loop with a different reporter, and
// two loops would drift.
func Apply(ctx context.Context, db v3db.Beginner, instance string, files []blueprintstorage.File, printf func(string, ...any)) error {
	eng := engine()
	changed := false
	for _, f := range files {
		report, err := eng.Apply(ctx, db, instance, f.Blueprint)
		if err != nil {
			return fmt.Errorf("%s: %w", f.Path, err)
		}
		counts := report.Counts()
		printf("%s: %d created, %d updated, %d removed, %d unchanged\n",
			f.Path, counts[seatdomain.OutcomeCreated], counts[seatdomain.OutcomeUpdated],
			counts[seatdomain.OutcomeRemoved], counts[seatdomain.OutcomeUnchanged])
		changed = changed || report.Changed()
	}
	if !changed {
		printf("converged: nothing to change\n")
	}
	return nil
}

// StartupConfig is the `Blueprints:` block of the start configuration.
type StartupConfig struct {
	// Dir is a blueprint directory applied to every instance on every start.
	// Empty means the feature is off — the only default that cannot surprise.
	Dir string
}

// ApplyOnStart applies a configured directory to every instance in the
// database. It is the same loop the CLI runs — one transaction per file,
// advisory-locked per instance — with the instance list coming from the
// database instead of a flag, because on boot nobody is there to type one.
func ApplyOnStart(ctx context.Context, cfg StartupConfig, db v3db.Pool) error {
	if cfg.Dir == "" {
		return nil
	}
	files, err := blueprintstorage.Load(cfg.Dir)
	if err != nil {
		return err
	}

	rows, err := db.Query(ctx, "SELECT id FROM nomen.instances ORDER BY id")
	if err != nil {
		return fmt.Errorf("listing instances: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var instances []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return err
		}
		instances = append(instances, id)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, instance := range instances {
		if err := Apply(ctx, db, instance, files, func(format string, a ...any) {
			fmt.Printf("blueprint[%s] "+format, append([]any{instance}, a...)...)
		}); err != nil {
			return fmt.Errorf("instance %s: %w", instance, err)
		}
	}
	return nil
}
