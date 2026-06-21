package mazevault

import (
	"fmt"
	"net/http"
)

// ListOrganizations retrieves all organizations
func (c *Client) ListOrganizations() ([]Organization, error) {
	req, err := c.newRequest(http.MethodGet, "/organizations", nil)
	if err != nil {
		return nil, err
	}

	var resp []Organization
	if err := c.Do(req, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

// GetOrganization retrieves an organization by ID
func (c *Client) GetOrganization(id string) (*Organization, error) {
	path := fmt.Sprintf("/organizations/%s", id)
	req, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp Organization
	if err := c.Do(req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// CreateOrganization creates a new organization
func (c *Client) CreateOrganization(name string) (*Organization, error) {
	reqBody := CreateOrganizationRequest{
		Name: name,
	}

	req, err := c.newRequest(http.MethodPost, "/organizations", reqBody)
	if err != nil {
		return nil, err
	}

	var resp Organization
	if err := c.Do(req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// UpdateOrganization updates an organization
func (c *Client) UpdateOrganization(id, name string) (*Organization, error) {
	reqBody := CreateOrganizationRequest{
		Name: name,
	}

	path := fmt.Sprintf("/organizations/%s", id)
	req, err := c.newRequest(http.MethodPut, path, reqBody)
	if err != nil {
		return nil, err
	}

	var resp Organization
	if err := c.Do(req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// DeleteOrganization deletes an organization
func (c *Client) DeleteOrganization(id string) error {
	path := fmt.Sprintf("/organizations/%s", id)
	req, err := c.newRequest(http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	return c.Do(req, nil)
}
