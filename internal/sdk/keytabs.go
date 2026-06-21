package mazevault

import (
	"fmt"
	"net/http"
)

// ListKeytabs lists all keytabs
func (c *Client) ListKeytabs() ([]Keytab, error) {
	r, err := c.newRequest(http.MethodGet, "/api/v1/keytabs", nil)
	if err != nil {
		return nil, err
	}
	var resp []Keytab
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// ImportKeytab imports a new keytab
func (c *Client) ImportKeytab(req *ImportKeytabRequest) (*Keytab, error) {
	r, err := c.newRequest(http.MethodPost, "/api/v1/keytabs", req)
	if err != nil {
		return nil, err
	}
	var resp Keytab
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetKeytab retrieves a keytab by ID
func (c *Client) GetKeytab(id string) (*Keytab, error) {
	r, err := c.newRequest(http.MethodGet, "/api/v1/keytabs/"+id, nil)
	if err != nil {
		return nil, err
	}
	var resp Keytab
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateKeytab updates a keytab's name or description
func (c *Client) UpdateKeytab(id string, req *UpdateKeytabRequest) (*Keytab, error) {
	r, err := c.newRequest(http.MethodPut, fmt.Sprintf("/api/v1/keytabs/%s", id), req)
	if err != nil {
		return nil, err
	}
	var resp Keytab
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteKeytab deletes a keytab
func (c *Client) DeleteKeytab(id string) error {
	r, err := c.newRequest(http.MethodDelete, "/api/v1/keytabs/"+id, nil)
	if err != nil {
		return err
	}
	return c.Do(r, nil)
}
