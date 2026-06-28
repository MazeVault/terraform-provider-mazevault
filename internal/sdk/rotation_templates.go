package mazevault

import (
	"fmt"
	"net/http"
)

// RotationConfigTemplate is a reusable policy template for rotation configs.
type RotationConfigTemplate struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	Description          string `json:"description,omitempty"`
	OrganizationID       string `json:"organization_id,omitempty"`
	IsDefault            bool   `json:"is_default"`
	RotationIntervalDays int    `json:"rotation_interval_days,omitempty"`
	LeadTimeDays         int    `json:"lead_time_days,omitempty"`
	GracePeriodDays      int    `json:"grace_period_days,omitempty"`
	MaxRetryAttempts     int    `json:"max_retry_attempts,omitempty"`
	TimeoutMinutes       int    `json:"timeout_minutes,omitempty"`
	CreatedAt            string `json:"created_at"`
	UpdatedAt            string `json:"updated_at"`
}

// CreateRotationTemplateRequest is the request body for creating or updating a template.
type CreateRotationTemplateRequest struct {
	Name                 string `json:"name"`
	Description          string `json:"description,omitempty"`
	OrganizationID       string `json:"organization_id,omitempty"`
	IsDefault            bool   `json:"is_default,omitempty"`
	RotationIntervalDays int    `json:"rotation_interval_days,omitempty"`
	LeadTimeDays         int    `json:"lead_time_days,omitempty"`
	GracePeriodDays      int    `json:"grace_period_days,omitempty"`
	MaxRetryAttempts     int    `json:"max_retry_attempts,omitempty"`
	TimeoutMinutes       int    `json:"timeout_minutes,omitempty"`
}

// ListRotationTemplatesResponse wraps the list response.
type ListRotationTemplatesResponse struct {
	Templates []RotationConfigTemplate `json:"templates"`
}

// CreateRotationTemplate creates a new rotation config template.
func (c *Client) CreateRotationTemplate(req *CreateRotationTemplateRequest) (*RotationConfigTemplate, error) {
	r, err := c.newRequest(http.MethodPost, "/api/v1/rotation/templates", req)
	if err != nil {
		return nil, err
	}
	var resp RotationConfigTemplate
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetRotationTemplate retrieves a rotation template by ID.
func (c *Client) GetRotationTemplate(id string) (*RotationConfigTemplate, error) {
	path := fmt.Sprintf("/api/v1/rotation/templates/%s", id)
	r, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var resp RotationConfigTemplate
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateRotationTemplate updates a rotation template.
func (c *Client) UpdateRotationTemplate(id string, req *CreateRotationTemplateRequest) (*RotationConfigTemplate, error) {
	path := fmt.Sprintf("/api/v1/rotation/templates/%s", id)
	r, err := c.newRequest(http.MethodPut, path, req)
	if err != nil {
		return nil, err
	}
	var resp RotationConfigTemplate
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteRotationTemplate deletes a rotation template.
func (c *Client) DeleteRotationTemplate(id string) error {
	path := fmt.Sprintf("/api/v1/rotation/templates/%s", id)
	r, err := c.newRequest(http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	return c.Do(r, nil)
}
