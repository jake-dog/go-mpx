package mpx

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestNewDsClient(t *testing.T) {
	creds := &AuthConfig{User: "bob", Password: "Test123!"}
	auth, _ := NewAuthClient(creds)
	result := NewDSClient("http://mpx.theplatform.com", "Test", auth, nil)
	if result == nil {
		t.Error("Expected a new DSClient, got ", result)
	}
	result = NewDSClient("https://mpx.theplatform.com", "Test", auth, nil)
	if result == nil {
		t.Error("Expected a new DSClient, got ", result)
	}
}

func TestBuildUrl(t *testing.T) {
	client := &dsClient{agent: "Test", baseUrl: "http://mpx.theplatform.com"}
	params := url.Values{}
	params.Set("schema", "1.0")
	params.Set("form", "json")
	result := client.buildUrl("/data/Media", params, nil)
	if result != "http://mpx.theplatform.com/data/Media?form=json&schema=1.0" {
		t.Error("Expected an encoded URL, got ", result)
	}
	result = client.buildUrl("/data/Media", params, []string{"123", "456"})
	if result != "http://mpx.theplatform.com/data/Media/123,456?form=json&schema=1.0" {
		t.Error("Expected an encoded URI, got ", result)
	}
	uris := []string{
		"http://mpx.theplatform.com/data/Media/123",
		"http://mpx.theplatform.com/data/Media/456",
	}
	result = client.buildUrl("/data/Media", params, uris)
	if result != "http://mpx.theplatform.com/data/Media/123,456?form=json&schema=1.0" {
		t.Error("Expected an encoded URI, got ", result)
	}
}

func TestBuildRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := "{\"signInResponse\": {\"token\": \"P9ahM3yCzEqIFWMww2qOAXC0wPCW0DBw\"}}"
		fmt.Fprintln(w, data)
	}))
	defer ts.Close()
	creds := &AuthConfig{User: "bob", Password: "Test123!", Idm: ts.URL}
	auth, _ := NewAuthClient(creds)
	client := &dsClient{
		agent: "Test", auth: auth,
		baseUrl: "http://mpx.theplatform.com",
	}
	reqUrl := "http://mpx.theplatform.com/data/Media/123,456?form=json&schema=1.0"
	result, err := client.buildRequest(reqUrl)
	if err != nil {
		t.Error("Expected an encoded URL, got ", result)
	}
}
