package mpx

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestNewAuthClient(t *testing.T) {
	auth := &AuthConfig{User: "bob", Password: "test"}
	authclient, err := NewAuthClient(auth)
	if err != nil {
		t.Error("Expected no errors got ", err)
	}
	if authclient.Account() != DefaultAcct {
		t.Error("Expected default account, instead got ", authclient.Account())
	}
	_, err = NewAuthClient(&AuthConfig{User: "bob"})
	if err == nil {
		t.Error("Invalid user/password not rejected")
	}
	_, err = NewAuthClient(&AuthConfig{Password: "bob"})
	if err == nil {
		t.Error("Invalid user/password not rejected")
	}
	auth = &AuthConfig{
		User: "bob", Password: "test",
		Acct: "VMS POC VOD Ingest",
	}
	_, err = NewAuthClient(auth)
	if err == nil {
		t.Error("Expected validatiion error, instead got AuthClient")
	}
}

func TestSignin(t *testing.T) {
	i := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var data string
		if user, _, _ := r.BasicAuth(); user == "bob" {
			data = "{\"signInResponse\": {\"token\": \"P9ahM3yCzEqIFWMww2qOAXC0wPCW0DBw\"}}"
			i += 1
		} else {
			data = "{\"isException\": true, \"title\": \"AuthenticationException\", \"description\": \"Could not authenticate user\"}"
		}
		fmt.Fprintln(w, data)
	}))
	defer ts.Close()
	auth := authentication{
		user: "bob", pass: "test",
		idm: ts.URL, idmClient: &http.Client{},
	}
	auth.Signin()
	if auth.token != "P9ahM3yCzEqIFWMww2qOAXC0wPCW0DBw" {
		t.Error("Expected no token, instead got ", auth.token)
	}
	auth.Signin()
	if i != 1 {
		t.Error("Expected no additional HTTP calls after signin")
	}
	auth = authentication{
		user: "joe", pass: "test",
		idm: ts.URL, idmClient: &http.Client{},
	}
	auth.Signin()
	if (auth.token != "") && (i != 1) {
		t.Error("Expected a failed signin with invalid user")
	}
}

func TestToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := "{\"signInResponse\": {\"token\": \"P9ahM3yCzEqIFWMww2qOAXC0wPCW0DBw\"}}"
		fmt.Fprintln(w, data)
	}))
	defer ts.Close()
	auth := authentication{
		user: "bob", pass: "test",
		idm: ts.URL, idmClient: &http.Client{},
	}
	auth.Token()
	if auth.token != "P9ahM3yCzEqIFWMww2qOAXC0wPCW0DBw" {
		t.Error("Expected no token, instead got ", auth.token)
	}
}

func TestSignout(t *testing.T) {
	i := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := "{\"signOutResponse\": {}}"
		i += 1
		fmt.Fprintln(w, data)
	}))
	defer ts.Close()
	auth := authentication{
		user: "bob", pass: "test",
		idm: ts.URL, idmClient: &http.Client{},
		token: "P9ahM3yCzEqIFWMww2qOAXC0wPCW0DBw",
	}
	auth.Signout()
	if auth.token != "" {
		t.Error("Expected no token, instead got ", auth.token)
	}
	auth.Signout()
	if i != 1 {
		t.Error("Expected no additional HTTP calls after signout")
	}
}

func TestResolveService(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := "{\"resolveDomainResponse\": {\"Media Data Service\": \"http://data.media.theplatform.com/media\"}}"
		fmt.Fprintln(w, data)
	}))
	defer ts.Close()
	params := url.Values{}
	params.Set("schema", "1.1")
	params.Set("form", "json")
	params.Set("_accountId", DefaultAcct)

	auth := &authentication{
		user: "bob", pass: "test", acct: DefaultAcct,
		token: "P9ahM3yCzEqIFWMww2qOAXC0wPCW0DBw",
	}
	auth.accessClient = NewDSClient(ts.URL, AuthAgent, auth, params)
	service := auth.ResolveService("Media Data Service")
	if service != "http://data.media.theplatform.com/media" {
		t.Error("Expected MDS URL, instead got ", service)
	}
}
