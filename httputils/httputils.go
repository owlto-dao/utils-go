package httputils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
func (c *Client) DoGet(ctx context.Context, url string, headers map[string]string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	// Add headers
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	return c.doRequest(req)
}

// DoPost 发送一个 HTTP POST 请求，支持 JSON、表单和纯文本数据
func (c *Client) DoPost(ctx context.Context, urlStr string, data interface{}, headers map[string]string) (string, error) {
	var body []byte
	var err error
	var contentType string

	// 根据 data 的类型推断内容类型并进行相应处理
	switch v := data.(type) {
	case map[string]interface{}: // 如果是 JSON 数据
		body, err = json.Marshal(v)
		if err != nil {
			return "", err
		}
		contentType = "application/json"
	case url.Values: // 如果是表单数据
		body = []byte(v.Encode())
		contentType = "application/x-www-form-urlencoded"
	case string: // 如果是纯文本数据
		body = []byte(v)
		contentType = "text/plain"
	default:
		return "", errors.New("不支持的数据类型")
	}

	// 创建 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", contentType)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("请求失败，状态码: " + resp.Status)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// Helper function for executing HTTP requests and handling responses.
func (c *Client) doRequest(req *http.Request) (string, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed with status: " + resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}