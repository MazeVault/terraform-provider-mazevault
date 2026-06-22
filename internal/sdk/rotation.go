package mazevault

import (
	"fmt"
	"net/http"
	"time"
)

// RotationConfigDetail represents a full rotation config as returned by the API.
type RotationConfigDetail struct {
	ID                   string     `json:"id"`
	SecretID             string     `json:"secret_id"`
	Enabled              bool       `json:"enabled"`
	Schedule             string     `json:"schedule"`
	RotationIntervalDays int        `json:"rotation_interval_days"`
	NotificationEmails   []string   `json:"notification_emails"`
	TargetEnvironment    string     `json:"target_environment"`
	BackupStrategy       string     `json:"backup_strategy"`
	TransactionMode      string     `json:"transaction_mode"`
	Status               string     `json:"status"`
	LastError            string     `json:"last_error,omitempty"`
	LastRotatedAt        *time.Time `json:"last_rotated_at,omitempty"`
	NextRotationAt       *time.Time `json:"next_rotation_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// CreateRotationConfigRequest represents the request body for creating a rotation config.
type CreateRotationConfigRequest struct {
	SecretID             string   `json:"secret_id"`
	Enabled              bool     `json:"enabled"`
	Schedule             string   `json:"schedule,omitempty"`
	RotationIntervalDays int      `json:"rotation_interval_days,omitempty"`
	NotificationEmails   []string `json:"notification_emails,omitempty"`
}

// UpdateRotationConfigRequest represents the request body for updating a rotation config.
type UpdateRotationConfigRequest struct {
	Enabled              bool     `json:"enabled"`
	Schedule             string   `json:"schedule,omitempty"`
	RotationIntervalDays int      `json:"rotation_interval_days,omitempty"`
	NotificationEmails   []string `json:"notification_emails,omitempty"`
}

// CreateRotationConfig creates a new rotation config for a secret.
func (c *Client) CreateRotationConfig(req *CreateRotationConfigRequest) (*RotationConfigDetail, error) {
	httpReq, err := c.newRequest(http.MethodPost, "/api/v1/rotation/configs", req)
	if err != nil {
		return nil, err
	}
	var result RotationConfigDetail
	if err := c.Do(httpReq, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetRotationConfig retrieves a rotation config by ID.
func (c *Client) GetRotationConfig(id string) (*RotationConfigDetail, error) {
	httpReq, err := c.newRequest(http.MethodGet, fmt.Sprintf("/api/v1/rotation/configs/%s", id), nil)
	if err != nil {
		return nil, err
	}
	var result RotationConfigDetail
	if err := c.Do(httpReq, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateRotationConfig updates an existing rotation config.
func (c *Client) UpdateRotationConfig(id string, req *UpdateRotationConfigRequest) (*RotationConfigDetail, error) {
	httpReq, err := c.newRequest(http.MethodPut, fmt.Sprintf("/api/v1/rotation/configs/%s", id), req)
	if err != nil {
		return nil, err
	}
	var result RotationConfigDetail
	if err := c.Do(httpReq, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteRotationConfig deletes a rotation config by ID.
func (c *Client) DeleteRotationConfig(id string) error {
	httpReq, err := c.newRequest(http.MethodDelete, fmt.Sprintf("/api/v1/rotation/configs/%s", id), nil)
	if err != nil {
		return err
	}
	return c.Do(httpReq, nil)
}

// RotationWorkflowDetail represents a rotation workflow (backed by same RotationConfig on the server side,
// but exposed with richer workflow step and integration data for the Terraform provider).
type RotationWorkflowDetail struct {
	ID                   string                 `json:"id"`
	SecretID             string                 `json:"secret_id"`
	Environment          string                 `json:"environment"`
	Enabled              bool                   `json:"enabled"`
	Schedule             string                 `json:"schedule"`
	RotationIntervalDays int                    `json:"rotation_interval_days"`
	RotationStrategy     string                 `json:"rotation_strategy,omitempty"`
	TargetEnvironment    string                 `json:"target_environment,omitempty"`
	GracePeriodMinutes   int                    `json:"grace_period_minutes,omitempty"`
	ResourceKind         string                 `json:"resource_kind,omitempty"`
	LinkedIntegrations   []string               `json:"linked_integrations,omitempty"`
	PostRotationActions  []PostRotationActionWF `json:"post_rotation_actions,omitempty"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
}

// PostRotationActionWF represents a single post-rotation action in a workflow.
type PostRotationActionWF struct {
	Type      string            `json:"type"`
	Config    map[string]string `json:"config,omitempty"`
	Order     int               `json:"order,omitempty"`
	OnFailure string            `json:"on_failure,omitempty"`
	// GatewayID pins this action to a specific gateway instead of using environment-based routing.
	GatewayID string `json:"gateway_id,omitempty"`
	// TargetEnvironment overrides the environment context used when resolving the target gateway.
	TargetEnvironment string `json:"target_environment,omitempty"`
}

// CreateRotationWorkflowRequest creates a rotation workflow via the rotation/configs endpoint.
type CreateRotationWorkflowRequest struct {
	SecretID             string                 `json:"secret_id"`
	Enabled              bool                   `json:"enabled"`
	Schedule             string                 `json:"schedule,omitempty"`
	RotationIntervalDays int                    `json:"rotation_interval_days,omitempty"`
	RotationStrategy     string                 `json:"rotation_strategy,omitempty"`
	TargetEnvironment    string                 `json:"target_environment,omitempty"`
	GracePeriodMinutes   int                    `json:"grace_period_minutes,omitempty"`
	ResourceKind         string                 `json:"resource_kind,omitempty"`
	LinkedIntegrations   []string               `json:"linked_integrations,omitempty"`
	PostRotationActions  []PostRotationActionWF `json:"post_rotation_actions,omitempty"`
}

// CreateRotationWorkflow creates a new rotation workflow.
func (c *Client) CreateRotationWorkflow(req *CreateRotationWorkflowRequest) (*RotationWorkflowDetail, error) {
	httpReq, err := c.newRequest(http.MethodPost, "/api/v1/rotation/configs", req)
	if err != nil {
		return nil, err
	}
	var result RotationWorkflowDetail
	if err := c.Do(httpReq, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetRotationWorkflow retrieves a rotation workflow by ID.
func (c *Client) GetRotationWorkflow(id string) (*RotationWorkflowDetail, error) {
	httpReq, err := c.newRequest(http.MethodGet, fmt.Sprintf("/api/v1/rotation/configs/%s", id), nil)
	if err != nil {
		return nil, err
	}
	var result RotationWorkflowDetail
	if err := c.Do(httpReq, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateRotationWorkflow updates a rotation workflow.
func (c *Client) UpdateRotationWorkflow(id string, req *CreateRotationWorkflowRequest) (*RotationWorkflowDetail, error) {
	httpReq, err := c.newRequest(http.MethodPut, fmt.Sprintf("/api/v1/rotation/configs/%s", id), req)
	if err != nil {
		return nil, err
	}
	var result RotationWorkflowDetail
	if err := c.Do(httpReq, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteRotationWorkflow deletes a rotation workflow by ID.
func (c *Client) DeleteRotationWorkflow(id string) error {
	httpReq, err := c.newRequest(http.MethodDelete, fmt.Sprintf("/api/v1/rotation/configs/%s", id), nil)
	if err != nil {
		return err
	}
	return c.Do(httpReq, nil)
}
