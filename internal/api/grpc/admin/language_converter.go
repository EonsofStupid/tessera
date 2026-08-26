package admin

import (
	"golang.org/x/text/language"

	"github.com/shippinAI/nomen/internal/domain"
	"github.com/shippinAI/nomen/pkg/grpc/admin"
)

func selectLanguagesToCommand(languages *admin.SelectLanguages) (tags []language.Tag, err error) {
	allowedLanguages := languages.GetList()
	if allowedLanguages == nil && languages != nil {
		allowedLanguages = make([]string, 0)
	}
	if allowedLanguages == nil {
		return nil, nil
	}
	return domain.ParseLanguage(allowedLanguages...)
}
