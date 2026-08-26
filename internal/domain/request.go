package domain

const (
	OrgDomainPrimaryScope = "urn:nomen:iam:org:domain:primary:"
	OrgIDScope            = "urn:nomen:iam:org:id:"
	OrgRoleIDScope        = "urn:nomen:iam:org:roles:id:"
	OrgDomainPrimaryClaim = "urn:nomen:iam:org:domain:primary"
	OrgIDClaim            = "urn:nomen:iam:org:id"
	ProjectIDScope        = "urn:nomen:iam:org:project:id:"
	ProjectIDScopeNOMEN = "nomen"
	AudSuffix             = ":aud"
	ProjectScopeNOMEN   = ProjectIDScope + ProjectIDScopeNOMEN + AudSuffix
	SelectIDPScope        = "urn:nomen:iam:org:idp:id:"
)

// TODO: Change AuthRequest to interface and let oidcauthreqesut implement it
type Request interface {
	Type() AuthRequestType
	IsValid() bool
	GetScopes() []string
}

type AuthRequestType int32

const (
	AuthRequestTypeOIDC AuthRequestType = iota
	AuthRequestTypeSAML
	AuthRequestTypeDevice
)

type AuthRequestOIDC struct {
	Scopes        []string
	ResponseType  OIDCResponseType
	ResponseMode  OIDCResponseMode
	Nonce         string
	CodeChallenge *OIDCCodeChallenge
}

func NewAuthRequestOIDC(
	scopes []string,
	responseType OIDCResponseType,
	responseMode OIDCResponseMode,
	nonce string,
	codeChallenge *OIDCCodeChallenge,
) *AuthRequestOIDC {
	return &AuthRequestOIDC{
		scopes, responseType, responseMode, nonce, codeChallenge,
	}
}

func (a *AuthRequestOIDC) Type() AuthRequestType {
	return AuthRequestTypeOIDC
}

func (a *AuthRequestOIDC) IsValid() bool {
	return len(a.Scopes) > 0 &&
		a.CodeChallenge == nil || a.CodeChallenge != nil && a.CodeChallenge.IsValid()
}

func (a *AuthRequestOIDC) GetScopes() []string {
	return a.Scopes
}

type AuthRequestSAML struct {
	ID          string
	BindingType string
	Code        string
	Issuer      string
	IssuerName  string
	Destination string
}

func (a *AuthRequestSAML) Type() AuthRequestType {
	return AuthRequestTypeSAML
}

func (a *AuthRequestSAML) IsValid() bool {
	return true
}

func (*AuthRequestSAML) GetScopes() []string {
	return nil
}

type AuthRequestDevice struct {
	ClientID    string
	DeviceCode  string
	UserCode    string
	Scopes      []string
	Audience    []string
	AppName     string
	ProjectName string
}

func NewAuthRequestDevice(
	clientID string,
	deviceCode string,
	userCode string,
	scopes []string,
	audience []string,
	appName string,
	projectName string,
) *AuthRequestDevice {
	return &AuthRequestDevice{
		clientID, deviceCode, userCode, scopes, audience, appName, projectName,
	}
}

func (*AuthRequestDevice) Type() AuthRequestType {
	return AuthRequestTypeDevice
}

func (a *AuthRequestDevice) IsValid() bool {
	return a.DeviceCode != "" && a.UserCode != ""
}

func (a *AuthRequestDevice) GetScopes() []string {
	return a.Scopes
}
