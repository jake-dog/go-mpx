package mpx

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

//TODO: Add an MPXError struct to report responseCode: Class: description
//TODO: https://blog.golang.org/http-tracing
//TODO: https://blog.golang.org/godoc-documenting-go-code
//TODO: We should be able to stream output using a chunked stream and
//limited byte slice with the io.Reader interface to buffer each chunk, parse it
//and pass it to the client for processing.  A linked list will be more
//preferential for buffering sets of objecst due to the lighter memory footprint

type DSClient interface {
	GetJSON(path string, params url.Values, ids []string) (map[string]interface{}, error)
	GetCount(path string, params url.Values, ids []string) (int, error)
}

type dsClient struct {
	agent, baseUrl string
	auth           AuthClient
	baseParams     url.Values
	client         *http.Client
}

func NewDSClient(baseUrl, agent string, auth AuthClient, baseParams url.Values) DSClient {
	var client *http.Client
	u, err := url.Parse(baseUrl)
	if err != nil {
		log.Fatal(err)
	}
	if u.Scheme == "https" {
		tr := &http.Transport{TLSClientConfig: &tls.Config{}}
		client = &http.Client{Transport: tr}
	} else {
		client = &http.Client{}
	}
	return &dsClient{
		agent: agent, baseUrl: baseUrl, auth: auth,
		baseParams: baseParams, client: client,
	}
}

func (c *dsClient) GetJSON(path string, params url.Values, ids []string) (map[string]interface{}, error) {
	reqUrl := c.buildUrl(path, params, ids)
	req, err := c.buildRequest(reqUrl)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var dat map[string]interface{}
	if err := json.Unmarshal(body, &dat); err != nil {
		return nil, err
	}
	if _, haserr := dat["isException"]; haserr {
		panic(dat["description"].(string))
	}
	return dat, nil
}

func (c *dsClient) GetCount(path string, params url.Values, ids []string) (int, error) {
	params.Set("entries", "false")
	params.Set("count", "true")
	dat, err := c.GetJSON(path, params, ids)
	if err != nil {
		return 0, err
	}
	return dat["totalResults"].(int), nil
}

func (c *dsClient) buildRequest(reqUrl string) (*http.Request, error) {
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", c.agent)
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(c.auth.Account(), c.auth.Token())
	return req, nil
}

func (c *dsClient) buildUrl(path string, params url.Values, ids []string) string {
	// MPX doesn't allow multi-value query params, so it's safe to use v[0]
	if params != nil {
		for k, v := range c.baseParams {
			if value := params.Get(k); value == "" {
				params.Set(k, v[0])
			}
		}
	} else {
		params = c.baseParams
	}
	//TODO: We're not removing dups, but we should be!
	if ids != nil {
		for i, id := range ids {
			if strings.LastIndex(id, "/") != -1 {
				ids[i] = id[strings.LastIndex(id, "/")+1:]
			}
		}
		return fmt.Sprintf("%s%s/%s?%s", c.baseUrl, path, strings.Join(ids, ","), params.Encode())
	}
	log.Print(fmt.Sprintf("%s%s?%s", c.baseUrl, path, params.Encode()))
	return fmt.Sprintf("%s%s?%s", c.baseUrl, path, params.Encode())
}
