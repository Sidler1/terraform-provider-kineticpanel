package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type Client struct {
	httpClient *http.Client
	BaseURL    string
	APIKey     string
}

var DebugEnabled = strings.EqualFold(os.Getenv("KINETICPANEL_DEBUG"), "true")

func NewClient(host, apiKey string, isApplication bool) *Client {
	base := strings.TrimRight(host, "/")
	if isApplication {
		base += "/api/application"
	} else {
		base += "/api/client"
	}
	c := &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		BaseURL:    base,
		APIKey:     apiKey,
	}
	if DebugEnabled {
		tflog.Info(context.Background(), "Client created", map[string]any{
			"base_url":        c.BaseURL,
			"application_api": isApplication,
		})
	}
	return c
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

	if DebugEnabled {
		tflog.Debug(context.Background(), "HTTP request", map[string]any{
			"method": method,
			"url":    url,
			"headers": map[string]string{
				"Authorization": "Bearer ***",
				"Accept":        req.Header.Get("Accept"),
				"Content-Type":  req.Header.Get("Content-Type"),
			},
		})
		if body != nil {
			if b, ok := body.(*bytes.Buffer); ok {
				tflog.Debug(context.Background(), "Request payload", map[string]any{"body": b.String()})
			}
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if DebugEnabled {
		tflog.Debug(context.Background(), "HTTP response", map[string]any{
			"status": resp.Status,
			"body":   string(respBody),
		})
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errMsg := fmt.Sprintf("API error %d: %s", resp.StatusCode, string(respBody))
		tflog.Error(context.Background(), errMsg)
		return nil, fmt.Errorf(errMsg)
	}
	return respBody, nil
}

func (c *Client) Get(path string) ([]byte, error) {
	return c.request("GET", path, nil)
}
func (c *Client) Post(path string, payload any) ([]byte, error) {
	if payload == nil {
		return c.request("POST", path, nil)
	}
	data, _ := json.Marshal(payload)
	return c.request("POST", path, bytes.NewBuffer(data))
}
func (c *Client) Patch(path string, payload any) ([]byte, error) {
	data, _ := json.Marshal(payload)
	return c.request("PATCH", path, bytes.NewBuffer(data))
}
func (c *Client) Delete(path string) error { _, err := c.request("DELETE", path, nil); return err }
