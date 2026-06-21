package mazevault

import (
	"fmt"
	"net/http"
)

// ListEnvironments lists environments for an organization
func (c *Client) ListEnvironments(organizationID string) ([]Environment, error) {
	r, err := c.newRequest(http.MethodGet, "/api/v1/organizations/"+organizationID+"/environments", nil)
	if err != nil {
		return nil, err
	}
	var resp []Environment
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// AddEnvironment adds a new environment to an organization
func (c *Client) AddEnvironment(organizationID string, req *CreateEnvironmentRequest) (*Environment, error) {
	r, err := c.newRequest(http.MethodPost, "/api/v1/organizations/"+organizationID+"/environments", req)
	if err != nil {
		return nil, err
	}
	var resp Environment
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateEnvironment patches an environment
func (c *Client) UpdateEnvironment(organizationID, environmentID string, req *UpdateEnvironmentRequest) (*Environment, error) {
	r, err := c.newRequest(http.MethodPatch, fmt.Sprintf("/api/v1/organizations/%s/environments/%s", organizationID, environmentID), req)
	if err != nil {
		return nil, err
	}
	var resp Environment
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// RemoveEnvironment deletes an environment from an organization
func (c *Client) RemoveEnvironment(organizationID, environmentID string) error {
	r, err := c.newRequest(http.MethodDelete, fmt.Sprintf("/api/v1/organizations/%s/environments/%s", organizationID, environmentID), nil)
	if err != nil {
		return err
	}
	return c.Do(r, nil)
}
