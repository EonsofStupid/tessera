package i18n

import (
	"net/http"

	"github.com/rakyll/statik/fs"
	"github.com/shippinAI/nomen/logging"

	// ensure fs is setup
	_ "github.com/shippinAI/nomen/internal/api/ui/login/statik"
	_ "github.com/shippinAI/nomen/internal/notification/statik"
	_ "github.com/shippinAI/nomen/internal/statik"
)

var nomenFS, loginFS, notificationFS http.FileSystem

type Namespace string

const (
	NOMEN        Namespace = "nomen"
	LOGIN        Namespace = "login"
	NOTIFICATION Namespace = "notification"
)

func LoadFilesystem(ns Namespace) http.FileSystem {
	var err error
	defer func() {
		if err != nil {
			logging.WithFields("namespace", ns).OnError(err).Panic("unable to get namespace")
		}
	}()
	switch ns {
	case NOMEN:
		if nomenFS != nil {
			return nomenFS
		}
		nomenFS, err = fs.NewWithNamespace(string(ns))
		return nomenFS
	case LOGIN:
		if loginFS != nil {
			return loginFS
		}
		loginFS, err = fs.NewWithNamespace(string(ns))
		return loginFS
	case NOTIFICATION:
		if notificationFS != nil {
			return notificationFS
		}
		notificationFS, err = fs.NewWithNamespace(string(ns))
		return notificationFS
	}
	return nil
}
