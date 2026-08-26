package console

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/shippinAI/nomen/logging"
	"github.com/shippinAI/nomen/oidc/v3/pkg/op"

	"github.com/shippinAI/nomen/cmd/build"
	"github.com/shippinAI/nomen/internal/api/authz"
	http_util "github.com/shippinAI/nomen/internal/api/http"
	"github.com/shippinAI/nomen/internal/api/http/middleware"
	console_path "github.com/shippinAI/nomen/internal/api/ui/console/path"
)

type Config struct {
	ShortCache            middleware.CacheConfig
	LongCache             middleware.CacheConfig
	InstanceManagementURL string
	Edition               string
	DemoCaps              bool
	PostHog               struct {
		Token string
		URL   string
	}
}

type spaHandler struct {
	fileSystem http.FileSystem
}

var (
	//go:embed static
	static embed.FS
)

const (
	envRequestPath    = "/assets/environment.json"
	originPlaceholder = "__NOMEN_ORIGIN__"
	// https://posthog.com/docs/advanced/content-security-policy
	posthogCSPHost = "https://*.i.posthog.com"
)

var (
	shortCacheFiles = []string{
		"/",
		"/index.html",
		"/manifest.webmanifest",
		"/ngsw.json",
		"/ngsw-worker.js",
		"/safety-worker.js",
		"/worker-basic.min.js",
	}
)

func LoginHintLink(origin, username string) string {
	return origin + console_path.HandlerPrefix + "?login_hint=" + username
}

func (i *spaHandler) Open(name string) (http.File, error) {
	ret, err := i.fileSystem.Open(name)
	if !os.IsNotExist(err) || path.Ext(name) != "" {
		return ret, err
	}

	f, err := i.fileSystem.Open("/index.html")
	if err != nil {
		return nil, err
	}
	return &file{File: f}, nil
}

// file wraps the http.File and fs.FileInfo interfaces
// to return the build.Date() as ModTime() of the file
type file struct {
	http.File
	fs.FileInfo
}

func (f *file) ModTime() time.Time {
	return build.Date()
}

func (f *file) Stat() (_ fs.FileInfo, err error) {
	f.FileInfo, err = f.File.Stat()
	if err != nil {
		return nil, err
	}
	return f, nil
}

func Start(config Config, externalSecure bool, issuer op.IssuerFromRequest, callDurationInterceptor, instanceHandler func(http.Handler) http.Handler, limitingAccessInterceptor *middleware.AccessInterceptor, customerPortal string) (http.Handler, error) {
	fSys, err := fs.Sub(static, "static")
	if err != nil {
		return nil, err
	}
	cache := assetsCacheInterceptorIgnoreManifest(
		config.ShortCache.MaxAge,
		config.ShortCache.SharedMaxAge,
		config.LongCache.MaxAge,
		config.LongCache.SharedMaxAge,
	)
	security := middleware.SecurityHeaders(csp(config.PostHog.URL), nil)

	handler := mux.NewRouter()
	handler.Use(security, limitingAccessInterceptor.WithoutLimiting().Handle)

	env := handler.NewRoute().Path(envRequestPath).Subrouter()
	env.Use(
		callDurationInterceptor,
		middleware.RequestDetailsHandler(),
		middleware.TraceHandler(),
		middleware.LogHandler("console"),
		instanceHandler,
	)
	env.HandleFunc("", func(w http.ResponseWriter, r *http.Request) {
		url := http_util.BuildOrigin(r.Host, externalSecure)
		ctx := r.Context()
		instance := authz.GetInstance(ctx)
		instanceMgmtURL, err := templateInstanceManagementURL(config.InstanceManagementURL, instance)
		if err != nil {
			http.Error(w, fmt.Sprintf("unable to template instance management url for the management console: %v", err), http.StatusInternalServerError)
			return
		}
		limited := limitingAccessInterceptor.Limit(w, r)
		environmentJSON, err := createEnvironmentJSON(url, issuer(r), instance.ManagementConsoleClientID(), customerPortal, instanceMgmtURL, config.PostHog.URL, config.PostHog.Token, limited, config.Edition, config.DemoCaps)
		if err != nil {
			http.Error(w, fmt.Sprintf("unable to marshal env for the management console: %v", err), http.StatusInternalServerError)
			return
		}
		_, err = w.Write(environmentJSON)
		logging.OnError(err).Error("error serving environment.json")
	})
	handler.SkipClean(true).PathPrefix("").Handler(cache(spaWithRequestOrigin(fSys, externalSecure)))
	return handler, nil
}

