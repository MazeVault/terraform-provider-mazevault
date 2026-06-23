package mazevault

import (
	"fmt"
	"net/http"
)

// CreateAPIToken creates a new API token
func (c *Client) CreateAPIToken(name string, scopes []string, duration string) (*APIToken, error) {
	reqBody := CreateAPITokenRequest{
		Name:     name,
		Scopes:   scopes,
		Duration: duration,
	}

	req, err := c.newRequest(http.MethodPost, "/api/v1/users/tokens", reqBody)
	if err != nil {
		return nil, err
	}

	var resp APIToken
	if err := c.Do(req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// ListAPITokens lists all API tokens
func (c *Client) ListAPITokens() ([]APIToken, error) {
	req, err := c.newRequest(http.MethodGet, "/api/v1/users/tokens", nil)
	if err != nil {
		return nil, err
	}

	var resp []APIToken
	if err := c.Do(req, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

// RevokeAPIToken revokes an API token
func (c *Client) RevokeAPIToken(id string) error {
	path := fmt.Sprintf("/api/v1/users/tokens/%s", id)
	req, err := c.newRequest(http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	return c.Do(req, nil)
}
