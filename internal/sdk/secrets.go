package mazevault

import (
	"fmt"
	"net/http"
)

// GetSecret retrieves a secret by key and project ID
func (c *Client) GetSecret(projectID, key string) (*Secret, error) {
	path := fmt.Sprintf("/secrets?project_id=%s&key=%s", projectID, key)
	req, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp Secret
	if err := c.Do(req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetSecretByID retrieves a secret by ID
func (c *Client) GetSecretByID(id string) (*Secret, error) {
	path := fmt.Sprintf("/secrets/%s", id)
	req, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp Secret
	if err := c.Do(req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// CreateSecret creates a secret
func (c *Client) CreateSecret(projectID, key, value, environment string, ttlHours int, metadata map[string]string, rotation *RotationConfig) (*CreateSecretResponse, error) {
	reqBody := CreateSecretRequest{
		ProjectID:   projectID,
		Key:         key,
		Value:       value,
		Environment: environment,
		TTLHours:    ttlHours,
		Metadata:    metadata,
		Rotation:    rotation,
	}

	req, err := c.newRequest(http.MethodPost, "/secrets", reqBody)
	if err != nil {
		return nil, err
	}

	var resp CreateSecretResponse
	if err := c.Do(req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// UpdateSecret updates a secret
func (c *Client) UpdateSecret(id, value string, ttlHours int, metadata map[string]string) (*Secret, error) {
	reqBody := struct {
		Value    string            `json:"value"`
		TTLHours int               `json:"ttl_hours,omitempty"`
		Metadata map[string]string `json:"metadata,omitempty"`
	}{
		Value:    value,
		TTLHours: ttlHours,
		Metadata: metadata,
	}

	path := fmt.Sprintf("/secrets/%s", id)
	req, err := c.newRequest(http.MethodPut, path, reqBody)
	if err != nil {
		return nil, err
	}

	var resp Secret
	if err := c.Do(req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// DeleteSecret deletes a secret by ID
func (c *Client) DeleteSecret(id string) error {
	path := fmt.Sprintf("/secrets/%s", id)
	req, err := c.newRequest(http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	return c.Do(req, nil)
}
