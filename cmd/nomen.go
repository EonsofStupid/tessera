package cmd

import (
	"bytes"
	"context"
	"crypto/fips140"
	_ "embed"
	"errors"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/shippinAI/nomen/backend/v3/instrumentation/logging"
	"github.com/shippinAI/nomen/cmd/admin"
	"github.com/shippinAI/nomen/cmd/blueprint"
	"github.com/shippinAI/nomen/cmd/build"
	"github.com/shippinAI/nomen/cmd/initialise"
	"github.com/shippinAI/nomen/cmd/key"
	"github.com/shippinAI/nomen/cmd/mirror"
	"github.com/shippinAI/nomen/cmd/ready"
	"github.com/shippinAI/nomen/cmd/seat"
	"github.com/shippinAI/nomen/cmd/setup"
	"github.com/shippinAI/nomen/cmd/start"
)

var (
	configFiles []string

	//go:embed defaults.yaml
	defaultConfig []byte

	//go:embed defaults_fips.yaml
	defaultFipsConfig []byte
)

func New(out io.Writer, in io.Reader, args []string, server chan<- *start.Server) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nomen",
		Short: "Manage the Nomen identity and authorization service",
		Long:  `Manage the Nomen identity and authorization service`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("no additional command provided")
		},
		Version:      build.Version(),
		SilenceUsage: true,
	}

	configureEnvironment(viper.GetViper())
	viper.SetConfigType("yaml")
	err := loadDefaultConfig()
	logging.OnError(context.Background(), err).Fatal("unable to read default config")

	cobra.OnInitialize(initConfig)
	cmd.PersistentFlags().StringArrayVar(&configFiles, "config", nil, "path to config file to overwrite system defaults")

	cmd.AddCommand(
		admin.New(), //is now deprecated, remove later on
		initialise.New(),
		setup.New(),
		start.New(server),
		start.NewStartFromInit(server),
		start.NewStartFromSetup(server),
		mirror.New(&configFiles),
		key.New(),
		ready.New(),
		seat.New(),
		blueprint.New(),
	)

	cmd.InitDefaultVersionFlag()

	return cmd
}

func configureEnvironment(v *viper.Viper) {
	v.AutomaticEnv()
	v.SetEnvPrefix("NOMEN")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

func loadDefaultConfig() error {
	return loadDefaultConfigInto(viper.GetViper())
}

func loadDefaultConfigInto(v *viper.Viper) error {
	if err := v.ReadConfig(bytes.NewBuffer(defaultConfig)); err != nil {
		return err
	}
	return mergeFipsDefaultConfig(v)
}

func mergeFipsDefaultConfig(v *viper.Viper) error {
	if !fips140.Enabled() {
		return nil
	}
	return applyFipsDefaultOverlay(v)
}

func applyFipsDefaultOverlay(v *viper.Viper) error {
	return v.MergeConfig(bytes.NewBuffer(defaultFipsConfig))
}

func initConfig() {
	for _, file := range configFiles {
		viper.SetConfigFile(file)
		err := viper.MergeInConfig()
		logging.OnError(context.Background(), err).Warn("unable to read config file", "file", file)
	}
}
