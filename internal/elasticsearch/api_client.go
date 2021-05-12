package esapiclient

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"net/url"

	appConfig "github.com/aberestyak/elasticsearch-security-operator/config"
	log "github.com/sirupsen/logrus"
)

// APIClient defines elasticsearch API client config
type APIClient struct {
	Cfg *Configuration
}

// Configuration defines http client config
type Configuration struct {
	Host          string            `json:"host,omitempty"`
	DefaultHeader map[string]string `json:"defaultHeader,omitempty"`
	UserAgent     string            `json:"userAgent,omitempty"`
	BasicAuth     BasicAuth         `json:"basicAuth,omitempty"`
	HTTPClient    *http.Client
}

// BasicAuth defins basic auth configuration for http client
type BasicAuth struct {
	UserName string `json:"userName,omitempty"`
	Password string `json:"password,omitempty"`
}

var (
	apiClientLogger = log.WithFields(log.Fields{
		"component": "ApiClient",
	})
)

func (c *APIClient) callAPI(request *http.Request) (*http.Response, error) {
	return c.Cfg.HTTPClient.Do(request)
}

func (c *APIClient) doAPIRequest(request *http.Request) ([]byte, *http.Response, error) {
	httpResponse, err := c.callAPI(request)
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

// PrepareAndCall prepare http request and do it
func (c *APIClient) PrepareAndCall(
	path string,
	method string,
	postBody []byte,
	headerParams map[string]string,
	queryParams url.Values) ([]byte, *http.Response, error) {

	r, err := c.prepareRequest(path, method, postBody, headerParams, queryParams)
	if err != nil {
		apiClientLogger.Error(err)
		return nil, nil, err
	}
	return c.doAPIRequest(r)
}

func (c *APIClient) prepareRequest(
	path string,
	method string,
	postBody []byte,
	headerParams map[string]string,
	queryParams url.Values) (localVarRequest *http.Request, err error) {

	// Setup path and query parameters
	parsedURL, err := url.Parse(c.Cfg.Host + "/" + path)
	if err != nil {
		apiClientLogger.Error(err)
		return nil, err
	}

	// Adding Query Param
	query := parsedURL.Query()
	for k, v := range queryParams {
		for _, iv := range v {
			query.Add(k, iv)
		}
	}

	// Encode the parameters.
	parsedURL.RawQuery = query.Encode()

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
		localVarRequest, _ = http.NewRequest(method, parsedURL.String(), bytes.NewBuffer(postBody))
	} else {
		localVarRequest, _ = http.NewRequest(method, parsedURL.String(), nil)
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
		localVarRequest.Host = parsedURL.Host
	}

	if c.Cfg.BasicAuth.UserName != "" && c.Cfg.BasicAuth.Password != "" {
		localVarRequest.Header.Add("Authorization", "Basic "+basicAuth(c.Cfg.BasicAuth.UserName, c.Cfg.BasicAuth.Password))
	}

	// Add the user agent to the request.
	localVarRequest.Header.Add("User-Agent", c.Cfg.UserAgent)
	for header, value := range c.Cfg.DefaultHeader {
		localVarRequest.Header.Add(header, value)
	}
	// Add custom CA certificate if appropriate env is set
	if appConfig.AppConfig.ExtraCACert != nil {
		c.addCustomCACert()
	}
	return localVarRequest, nil
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (c *APIClient) addCustomCACert() {
	// Setup HTTPS client
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		RootCAs:    appConfig.AppConfig.ExtraCACert,
	}
	tlsConfig.BuildNameToCertificate()
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	c.Cfg.HTTPClient = &http.Client{Transport: transport}
}
