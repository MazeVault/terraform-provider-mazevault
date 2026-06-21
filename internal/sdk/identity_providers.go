package mazevault

import (
	"fmt"
	"net/http"
)

// ListIdentityProviders lists all configured identity providers
func (c *Client) ListIdentityProviders() ([]IdentityProvider, error) {
	r, err := c.newRequest(http.MethodGet, "/api/v1/identity-providers", nil)
	if err != nil {
		return nil, err
	}
	var resp []IdentityProvider
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// CreateIdentityProvider creates a new identity provider
func (c *Client) CreateIdentityProvider(req *CreateIdentityProviderRequest) (*IdentityProvider, error) {
	r, err := c.newRequest(http.MethodPost, "/api/v1/identity-providers", req)
	if err != nil {
		return nil, err
	}
	var resp IdentityProvider
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetIdentityProvider retrieves an identity provider by ID
func (c *Client) GetIdentityProvider(id string) (*IdentityProvider, error) {
	r, err := c.newRequest(http.MethodGet, "/api/v1/identity-providers/"+id, nil)
	if err != nil {
		return nil, err
	}
	var resp IdentityProvider
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateIdentityProvider updates an existing identity provider
func (c *Client) UpdateIdentityProvider(id string, req *CreateIdentityProviderRequest) (*IdentityProvider, error) {
	r, err := c.newRequest(http.MethodPut, fmt.Sprintf("/api/v1/identity-providers/%s", id), req)
	if err != nil {
		return nil, err
	}
	var resp IdentityProvider
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteIdentityProvider deletes an identity provider
func (c *Client) DeleteIdentityProvider(id string) error {
	r, err := c.newRequest(http.MethodDelete, "/api/v1/identity-providers/"+id, nil)
	if err != nil {
		return err
	}
	return c.Do(r, nil)
}
