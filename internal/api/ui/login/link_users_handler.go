package login

import (
	"net/http"

	http_mw "github.com/shippinAI/nomen/internal/api/http/middleware"
	"github.com/shippinAI/nomen/internal/domain"
)

const (
	tmplLinkUsersDone = "linkusersdone"
)

func (l *Login) linkUsers(w http.ResponseWriter, r *http.Request, authReq *domain.AuthRequest, err error) {
	userAgentID, _ := http_mw.UserAgentIDFromCtx(r.Context())
	err = l.authRepo.LinkExternalUsers(setContext(r.Context(), authReq.UserOrgID), authReq.ID, userAgentID, domain.BrowserInfoFromRequest(r))
	l.renderLinkUsersDone(w, r, authReq, err)
}

func (l *Login) renderLinkUsersDone(w http.ResponseWriter, r *http.Request, authReq *domain.AuthRequest, err error) {
	translator := l.getTranslator(r.Context(), authReq)
	data := l.getUserData(r, authReq, translator, "LinkingUsersDone.Title", "LinkingUsersDone.Description", err)
	l.renderer.RenderTemplate(w, r, translator, l.renderer.Templates[tmplLinkUsersDone], data, nil)
}
