package mazevault

import (
	"fmt"
	"net/http"
)

// ListCAAccounts lists CA accounts for an organization
func (c *Client) ListCAAccounts(organizationID string) ([]CAAccount, error) {
	r, err := c.newRequest(http.MethodGet, "/api/v1/organizations/"+organizationID+"/ca-accounts/", nil)
	if err != nil {
		return nil, err
	}
	var resp []CAAccount
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// ConnectCAAccount connects a CA account to an organization
func (c *Client) ConnectCAAccount(organizationID string, req *CreateCAAccountRequest) (*CAAccount, error) {
	r, err := c.newRequest(http.MethodPost, "/api/v1/organizations/"+organizationID+"/ca-accounts/", req)
	if err != nil {
		return nil, err
	}
	var resp CAAccount
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCAAccount retrieves a CA account by ID
func (c *Client) GetCAAccount(organizationID, accountID string) (*CAAccount, error) {
	r, err := c.newRequest(http.MethodGet, fmt.Sprintf("/api/v1/organizations/%s/ca-accounts/%s", organizationID, accountID), nil)
	if err != nil {
		return nil, err
	}
	var resp CAAccount
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateCAAccount updates a CA account
func (c *Client) UpdateCAAccount(organizationID, accountID string, req *CreateCAAccountRequest) (*CAAccount, error) {
	r, err := c.newRequest(http.MethodPut, fmt.Sprintf("/api/v1/organizations/%s/ca-accounts/%s", organizationID, accountID), req)
	if err != nil {
		return nil, err
	}
	var resp CAAccount
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteCAAccount disconnects a CA account
func (c *Client) DeleteCAAccount(organizationID, accountID string) error {
	r, err := c.newRequest(http.MethodDelete, fmt.Sprintf("/api/v1/organizations/%s/ca-accounts/%s", organizationID, accountID), nil)
	if err != nil {
		return err
	}
	return c.Do(r, nil)
}
