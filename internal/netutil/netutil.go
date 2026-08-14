package netutil

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"migu-video-go/internal/log"
)

var Client = &http.Client{
	Timeout: 6 * time.Second,
}

// FetchJSON performs an HTTP GET and decodes the JSON response.
func FetchJSON(url string, headers map[string]string, result interface{}) error {
	return FetchJSONWithTimeout(url, headers, 6*time.Second, result)
}

// FetchJSONWithTimeout performs an HTTP GET with custom timeout and decodes JSON.
func FetchJSONWithTimeout(url string, headers map[string]string, timeout time.Duration, result interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		log.Red("请求失败: " + err.Error())
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, result)
}

// FetchJSONWithClient performs an HTTP GET using a custom http.Client and decodes JSON.
func FetchJSONWithClient(url string, headers map[string]string, client *http.Client, result interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Red("请求失败: " + err.Error())
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, result)
}

// FetchRaw performs an HTTP GET and returns the raw response body bytes.
func FetchRaw(url string, headers map[string]string) ([]byte, http.Header, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := Client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.Header, err
	}
	return body, resp.Header, nil
}

// FetchPostJSON performs an HTTP POST with JSON body and decodes the response.
func FetchPostJSON(url string, headers map[string]string, body string, result interface{}) error {
	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := Client.Do(req)
	if err != nil {
		log.Red("请求失败: " + err.Error())
		return err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(respBody, result)
}
