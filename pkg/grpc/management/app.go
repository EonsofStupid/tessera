package management

import "github.com/shippinAI/nomen/internal/api/grpc/server/middleware"

func (a *ListAppsResponse) Localizers() []middleware.Localizer {
	if a == nil {
		return nil
	}
	localizers := make([]middleware.Localizer, 0)
	for _, app := range a.Result {
		localizers = append(localizers, app.Localizers()...)
	}
	return localizers
}

func (a *GetAppByIDResponse) Localizers() []middleware.Localizer {
	if a == nil || a.App == nil {
		return nil
	}
	return a.App.Localizers()
}

func (a *AddOIDCAppResponse) Localizers() []middleware.Localizer {
	if a == nil || !a.NoneCompliant {
		return nil
	}
	localizers := make([]middleware.Localizer, len(a.ComplianceProblems))
	for i, problem := range a.ComplianceProblems {
		localizers[i] = problem
	}
	return localizers
}
