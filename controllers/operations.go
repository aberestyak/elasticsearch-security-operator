package controllers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	config "github.com/aberestyak/elasticsearch-security-operator/config"
	elasticsearch_api_client "github.com/aberestyak/elasticsearch-security-operator/internal/elasticsearch"
	log "github.com/sirupsen/logrus"
)

var (
	defaultRequest = elasticsearch_api_client.APIClient{
		Cfg: &elasticsearch_api_client.Configuration{
			Host:      config.AppConfig.ElasticsearchEndpoint,
			UserAgent: "elasticsearch-security-operator-client/go",
			BasicAuth: elasticsearch_api_client.BasicAuth{
				UserName: config.AppConfig.ElasticsearchUsername,
				Password: config.AppConfig.ElasticsearchPassword,
			},
			HTTPClient: &http.Client{},
		},
	}
	apiClientWrapperLogger = log.WithFields(log.Fields{
		"component": "ApiClientWrapper",
	})
)

// MakeAPIRequest - make request to endpoint
func MakeAPIRequest(method string, path string, jsonBody []byte) (ObjectID string, Status string, ResponseBody []byte, Error error) {
	responseBody, httpResponse, err := defaultRequest.PrepareAndCall(path, method, jsonBody, nil, url.Values{})
	if err != nil {
		apiClientWrapperLogger.Errorf("Error when creating new object: %v", err.Error())
		return "", "", nil, err
	}
	RequesDebugtLogger := log.WithFields(log.Fields{
		"component": "RequestDebug",
	})
	RequesDebugtLogger.Debugf("Method: %v. Path: %v. Body: %v. ResponseCode: %v. ResponseBody: %v", method, path, string(jsonBody), httpResponse.StatusCode, string(responseBody))
	return GetResponseObjectID(responseBody), GetResponseStatus(httpResponse), responseBody, nil
}

// GetResponseStatus - return Error or Deployed based on http status code
func GetResponseStatus(response *http.Response) string {
	if (response.StatusCode < 300) && (response.StatusCode >= 200) {
		return "Deployed"
	}
	return "Error"
}

// GetResponseObjectID - get object ID, if exists
func GetResponseObjectID(responseBody []byte) string {
	var result map[string]string
	if err := json.Unmarshal(responseBody, &result); err != nil {
		apiClientWrapperLogger.Tracef("Can't unmarshal response body: %v", err)
	}
	return result["_id"]
}

// SanitizeQuery - parse RawMessage to delete escape characters
func SanitizeQuery(query json.RawMessage) json.RawMessage {
	sanitizedQuery := strings.ReplaceAll(string(query), "\\n", "")
	sanitizedQuery = strings.ReplaceAll(sanitizedQuery, "\\\"", "\"")
	sanitizedQuery = strings.Trim(sanitizedQuery, "\"")
	return []byte(sanitizedQuery)
}

// GetExistingObject - make GET request to get existing elsticsearch object
func GetExistingObject(path, ID string) (bool, []byte, error) {
	_, responseResult, responseBody, err := MakeAPIRequest("GET", path+"/"+ID, nil)
	if err != nil {
		return false, nil, err
	}
	if responseResult == "Error" {
		return false, nil, nil
	}
	return true, responseBody, nil
}
