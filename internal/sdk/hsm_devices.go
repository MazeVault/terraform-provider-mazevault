package mazevault

import (
	"fmt"
	"net/http"
)

// HSMDevice represents an organization-level HSM/KMS device.
type HSMDevice struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Provider     string   `json:"provider"`
	Environments []string `json:"environments"`
	AppliesTo    []string `json:"applies_to"`
	Priority     int      `json:"priority"`
	Status       string   `json:"status"`
	HealthStatus string   `json:"health_status"`
	IsKEKDevice  bool     `json:"is_kek_device"`
}

// HSMDeviceListResponse wraps a list of HSM devices.
type HSMDeviceListResponse struct {
	Devices []HSMDevice `json:"devices"`
	Total   int         `json:"total"`
}

// CreateHSMDeviceRequest represents the request body for creating an HSM device.
type CreateHSMDeviceRequest struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	Provider     string                 `json:"provider"`
	Config       map[string]interface{} `json:"config"`
	Environments []string               `json:"environments,omitempty"`
	AppliesTo    []string               `json:"applies_to,omitempty"`
	Priority     int                    `json:"priority,omitempty"`
}

// UpdateHSMDeviceRequest represents the request body for updating an HSM device.
type UpdateHSMDeviceRequest struct {
	Name         string                 `json:"name,omitempty"`
	Description  string                 `json:"description,omitempty"`
	Config       map[string]interface{} `json:"config,omitempty"`
	Environments []string               `json:"environments,omitempty"`
	AppliesTo    []string               `json:"applies_to,omitempty"`
	Priority     int                    `json:"priority,omitempty"`
}

// ListHSMDevices lists all HSM devices for an organization.
func (c *Client) ListHSMDevices(orgID string) (*HSMDeviceListResponse, error) {
	req, err := c.newRequest(http.MethodGet, fmt.Sprintf("/api/v1/organizations/%s/hsm-devices", orgID), nil)
	if err != nil {
		return nil, err
	}
	var result HSMDeviceListResponse
	err = c.Do(req, &result)
	return &result, err
}

// GetHSMDevice retrieves a single HSM device by ID.
func (c *Client) GetHSMDevice(orgID, deviceID string) (*HSMDevice, error) {
	req, err := c.newRequest(http.MethodGet, fmt.Sprintf("/api/v1/organizations/%s/hsm-devices/%s", orgID, deviceID), nil)
	if err != nil {
		return nil, err
	}
	var result HSMDevice
	err = c.Do(req, &result)
	return &result, err
}

// CreateHSMDevice creates a new HSM device for an organization.
func (c *Client) CreateHSMDevice(orgID string, body *CreateHSMDeviceRequest) (*HSMDevice, error) {
	req, err := c.newRequest(http.MethodPost, fmt.Sprintf("/api/v1/organizations/%s/hsm-devices", orgID), body)
	if err != nil {
		return nil, err
	}
	var result HSMDevice
	err = c.Do(req, &result)
	return &result, err
}

// UpdateHSMDevice updates an existing HSM device.
func (c *Client) UpdateHSMDevice(orgID, deviceID string, body *UpdateHSMDeviceRequest) (*HSMDevice, error) {
	req, err := c.newRequest(http.MethodPut, fmt.Sprintf("/api/v1/organizations/%s/hsm-devices/%s", orgID, deviceID), body)
	if err != nil {
		return nil, err
	}
	var result HSMDevice
	err = c.Do(req, &result)
	return &result, err
}

// DeleteHSMDevice soft-deletes an HSM device.
func (c *Client) DeleteHSMDevice(orgID, deviceID string) error {
	req, err := c.newRequest(http.MethodDelete, fmt.Sprintf("/api/v1/organizations/%s/hsm-devices/%s", orgID, deviceID), nil)
	if err != nil {
		return err
	}
	return c.Do(req, nil)
}
