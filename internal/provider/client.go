package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	httpClient *http.Client
	BaseURL    string
	APIKey     string
}

func NewClient(host, apiKey string, isApplication bool) *Client {
	base := strings.TrimRight(host, "/")
	if isApplication {
		base += "/api/application"
	} else {
		base += "/api/client"
	}

	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		BaseURL:    base,
		APIKey:     apiKey,
	}
}

func (c *Client) request(method, path string, body io.Reader) ([]byte, error) {
	url := fmt.Sprintf("%s%s", c.BaseURL, path)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return bodyBytes, nil
}

func (c *Client) Get(path string) ([]byte, error) {
	return c.request("GET", path, nil)
}

func (c *Client) Post(path string, payload any) ([]byte, error) {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(jsonData)
	}
	return c.request("POST", path, body)
}

func (c *Client) Put(path string, payload any) ([]byte, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return c.request("PUT", path, bytes.NewBuffer(jsonData))
}
func (c *Client) Patch(path string, payload any) ([]byte, error) {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(jsonData)
	}
	return c.request("PATCH", path, body)
}

func (c *Client) Delete(path string) error {
	_, err := c.request("DELETE", path, nil)
	return err
}
