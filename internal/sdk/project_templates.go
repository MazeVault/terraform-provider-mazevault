package mazevault

import (
	"fmt"
	"net/http"
	"time"
)

// ProjectTemplate is an organization-level reusable template that appears in the
// frontend "Project Templates" view. For Category "certificates" the backend
// automatically creates (or reuses) the underlying CertificateTemplate and stores
// its ID in ProtocolConfig["certificate_template_id"].
type ProjectTemplate struct {
	ID               string                 `json:"id"`
	OrganizationID   string                 `json:"organization_id"`
	Name             string                 `json:"name"`
	Description      string                 `json:"description,omitempty"`
	Category         string                 `json:"category"`
	Subtype          string                 `json:"subtype,omitempty"`
	Version          string                 `json:"version,omitempty"`
	ProtocolConfig   map[string]interface{} `json:"protocol_config,omitempty"`
	EnvironmentRules map[string]interface{} `json:"environment_rules,omitempty"`
	VariableSchema   []interface{}          `json:"variable_schema,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// ProjectTemplateInput is the request body for creating/updating a project template
// via POST/PUT /api/v1/organizations/:id/templates.
type ProjectTemplateInput struct {
	Name             string                 `json:"name"`
	Description      string                 `json:"description,omitempty"`
	Category         string                 `json:"category"`
	Subtype          string                 `json:"subtype,omitempty"`
	Version          string                 `json:"version,omitempty"`
	ProtocolConfig   map[string]interface{} `json:"protocol_config,omitempty"`
	EnvironmentRules map[string]interface{} `json:"environment_rules,omitempty"`
	VariableSchema   []interface{}          `json:"variable_schema,omitempty"`
}

// ListOrgProjectTemplates lists an organization's project templates (the entities
// shown in the frontend "Project Templates" view).
func (c *Client) ListOrgProjectTemplates(organizationID string) ([]ProjectTemplate, error) {
	r, err := c.newRequest(http.MethodGet, "/api/v1/organizations/"+organizationID+"/templates", nil)
	if err != nil {
		return nil, err
	}
	var resp []ProjectTemplate
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetOrgProjectTemplate fetches a single project template by ID via the list
// endpoint (the backend exposes no direct GET-by-ID for project templates).
func (c *Client) GetOrgProjectTemplate(organizationID, templateID string) (*ProjectTemplate, error) {
	templates, err := c.ListOrgProjectTemplates(organizationID)
	if err != nil {
		return nil, err
	}
	for i := range templates {
		if templates[i].ID == templateID {
			return &templates[i], nil
		}
	}
	return nil, nil
}

// CreateOrgProjectTemplate creates a project template. For Category "certificates"
// the backend auto-creates the underlying CertificateTemplate.
func (c *Client) CreateOrgProjectTemplate(organizationID string, req *ProjectTemplateInput) (*ProjectTemplate, error) {
	r, err := c.newRequest(http.MethodPost, "/api/v1/organizations/"+organizationID+"/templates", req)
	if err != nil {
		return nil, err
	}
	var resp ProjectTemplate
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateOrgProjectTemplate updates a project template by ID.
func (c *Client) UpdateOrgProjectTemplate(templateID string, req *ProjectTemplateInput) (*ProjectTemplate, error) {
	r, err := c.newRequest(http.MethodPut, fmt.Sprintf("/api/v1/organizations/templates/%s", templateID), req)
	if err != nil {
		return nil, err
	}
	var resp ProjectTemplate
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteOrgProjectTemplate deletes a project template by ID.
func (c *Client) DeleteOrgProjectTemplate(templateID string) error {
	r, err := c.newRequest(http.MethodDelete, fmt.Sprintf("/api/v1/organizations/templates/%s", templateID), nil)
	if err != nil {
		return err
	}
	return c.Do(r, nil)
}
