package tls

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	flagTLSMode = "tlsMode"
)

var (
	ErrValidValue = errors.New("value must either be `enabled`, `external` or `disabled`")
)

func AddTLSModeFlag(cmd *cobra.Command) {
	if cmd.PersistentFlags().Lookup(flagTLSMode) != nil {
		return
	}
	cmd.PersistentFlags().String(flagTLSMode, "", "start Tessera with (enabled), without (disabled) TLS or an external component such as a reverse proxy (external) terminating TLS; this flag overrides `externalSecure` and `tls.enabled` in config files")
}

func ModeFromFlag(cmd *cobra.Command) error {
	tlsMode, _ := cmd.Flags().GetString(flagTLSMode)
	var tlsEnabled, externalSecure bool
	switch tlsMode {
	case "enabled":
		tlsEnabled = true
		externalSecure = true
	case "external":
		tlsEnabled = false
		externalSecure = true
	case "disabled":
		tlsEnabled = false
		externalSecure = false
	case "":
		return nil
	default:
		return ErrValidValue
	}
	viper.Set("tls.enabled", tlsEnabled)
	viper.Set("externalSecure", externalSecure)
	return nil
}
