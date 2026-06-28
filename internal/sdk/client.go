package mazevault

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is the MazeVault API client
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Token      string
}

// NewClient creates a new MazeVault API client
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
	}
}

// SetToken sets the authentication token
func (c *Client) SetToken(token string) {
	c.Token = token
}

// Do performs an HTTP request
func (c *Client) Do(req *http.Request, v interface{}) error {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode >= 400 {
		msg := fmt.Sprintf("api error: status code %d, body: %s", res.StatusCode, string(body))
		var errResp ErrorResponse
		if jsonErr := json.Unmarshal(body, &errResp); jsonErr == nil && errResp.Error != "" {
			msg = fmt.Sprintf("api error: %s (details: %s)", errResp.Error, errResp.Details)
		}
		return &APIError{StatusCode: res.StatusCode, Message: msg}
	}

	if v != nil {
		if err := json.Unmarshal(body, v); err != nil {
			return err
		}
	}

	return nil
}

// newRequest creates a new HTTP request
func (c *Client) newRequest(method, path string, body interface{}) (*http.Request, error) {
	var buf io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		buf = bytes.NewBuffer(jsonBody)
	}

	return http.NewRequest(method, c.BaseURL+path, buf)
}
