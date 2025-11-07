package provider

import (
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

func (c *Client) DoRequest(method, path string, body io.Reader) ([]byte, error) {
	url := fmt.Sprintf("%s%s", c.BaseURL, path)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, resp.Status)
	}

	return io.ReadAll(resp.Body)
}
