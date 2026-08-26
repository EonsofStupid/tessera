package start

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/shippinAI/nomen/backend/v3/instrumentation/logging"
	"github.com/shippinAI/nomen/cmd/key"
	"github.com/shippinAI/nomen/cmd/tls"
)

var (
	startFlagSet = &pflag.FlagSet{}
)

func init() {
	startFlagSet.Uint16("port", 0, "port to run Nomen on")
	startFlagSet.String("externalDomain", "", "domain Nomen will be exposed on")
	startFlagSet.String("externalPort", "", "port Nomen will be exposed on")
}

func startFlags(cmd *cobra.Command) {
	cmd.Flags().AddFlagSet(startFlagSet)
	logging.OnError(
		cmd.Context(),
		viper.BindPFlags(startFlagSet),
	).Fatal("start flags")

	tls.AddTLSModeFlag(cmd)
	key.AddMasterKeyFlag(cmd)
}
