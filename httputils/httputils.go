package httputils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client wraps an HTTP client with additional functionality.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new HTTP client with a customizable timeout.
func NewClient(timeout time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// DoGet sends an HTTP GET request with context support and headers.
func (c *Client) DoGet(ctx context.Context, url string, headers map[string]string, response interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	// Add headers
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	return c.doRequest(req, response)
}

// DoPost sends an HTTP POST request and supports JSON, form, and plain text payloads.
func (c *Client) DoPost(ctx context.Context, urlStr string, data interface{}, headers map[string]string, response interface{}) error {
	var body []byte
	var err error
	var contentType string

	// Infer the content type from data and handle it accordingly.
	switch v := data.(type) {
	case string: // Plain text payload.
		body = []byte(v)
		contentType = "text/plain"
	case url.Values: // Form payload.
		body = []byte(v.Encode())
		contentType = "application/x-www-form-urlencoded"
	default:
		// Try to serialize the payload as JSON.
		body, err = json.Marshal(data)
		if err != nil {
			return fmt.Errorf("无法将数据序列化为 JSON：%v", err)
		}
		contentType = "application/json"
	}

	// Create the HTTP request.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", contentType)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("请求失败，状态码: " + resp.Status)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(body, response); err != nil {
		return err
	}
	return nil
}

// Helper function for executing HTTP requests and handling responses.
func (c *Client) doRequest(req *http.Request, response interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed with status: " + resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, response)
	if err != nil {
		return fmt.Errorf("json.Unmarshal err: %w, body: %v", err, string(body))
	}
	return nil
}
