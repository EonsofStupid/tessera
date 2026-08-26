package start

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/shippinAI/nomen/backend/v3/instrumentation/logging"
	"github.com/shippinAI/nomen/cmd/key"
	"github.com/shippinAI/nomen/cmd/setup"
	"github.com/shippinAI/nomen/cmd/tls"
)

func NewStartFromSetup(server chan<- *Server) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start-from-setup",
		Short: "cold starts Nomen",
		Long: `cold starts Nomen.
First the initial events are created.
Last Nomen starts.

Requirements:
- database
- database is initialized
`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer func() {
				logging.OnError(cmd.Context(), err).Error("nomen start-from-setup command failed")
			}()

			err = tls.ModeFromFlag(cmd)
			if err != nil {
				return err
			}

			masterKey, err := key.MasterKey(cmd)
			if err != nil {
				return err
			}

			err = setup.BindInitProjections(cmd)
			if err != nil {
				return err
			}

			setupConfig, shutdown, err := setup.NewConfig(cmd, viper.GetViper())
			if err != nil {
				return err
			}
			defer func() {
				err = errors.Join(err, shutdown(cmd.Context()))
			}()

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
