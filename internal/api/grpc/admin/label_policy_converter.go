package admin

import (
	"github.com/EonsofStupid/tessera/internal/domain"
	admin_pb "github.com/EonsofStupid/tessera/pkg/grpc/admin"
	policy_pb "github.com/EonsofStupid/tessera/pkg/grpc/policy"
)

func updateLabelPolicyToDomain(policy *admin_pb.UpdateLabelPolicyRequest) *domain.LabelPolicy {
	return &domain.LabelPolicy{
		PrimaryColor:        policy.PrimaryColor,
		BackgroundColor:     policy.BackgroundColor,
		WarnColor:           policy.WarnColor,
		FontColor:           policy.FontColor,
		PrimaryColorDark:    policy.PrimaryColorDark,
		BackgroundColorDark: policy.BackgroundColorDark,
		WarnColorDark:       policy.WarnColorDark,
		FontColorDark:       policy.FontColorDark,
		HideLoginNameSuffix: policy.HideLoginNameSuffix,
		DisableWatermark:    policy.DisableWatermark,
		ThemeMode:           themeModeToDomain(policy.ThemeMode),
	}
}

func themeModeToDomain(theme policy_pb.ThemeMode) domain.LabelPolicyThemeMode {
	switch theme {
	case policy_pb.ThemeMode_THEME_MODE_AUTO:
		return domain.LabelPolicyThemeAuto
	case policy_pb.ThemeMode_THEME_MODE_DARK:
		return domain.LabelPolicyThemeDark
	case policy_pb.ThemeMode_THEME_MODE_LIGHT:
		return domain.LabelPolicyThemeLight
	case policy_pb.ThemeMode_THEME_MODE_UNSPECIFIED:
		return domain.LabelPolicyThemeAuto
	default:
		return domain.LabelPolicyThemeAuto
	}
}
