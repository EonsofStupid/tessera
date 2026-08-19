package start

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/EonsofStupid/tessera/backend/v3/instrumentation/logging"
	"github.com/EonsofStupid/tessera/cmd/key"
	"github.com/EonsofStupid/tessera/cmd/tls"
)

var (
	startFlagSet = &pflag.FlagSet{}
)

func init() {
	startFlagSet.Uint16("port", 0, "port to run Tessera on")
	startFlagSet.String("externalDomain", "", "domain Tessera will be exposed on")
	startFlagSet.String("externalPort", "", "port Tessera will be exposed on")
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
