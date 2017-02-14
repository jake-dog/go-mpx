package mpx

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

const (
	DefaultAccess = "http://access.auth.theplatform.com"
	DefaultIdm    = "https://identity.auth.theplatform.com/idm"
	DefaultAcct   = "http://access.auth.theplatform.com/data/Account/1"
	AuthAgent     = "Go-MPX AuthClient"
	SignInPath    = "/web/Authentication/signIn"
	SignOutPath   = "/web/Authentication/signOut"
	RegistryPath  = "/web/Registry/resolveDomain"
)

type badStringError struct {
	what string
	str  string
}

func (e *badStringError) Error() string { return fmt.Sprintf("%s %q", e.what, e.str) }

type AuthClient interface {
	Signin()
	Signout()
	Token() string
	Account() string
	ResolveService(service string) string
	GetSelf() map[string]interface{}
}

type AuthConfig struct {
	User, Password    string
	Acct, Access, Idm string
}

type authentication struct {
	token, idm       string
	user, pass, acct string
	idmClient        *http.Client
	accessClient     DSClient
	registry         map[string]interface{}
}

func NewAuthClient(conf *AuthConfig) (a AuthClient, err error) {
	// Grab our authentication parameters if specified
	access := StringParamDefault(conf.Access, DefaultAccess)
	idm := StringParamDefault(conf.Idm, DefaultIdm)
	acct := StringParamDefault(conf.Acct, DefaultAcct)

	// Perform a quick parameter validation
	for _, serviceurl := range []string{access, idm, acct} {
		// For some reason using "err" has a linter warning...?
		_, urlerr := url.ParseRequestURI(serviceurl)
		if urlerr != nil {
			return nil, urlerr
		}
	}
	if conf.User == "" {
		return nil, &badStringError{"Missing parameter", "User"}
	}
	if conf.Password == "" {
		return nil, &badStringError{"Missing parameter", "Password"}
	}

	// Declare our url.Values for DSClient
	params := url.Values{}
	params.Set("schema", "1.1")
	params.Set("form", "json")
	params.Set("_accountId", acct)

	// idmClient is not a regular DSClient because we haven't retrieved a token.
	// We also already validated "idm", so we can skip the error check here
	var idmClient *http.Client
	if u, _ := url.Parse(idm); u.Scheme == "https" {
		tr := &http.Transport{TLSClientConfig: &tls.Config{}}
		idmClient = &http.Client{Transport: tr}
	} else {
		idmClient = &http.Client{}
	}

	// Little bit of initialization magic here since we're going to create our
	// authentication object, then pass a pointer to itself to it's own
	// instance of 'accessClient'
	temp_a := &authentication{
		user: conf.User, pass: conf.Password, acct: acct,
		idmClient: idmClient, idm: idm,
	}
	temp_a.accessClient = NewDSClient(access, AuthAgent, temp_a, params)
	a = temp_a

	return a, nil
}

func (t *authentication) Token() string {
	if t.token == "" {
		t.Signin()
	}
	return t.token
}

func (t *authentication) Account() string {
	return t.acct
}

func (t *authentication) Signin() {
	if t.token != "" {
		return
	}
	params := url.Values{}
	params.Set("schema", "1.0")
	params.Set("form", "json")

	// We can safely ignore errors when generating this request because we already
	// validated it in the public struct constructor
	reqUrl := fmt.Sprintf("%s%s?%s", t.idm, SignInPath, params.Encode())
	req, _ := http.NewRequest("GET", reqUrl, nil)
	req.Header.Add("User-Agent", AuthAgent)
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(t.user, t.pass)

	// Actually make our request
	resp, err := t.idmClient.Do(req)
	defer resp.Body.Close()
	if err != nil {
		log.Print(err)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	var dat map[string]interface{}
	if err := json.Unmarshal(body, &dat); err != nil {
		log.Print(err)
		return
	}
	if _, haserr := dat["isException"]; haserr {
		log.Print(dat["description"].(string))
		return
	}
	t.token = dat["signInResponse"].(map[string]interface{})["token"].(string)
	//log.Print(dat)
	//log.Print(t.token)
}

func (t *authentication) Signout() {
	if t.token == "" {
		return
	}
	params := url.Values{}
	params.Set("schema", "1.0")
	params.Set("form", "json")
	params.Set("_token", t.token)

	// We can safely ignore errors when generating this request because we already
	// validated it in the public struct constructor
	reqUrl := fmt.Sprintf("%s%s?%s", t.idm, SignOutPath, params.Encode())
	req, _ := http.NewRequest("GET", reqUrl, nil)
	req.Header.Add("User-Agent", AuthAgent)
	req.Header.Add("Content-Type", "application/json")

	// We really don't care if this fails, so we can ignore errors
	resp, _ := t.idmClient.Do(req)
	defer resp.Body.Close()
	t.token = ""
	t.registry = nil
}

func (t *authentication) ResolveService(service string) string {
	if t.registry == nil {
		regResp, err := t.accessClient.GetJSON(RegistryPath, nil, nil)
		if err != nil {
			return ""
		}
		log.Print(regResp)
		t.registry = regResp["resolveDomainResponse"].(map[string]interface{})
	}

	return (t.registry[service]).(string)
}

func (t *authentication) GetSelf() map[string]interface{} {
	return make(map[string]interface{})
}
