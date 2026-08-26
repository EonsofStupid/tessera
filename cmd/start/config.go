package start

import (
	"errors"
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"
	old_logging "github.com/shippinAI/nomen/logging" //nolint:staticcheck
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/shippinAI/nomen/backend/v3/instrumentation"
	"github.com/shippinAI/nomen/backend/v3/instrumentation/logging"
	nomen_blueprint "github.com/shippinAI/nomen/cmd/blueprint"
	"github.com/shippinAI/nomen/cmd/encryption"
	"github.com/shippinAI/nomen/cmd/hooks"
	"github.com/shippinAI/nomen/internal/actions"
	admin_es "github.com/shippinAI/nomen/internal/admin/repository/eventsourcing"
	"github.com/shippinAI/nomen/internal/api/authz"
	"github.com/shippinAI/nomen/internal/api/http"
	"github.com/shippinAI/nomen/internal/api/http/middleware"
	"github.com/shippinAI/nomen/internal/api/oidc"
	"github.com/shippinAI/nomen/internal/api/saml"
	scim_config "github.com/shippinAI/nomen/internal/api/scim/config"
	"github.com/shippinAI/nomen/internal/api/ui/console"
	"github.com/shippinAI/nomen/internal/api/ui/login"
	"github.com/shippinAI/nomen/internal/api/well_known"
	auth_es "github.com/shippinAI/nomen/internal/auth/repository/eventsourcing"
	"github.com/shippinAI/nomen/internal/cache/connector"
	"github.com/shippinAI/nomen/internal/command"
	"github.com/shippinAI/nomen/internal/config/hook"
	"github.com/shippinAI/nomen/internal/config/network"
	"github.com/shippinAI/nomen/internal/config/systemdefaults"
	"github.com/shippinAI/nomen/internal/database"
	"github.com/shippinAI/nomen/internal/denylist"
	"github.com/shippinAI/nomen/internal/domain"
	"github.com/shippinAI/nomen/internal/eventstore"
	"github.com/shippinAI/nomen/internal/execution"
	"github.com/shippinAI/nomen/internal/id"
	"github.com/shippinAI/nomen/internal/logstore"
	"github.com/shippinAI/nomen/internal/notification/handlers"
	"github.com/shippinAI/nomen/internal/query/projection"
	"github.com/shippinAI/nomen/internal/serviceping"
	static_config "github.com/shippinAI/nomen/internal/static/config"
)

type Config struct {
	Instrumentation     instrumentation.Config
	Log                 *old_logging.Config
	Port                uint16
	ExternalPort        uint16
	ExternalDomain      string
	ExternalSecure      bool
	TLS                 network.TLS
	InstanceHostHeaders []string
	PublicHostHeaders   []string
	HTTP2HostHeader     string
	HTTP1HostHeader     string
	WebAuthNName        string
	Edition             string
	DemoCaps            bool
	Database            database.Config
	// Blueprints.Dir, when set, is applied to every instance on every start —
	// declared identity state, from reviewed files, with no human in the loop.
	Blueprints        nomen_blueprint.StartupConfig
	Caches            *connector.CachesConfig
	Tracing           *instrumentation.LegacyTraceConfig
	Metrics           *instrumentation.LegacyMetricConfig
	Profiler          *instrumentation.LegacyProfileConfig
	Projections       projection.Config
	Notifications     handlers.WorkerConfig
	Executions        execution.WorkerConfig
	Auth              auth_es.Config
	Admin             admin_es.Config
	UserAgentCookie   *middleware.UserAgentCookieConfig
	OIDC              oidc.Config
	SAML              saml.Config
	SCIM              scim_config.Config
	Login             login.Config
	Console           console.Config
	WellKnown         well_known.Config
	AssetStorage      static_config.AssetStorageConfig
	InternalAuthZ     authz.Config
	SystemAuthZ       authz.Config
	SystemDefaults    systemdefaults.SystemDefaults
	EncryptionKeys    *encryption.EncryptionKeyConfig
	DefaultInstance   command.InstanceSetup
	AuditLogRetention time.Duration
	SystemAPIUsers    map[string]*authz.SystemAPIUser
	CustomerPortal    string
	Machine           *id.Config
	Actions           *actions.Config
	Eventstore        *eventstore.Config
	LogStore          *logstore.Configs
	Quotas            *QuotasConfig
	Telemetry         *handlers.TelemetryPusherConfig
	ServicePing       *serviceping.Config
	HTTPClient        *http.ClientConfig
}

type QuotasConfig struct {
	Access struct {
		logstore.EmitterConfig  `mapstructure:",squash"`
		middleware.AccessConfig `mapstructure:",squash"`
	}
	Execution *logstore.EmitterConfig
}

func NewConfig(cmd *cobra.Command, v *viper.Viper) (*Config, instrumentation.ShutdownFunc, error) {
	config, err := readConfig(v)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to read config: %w", err)
	}

	config.Instrumentation.Trace.SetLegacyConfig(config.Tracing)
	config.Instrumentation.Metric.SetLegacyConfig(config.Metrics)
	config.Instrumentation.Log.SetLegacyConfig(config.Log)
	config.Instrumentation.Profile.SetLegacyConfig(config.Profiler)
	shutdown, err := instrumentation.Start(cmd.Context(), config.Instrumentation)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to start instrumentation: %w", err)
	}
	cmd.SetContext(logging.NewCtx(cmd.Context(), logging.StreamRuntime))

	// Legacy logger
	err = config.Log.SetLogger()
	if err != nil {
		err = errors.Join(err, shutdown(cmd.Context()))
		return nil, nil, fmt.Errorf("unable to set logger: %w", err)
	}

	id.Configure(config.Machine)

	var actionsDenylist []denylist.AddressChecker
	if config.Actions != nil {
		actionsDenylist = config.Actions.HTTP.DenyList
	}
	config.HTTPClient.MergeDeprecatedDenylists(actionsDenylist, config.Executions.DenyList)

	err = config.SystemDefaults.Validate()
	if err != nil {
		err = errors.Join(err, shutdown(cmd.Context()))
		return nil, nil, fmt.Errorf("system defaults config invalid: %w", err)
	}
	// Copy the global role permissions mappings to the instance until we allow instance-level configuration over the API.
	config.DefaultInstance.RolePermissionMappings = config.InternalAuthZ.RolePermissionMappings

	return config, shutdown, nil
}

func readConfig(v *viper.Viper) (*Config, error) {
	config := new(Config)

	err := v.Unmarshal(config,
		viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
			hooks.SliceTypeStringDecode[*domain.CustomMessageText],
			hooks.SliceTypeStringDecode[authz.RoleMapping],
			hooks.MapTypeStringDecode[string, *authz.SystemAPIUser],
			hooks.MapHTTPHeaderStringDecode,
			database.DecodeHook(false),
			actions.HTTPConfigDecodeHook,
			denylist.DenyListDecodeHook,
			hook.EnumHookFunc(authz.MemberTypeString),
			hooks.MapTypeStringDecode[domain.Feature, any],
			hooks.SliceTypeStringDecode[*command.SetQuota],
			hook.Base64ToBytesHookFunc(),
			hook.TagToLanguageHookFunc(),
			hook.StringToURLHookFunc(),
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToTimeHookFunc(time.RFC3339),
			mapstructure.StringToSliceHookFunc(","),
			mapstructure.TextUnmarshallerHookFunc(),
		)),
	)
	return config, err
}
