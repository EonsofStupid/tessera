package admin

import (
	_ "embed"
	"errors"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/shippinAI/nomen/cmd/initialise"
	"github.com/shippinAI/nomen/cmd/key"
	"github.com/shippinAI/nomen/cmd/setup"
	"github.com/shippinAI/nomen/cmd/start"
)

func New() *cobra.Command {
	adminCMD := &cobra.Command{
		Use:        "admin",
		Short:      "The Nomen admin CLI lets you interact with your instance",
		Long:       `The Nomen admin CLI lets you interact with your instance`,
		Deprecated: "please use subcommands directly, e.g. `nomen start`",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer func() {
				if err != nil {
					slog.Error("nomen admin command failed", "err", err)
				}
			}()
			return errors.New("no additional command provided")
		},
	}

	adminCMD.AddCommand(
		initialise.New(),
		setup.New(),
		start.New(nil),
		start.NewStartFromInit(nil),
		key.New(),
	)

	return adminCMD
}
