package elasticsearch_api_client

import (
	"bytes"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"

	log "github.com/sirupsen/logrus"
)

type APIClient struct {
	Cfg *Configuration
}

type Configuration struct {
	Host          string            `json:"host,omitempty"`
	DefaultHeader map[string]string `json:"defaultHeader,omitempty"`
	UserAgent     string            `json:"userAgent,omitempty"`
	BasicAuth     BasicAuth         `json:"basicAuth,omitempty"`
	HTTPClient    *http.Client
}

type BasicAuth struct {
	UserName string `json:"userName,omitempty"`
	Password string `json:"password,omitempty"`
}

var (
	jsonCheck       = regexp.MustCompile("(?i:(?:application|text)/json)")
	apiClientLogger = log.WithFields(log.Fields{
		"component": "ApiClient",
	})
)

func (c *APIClient) CallAPI(request *http.Request) (*http.Response, error) {
	return c.Cfg.HTTPClient.Do(request)
}

func (c *APIClient) DoApiRequest(request *http.Request) ([]byte, *http.Response, error) {
	httpResponse, err := c.CallAPI(request)
	if err != nil || httpResponse == nil {
		apiClientLogger.Error(err)
		return nil, httpResponse, err
	}
	responseBody, err := ioutil.ReadAll(httpResponse.Body)
	_ = httpResponse.Body.Close()
	if err != nil {
		apiClientLogger.Error(err)
		return nil, httpResponse, err
	}
	return responseBody, httpResponse, err
}

func (c *APIClient) PrepareAndCall(
	path string,
	method string,
	postBody []byte,
	headerParams map[string]string,
	queryParams url.Values) ([]byte, *http.Response, error) {

	r, err := c.PrepareRequest(path, method, postBody, headerParams, queryParams)
	if err != nil {
		apiClientLogger.Error(err)
		return nil, nil, err
	}
	return c.DoApiRequest(r)
}

func (c *APIClient) PrepareRequest(
	path string,
	method string,
	postBody []byte,
	headerParams map[string]string,
	queryParams url.Values) (localVarRequest *http.Request, err error) {

	// Setup path and query parameters
	parsedUrl, err := url.Parse(c.Cfg.Host + "/" + path)
	if err != nil {
		apiClientLogger.Error(err)
		return nil, err
	}

	// Adding Query Param
	query := parsedUrl.Query()
	for k, v := range queryParams {
		for _, iv := range v {
			query.Add(k, iv)
		}
	}

	// Encode the parameters.
	parsedUrl.RawQuery = query.Encode()

	// Detect postBody type and make request.
	if postBody != nil {
		if headerParams == nil {
			headerParams = make(map[string]string)
		}
		contentType := headerParams["Content-Type"]
		if contentType == "" {
			contentType = "application/json; charset=utf-8"
			headerParams["Content-Type"] = contentType
		}
		localVarRequest, err = http.NewRequest(method, parsedUrl.String(), bytes.NewBuffer(postBody))
	} else {
		localVarRequest, err = http.NewRequest(method, parsedUrl.String(), nil)
	}

	// Add header parameters, if any
	if len(headerParams) > 0 {
		headers := http.Header{}
		for h, v := range headerParams {
			headers.Set(h, v)
		}
		localVarRequest.Header = headers
	}

	// Override request host, if applicable
	if c.Cfg.Host != "" {
		localVarRequest.Host = parsedUrl.Host
	}

	if c.Cfg.BasicAuth.UserName != "" && c.Cfg.BasicAuth.Password != "" {
		localVarRequest.Header.Add("Authorization", "Basic "+basicAuth(c.Cfg.BasicAuth.UserName, c.Cfg.BasicAuth.Password))
	}

	// Add the user agent to the request.
	localVarRequest.Header.Add("User-Agent", c.Cfg.UserAgent)
	for header, value := range c.Cfg.DefaultHeader {
		localVarRequest.Header.Add(header, value)
	}
	return localVarRequest, nil
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