// spaWithRequestOrigin serves the embedded application and resolves social
// metadata against the deployment origin. The public domain is runtime state,
// not a value the frontend build can safely guess; resolving it here keeps one
// image artifact correct for self-hosted and managed-customer deployments.
func spaWithRequestOrigin(fileSystem fs.FS, externalSecure bool) http.Handler {
	staticFiles := http.FileServer(&spaHandler{http.FS(fileSystem)})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isIndexOrSubPath(r.URL.Path) {
			staticFiles.ServeHTTP(w, r)
			return
		}
		index, err := fs.ReadFile(fileSystem, "index.html")
		if err != nil {
			http.Error(w, "Nomen application entry is unavailable", http.StatusInternalServerError)
			return
		}
		origin := template.HTMLEscapeString(http_util.BuildOrigin(r.Host, externalSecure))
		index = bytes.ReplaceAll(index, []byte(originPlaceholder), []byte(origin))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(index)
	})
}

func templateInstanceManagementURL(templateableCookieValue string, instance authz.Instance) (string, error) {
	cookieValueTemplate, err := template.New("cookievalue").Parse(templateableCookieValue)
	if err != nil {
		return templateableCookieValue, err
	}
	cookieValue := new(bytes.Buffer)
	if err = cookieValueTemplate.Execute(cookieValue, instance); err != nil {
		return templateableCookieValue, err
	}
	return cookieValue.String(), nil
}

func csp(posthogURL string) *middleware.CSP {
	csp := middleware.DefaultSCP
	csp.StyleSrc = csp.StyleSrc.AddInline()
	csp.ScriptSrc = csp.ScriptSrc.AddEval()
	csp.ConnectSrc = csp.ConnectSrc.AddOwnHost()
	csp.ImgSrc = csp.ImgSrc.AddOwnHost().AddScheme("blob")
	if posthogURL != "" {
		// https://posthog.com/docs/advanced/content-security-policy#enabling-the-toolbar
		csp.ScriptSrc = csp.ScriptSrc.AddHost(posthogCSPHost)
		csp.ConnectSrc = csp.ConnectSrc.AddHost(posthogCSPHost)
		csp.ImgSrc = csp.ImgSrc.AddHost(posthogCSPHost)
		csp.StyleSrc = csp.StyleSrc.AddHost(posthogCSPHost)
		csp.FontSrc = csp.FontSrc.AddHost(posthogCSPHost)
		csp.MediaSrc = middleware.CSPSourceOpts().AddHost(posthogCSPHost)
	}

	return &csp
}

func createEnvironmentJSON(api, issuer, clientID, customerPortal, instanceMgmtUrl, postHogURL, postHogToken string, exhausted bool, edition string, demoCaps bool) ([]byte, error) {
	if edition == "" {
		edition = "public"
	}
	environment := struct {
		API                   string `json:"api,omitempty"`
		Issuer                string `json:"issuer,omitempty"`
		ClientID              string `json:"clientid,omitempty"`
		CustomerPortal        string `json:"customer_portal,omitempty"`
		InstanceManagementURL string `json:"instance_management_url,omitempty"`
		PostHogURL            string `json:"posthog_url,omitempty"`
		PostHogToken          string `json:"posthog_token,omitempty"`
		Exhausted             bool   `json:"exhausted,omitempty"`
		Edition               string `json:"edition,omitempty"`
		DemoCaps              bool   `json:"demo_caps,omitempty"`
		Version               string `json:"version,omitempty"`
	}{
		API:                   api,
		Issuer:                issuer,
		ClientID:              clientID,
		CustomerPortal:        customerPortal,
		InstanceManagementURL: instanceMgmtUrl,
		PostHogURL:            postHogURL,
		PostHogToken:          postHogToken,
		Exhausted:             exhausted,
		Edition:               edition,
		DemoCaps:              demoCaps,
		Version:               build.Version(),
	}
	return json.Marshal(environment)
}

func assetsCacheInterceptorIgnoreManifest(shortMaxAge, shortSharedMaxAge, longMaxAge, longSharedMaxAge time.Duration) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, file := range shortCacheFiles {
				if r.URL.Path == file || isIndexOrSubPath(r.URL.Path) {
					middleware.AssetsCacheInterceptor(shortMaxAge, shortSharedMaxAge).Handler(handler).ServeHTTP(w, r)
					return
				}
			}
			middleware.AssetsCacheInterceptor(longMaxAge, longSharedMaxAge).Handler(handler).ServeHTTP(w, r)
		})
	}
}

func isIndexOrSubPath(path string) bool {
	//files will have an extension
	return !strings.Contains(path, ".")
}
