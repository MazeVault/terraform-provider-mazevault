package mazevault

import (
	"net/http"
)

// Integration represents an external system integration
type Integration struct {
	ID        string                 `json:"id"`
	ProjectID string                 `json:"project_id"`
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	Provider  string                 `json:"provider"`
	Config    map[string]interface{} `json:"config"`
}

// ListIntegrations lists integrations for a project
func (c *Client) ListIntegrations(projectID string) ([]Integration, error) {
	req, err := c.newRequest(http.MethodGet, "/api/v1/projects/"+projectID+"/integrations", nil)
	if err != nil {
		return nil, err
	}
	var resp []Integration
	if err := c.Do(req, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// CreateIntegration creates an integration
func (c *Client) CreateIntegration(projectID, name, typ, provider, environment string, config map[string]interface{}) (*Integration, error) {
	body := map[string]interface{}{
		"name":        name,
		"type":        typ,
		"provider":    provider,
		"environment": environment,
		"config":      config,
	}
	req, err := c.newRequest(http.MethodPost, "/api/v1/projects/"+projectID+"/integrations", body)
	if err != nil {
		return nil, err
	}
	var resp Integration
	if err := c.Do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteIntegration removes an integration
func (c *Client) DeleteIntegration(projectID, integrationID string) error {
	req, err := c.newRequest(http.MethodDelete, "/api/v1/projects/integrations/"+integrationID, nil)
	if err != nil {
		return err
	}
	return c.Do(req, nil)
}

// SecretLink represents a link between a secret and an integration
type SecretLink struct {
	ID            string `json:"id"`
	IntegrationID string `json:"integration_id"`
	DatabaseUser  string `json:"database_username,omitempty"`
	TargetPath    string `json:"target_path,omitempty"`
	FileFormat    string `json:"file_format,omitempty"`
	SecretKey     string `json:"secret_key,omitempty"`
	VariableName  string `json:"variable_name,omitempty"`
	Environment   string `json:"environment,omitempty"`
}

// ListSecretLinks lists links for a secret
func (c *Client) ListSecretLinks(secretID string) ([]SecretLink, error) {
	req, err := c.newRequest(http.MethodGet, "/api/v1/secrets/"+secretID+"/links", nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Links []SecretLink `json:"links"`
	}
	if err := c.Do(req, &resp); err != nil {
		return nil, err
	}
	return resp.Links, nil
}

// CreateSecretLink creates a link between secret and integration
func (c *Client) CreateSecretLink(secretID string, payload map[string]interface{}) (*SecretLink, error) {
	req, err := c.newRequest(http.MethodPost, "/api/v1/secrets/"+secretID+"/links", payload)
	if err != nil {
		return nil, err
	}
	var resp SecretLink
	if err := c.Do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteSecretLink deletes a link
func (c *Client) DeleteSecretLink(secretID, linkID string) error {
	req, err := c.newRequest(http.MethodDelete, "/api/v1/secrets/"+secretID+"/links/"+linkID, nil)
	if err != nil {
		return err
	}
	return c.Do(req, nil)
}

// UpdateIntegration updates an existing project integration
func (c *Client) UpdateIntegration(integrationID string, name, typ, provider string, config map[string]interface{}) (*Integration, error) {
	body := map[string]interface{}{
		"name":     name,
		"type":     typ,
		"provider": provider,
		"config":   config,
	}
	req, err := c.newRequest(http.MethodPut, "/api/v1/projects/integrations/"+integrationID, body)
	if err != nil {
		return nil, err
	}
	var resp Integration
	if err := c.Do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetIntegration retrieves a single project integration by ID
func (c *Client) GetIntegration(integrationID string) (*Integration, error) {
	req, err := c.newRequest(http.MethodGet, "/api/v1/projects/integrations/"+integrationID, nil)
	if err != nil {
		return nil, err
	}
	var resp Integration
	if err := c.Do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
