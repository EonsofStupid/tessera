package start

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/shippinAI/nomen/backend/v3/instrumentation/logging"
	"github.com/shippinAI/nomen/cmd/initialise"
	"github.com/shippinAI/nomen/cmd/key"
	"github.com/shippinAI/nomen/cmd/setup"
	"github.com/shippinAI/nomen/cmd/tls"
)

func NewStartFromInit(server chan<- *Server) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start-from-init",
		Short: "cold starts Nomen",
		Long: `cold starts Nomen.
First the minimum requirements to start Nomen are set up.
Second the initial events are created.
Last Nomen starts.

Requirements:
- postgreSQL`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer func() {
				logging.OnError(cmd.Context(), err).Error("nomen start-from-init command failed")
			}()

			err = tls.ModeFromFlag(cmd)
			if err != nil {
				return fmt.Errorf("invalid tlsMode: %w", err)
			}

			masterKey, err := key.MasterKey(cmd)
			if err != nil {
				return fmt.Errorf("no master key provided: %w", err)
			}

			initConfig, shutdown, err := initialise.NewConfig(cmd, viper.GetViper())
			if err != nil {
				return err
			}
			defer func() {
				err = errors.Join(err, shutdown(cmd.Context()))
			}()
			initCtx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			err = initialise.InitAll(initCtx, initConfig)
			if err != nil {
				return err
			}

			err = setup.BindInitProjections(cmd)
			if err != nil {
				return fmt.Errorf("unable to bind \"init-projections\" flag: %w", err)
			}

			setupConfig, _, err := setup.NewConfig(cmd, viper.GetViper())
			if err != nil {
				return err
			}

			setupSteps, err := setup.NewSteps(cmd.Context(), viper.New())
			if err != nil {
				return err
			}

			err = setup.Setup(cmd.Context(), setupConfig, setupSteps, masterKey)
			if err != nil {
				return err
			}

			startConfig, _, err := NewConfig(cmd, viper.GetViper())
			if err != nil {
				return err
			}

			return startNomen(cmd.Context(), startConfig, masterKey, server)
		},
	}

	startFlags(cmd)
	setup.Flags(cmd)

	return cmd
}
