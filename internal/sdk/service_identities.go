package mazevault

import "net/http"

// CreateServiceIdentity creates a new service identity (machine account)
func (c *Client) CreateServiceIdentity(req *CreateServiceIdentityRequest) (*CreateServiceIdentityResponse, error) {
	r, err := c.newRequest(http.MethodPost, "/api/v1/service-identities/", req)
	if err != nil {
		return nil, err
	}
	var resp CreateServiceIdentityResponse
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetServiceIdentity retrieves a service identity by ID
func (c *Client) GetServiceIdentity(id string) (*ServiceIdentity, error) {
	r, err := c.newRequest(http.MethodGet, "/api/v1/service-identities/"+id, nil)
	if err != nil {
		return nil, err
	}
	var resp ServiceIdentity
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateServiceIdentity updates display name, description, or owner email
func (c *Client) UpdateServiceIdentity(id string, req *CreateServiceIdentityRequest) (*ServiceIdentity, error) {
	r, err := c.newRequest(http.MethodPut, "/api/v1/service-identities/"+id, req)
	if err != nil {
		return nil, err
	}
	var resp ServiceIdentity
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteServiceIdentity deletes a service identity
func (c *Client) DeleteServiceIdentity(id string) error {
	r, err := c.newRequest(http.MethodDelete, "/api/v1/service-identities/"+id, nil)
	if err != nil {
		return err
	}
	return c.Do(r, nil)
}

// ListServiceIdentities lists all service identities
func (c *Client) ListServiceIdentities() ([]ServiceIdentity, error) {
	r, err := c.newRequest(http.MethodGet, "/api/v1/service-identities/", nil)
	if err != nil {
		return nil, err
	}
	var resp []ServiceIdentity
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}
