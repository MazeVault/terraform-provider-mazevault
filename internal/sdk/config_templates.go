package mazevault

import (
	"fmt"
	"net/http"
)

// ListConfigTemplates lists config templates for an organization
func (c *Client) ListConfigTemplates(organizationID string) ([]ConfigTemplate, error) {
	r, err := c.newRequest(http.MethodGet, "/api/v1/organizations/"+organizationID+"/config-templates", nil)
	if err != nil {
		return nil, err
	}
	var resp []ConfigTemplate
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// CreateConfigTemplate creates a config template
func (c *Client) CreateConfigTemplate(organizationID string, req *CreateConfigTemplateRequest) (*ConfigTemplate, error) {
	r, err := c.newRequest(http.MethodPost, "/api/v1/organizations/"+organizationID+"/config-templates", req)
	if err != nil {
		return nil, err
	}
	var resp ConfigTemplate
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetConfigTemplate retrieves a config template by searching the list (no dedicated GET)
func (c *Client) GetConfigTemplate(organizationID, templateID string) (*ConfigTemplate, error) {
	all, err := c.ListConfigTemplates(organizationID)
	if err != nil {
		return nil, err
	}
	for i, t := range all {
		if t.ID == templateID {
			return &all[i], nil
		}
	}
	return nil, nil
}

// UpdateConfigTemplate updates an existing config template
func (c *Client) UpdateConfigTemplate(organizationID, templateID string, req *CreateConfigTemplateRequest) (*ConfigTemplate, error) {
	r, err := c.newRequest(http.MethodPut, fmt.Sprintf("/api/v1/organizations/%s/config-templates/%s", organizationID, templateID), req)
	if err != nil {
		return nil, err
	}
	var resp ConfigTemplate
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteConfigTemplate deletes a config template
func (c *Client) DeleteConfigTemplate(organizationID, templateID string) error {
	r, err := c.newRequest(http.MethodDelete, fmt.Sprintf("/api/v1/organizations/%s/config-templates/%s", organizationID, templateID), nil)
	if err != nil {
		return err
	}
	return c.Do(r, nil)
}
