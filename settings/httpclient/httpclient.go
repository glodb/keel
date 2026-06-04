package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/glodb/keel/settings/logger"

	"github.com/bytedance/sonic"
)

type httpclient struct {
}

var getInstance = sync.OnceValue(func() *httpclient {
	instance := &httpclient{}
	return instance
})

func GetInstance() *httpclient {
	return getInstance()
}

func (c *httpclient) HTTPGet(url string, queryParams map[string]string, result interface{}) error {
	// Create URL with query parameters
	reqURL, err := c.addQueryParams(url, queryParams)
	if err != nil {
		return fmt.Errorf("failed to add query parameters: %v", err)
	}

	// Make HTTP GET request
	response, err := http.Get(reqURL)
	if err != nil {
		return fmt.Errorf("HTTP GET request failed: %v", err)
	}
	defer response.Body.Close()

	// Check if the response status code is OK (200)
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP GET request returned non-OK status: %s", response.Status)
	}
	logger.Log().Info("HTTP GET request", logger.StringField("url", url), logger.AnyField("queryParams", queryParams), logger.StringField("responseStatus", response.Status))
	// Decode JSON into the provided result interface
	err = json.NewDecoder(response.Body).Decode(result)
	if err != nil {
		logger.Log().Error("JSON decoding failed", logger.ErrorField("error", err))
		return err
	}

	return nil
}

func (c *httpclient) HTTPGetWithHeaders(url string, queryParams map[string]string, headers map[string]string, result interface{}) error {
	// Create URL with query parameters
	reqURL, err := c.addQueryParams(url, queryParams)
	if err != nil {
		return fmt.Errorf("failed to add query parameters: %v", err)
	}

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return fmt.Errorf("failed to add query parameters: %v", err)
	}

	if len(headers) > 0 {
		for key, value := range headers {
			req.Header.Add(key, value)
		}
	}

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		logger.Log().Error("HTTP GET request failed", logger.ErrorField("error", err))
		return fmt.Errorf("HTTP GET request failed: %v", err)
	}
	defer response.Body.Close()

	// Check if the response status code is OK (200)
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP GET request returned non-OK status: %s", response.Status)
	}

	logger.Log().Info("HTTP GET request", logger.StringField("url", url), logger.AnyField("queryParams", queryParams), logger.StringField("responseStatus", response.Status))
	// Decode JSON into the provided result interface
	err = json.NewDecoder(response.Body).Decode(result)
	if err != nil {
		logger.Log().Error("JSON decoding failed", logger.ErrorField("error", err))
		return err
	}

	return nil
}

func (c *httpclient) HTTPDelete(url string, queryParams map[string]string, headers map[string]string, result interface{}) error {
	// Create URL with query parameters
	reqURL, err := c.addQueryParams(url, queryParams)
	if err != nil {
		return fmt.Errorf("failed to add query parameters: %v", err)
	}

	req, err := http.NewRequest(http.MethodDelete, reqURL, nil)
	if err != nil {
		return fmt.Errorf("failed to add query parameters: %v", err)
	}

	if len(headers) > 0 {
		for key, value := range headers {
			req.Header.Add(key, value)
		}
	}

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		logger.Log().Error("HTTP DELETE request failed", logger.ErrorField("error", err))
		return fmt.Errorf("HTTP DELETE request failed: %v", err)
	}
	defer response.Body.Close()

	// Check if the response status code is OK (200)
	if response.StatusCode != http.StatusOK {
		json.NewDecoder(response.Body).Decode(result)
		return fmt.Errorf("HTTP DELETE request returned non-OK status: %s", response.Status)
	}

	logger.Log().Info("HTTP DELETE request", logger.StringField("url", url), logger.AnyField("queryParams", queryParams), logger.StringField("responseStatus", response.Status))
	// Decode JSON into the provided result interface
	err = json.NewDecoder(response.Body).Decode(result)
	if err != nil {
		logger.Log().Error("JSON decoding failed", logger.ErrorField("error", err))
		return err
	}

	return nil
}

func (c *httpclient) postRequest(req *http.Request, url string, headers map[string]string, result interface{}) error {
	// Create URL with query parameters
	if len(headers) > 0 {
		for key, value := range headers {
			req.Header.Add(key, value)
		}
	}

	// Use http.Client to send the request
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		logger.Log().Error("HTTP POST request failed", logger.ErrorField("error", err))
		return fmt.Errorf("HTTP POST request failed: %v", err)
	}
	defer response.Body.Close()

	// Check if the response status code is OK (200)
	if response.StatusCode != http.StatusOK {

		type RefreshTokenError struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		var refreshTokenError RefreshTokenError
		bodyText, _ := io.ReadAll(response.Body)
		sonic.Unmarshal(bodyText, &refreshTokenError)

		logger.Log().Error("HTTP POST request returned non-OK status", logger.StringField("responseStatus", response.Status), logger.StringField("error", refreshTokenError.Error), logger.StringField("description", refreshTokenError.ErrorDescription))
		return fmt.Errorf("HTTP POST request returned non-OK status: %s", response.Status)
	}

	logger.Log().Info("Sending Request", logger.StringField("url", url), logger.AnyField("headers", headers), logger.StringField("responseStatus", response.Status), logger.AnyField("responseBody", response.Body))
	// Decode JSON into the provided result interface
	err = json.NewDecoder(response.Body).Decode(result)
	if err != nil {
		logger.Log().Error("JSON decoding failed", logger.ErrorField("error", err))
		return err
	}

	return nil

}

func (c *httpclient) HTTPFormPost(url string, headers map[string]string, payload *strings.Reader, result interface{}) error {
	// Create an HTTP POST request
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return fmt.Errorf("failed to create HTTP POST request: %v", err)
	}

	return c.postRequest(req, url, headers, result)
}

func (c *httpclient) HTTPPost(url string, headers map[string]string, payload map[string]string, result interface{}) error {

	jsonPayload, err := sonic.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON payload: %v", err)
	}
	// Create an HTTP POST request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create HTTP POST request: %v", err)
	}

	return c.postRequest(req, url, headers, result)
}

// addQueryParams adds query parameters to the given URL
func (c *httpclient) addQueryParams(urlStr string, params map[string]string) (string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	query := parsedURL.Query()
	for key, value := range params {
		query.Add(key, value)
	}

	parsedURL.RawQuery = query.Encode()
	return parsedURL.String(), nil
}
