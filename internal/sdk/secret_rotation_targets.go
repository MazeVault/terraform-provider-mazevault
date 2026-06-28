package mazevault

import (
	"fmt"
	"net/http"
	"time"
)

// SecretRotationTarget represents a structured deployment target in a secret's
// rotation pipeline. Mirrors models.RotationTarget on the backend.
//
// Targets are ordered by priority (lower = executed first) and describe WHERE
// and HOW a rotated secret value must be delivered after each rotation:
// Kubernetes secret, database password rotation, agent file sync, DevOps
// variable, or cloud vault (AWS SM / GCP SM / OCI Vault).
type SecretRotationTarget struct {
	ID                 string                 `json:"id"`
	RotationResourceID string                 `json:"rotation_resource_id"`
	TargetType         string                 `json:"target_type"`
	TargetRole         string                 `json:"target_role"`
	BindingRef         string                 `json:"binding_ref,omitempty"`
	Priority           int                    `json:"priority"`
	Enabled            bool                   `json:"enabled"`
	SyncOnSuccess      bool                   `json:"sync_on_success"`
	VerificationMode   string                 `json:"verification_mode,omitempty"`
	RecoveryPolicy     string                 `json:"recovery_policy,omitempty"`
	Config             map[string]interface{} `json:"config"`
	LastDeliveryStatus string                 `json:"last_delivery_status,omitempty"`
	LastDeliveryAt     *time.Time             `json:"last_delivery_at,omitempty"`
	LastDeliveryError  string                 `json:"last_delivery_error,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

// CreateSecretRotationTargetRequest is the request body for creating a deployment target.
type CreateSecretRotationTargetRequest struct {
	TargetType       string                 `json:"target_type"`
	TargetRole       string                 `json:"target_role,omitempty"`
	Priority         int                    `json:"priority,omitempty"`
	Enabled          bool                   `json:"enabled"`
	SyncOnSuccess    bool                   `json:"sync_on_success"`
	VerificationMode string                 `json:"verification_mode,omitempty"`
	RecoveryPolicy   string                 `json:"recovery_policy,omitempty"`
	BindingRef       string                 `json:"binding_ref,omitempty"`
	Config           map[string]interface{} `json:"config"`
}

// UpdateSecretRotationTargetRequest is the request body for updating a deployment target.
// All fields are optional (partial-update semantics).
type UpdateSecretRotationTargetRequest struct {
	TargetType       *string                `json:"target_type,omitempty"`
	TargetRole       *string                `json:"target_role,omitempty"`
	Priority         *int                   `json:"priority,omitempty"`
	Enabled          *bool                  `json:"enabled,omitempty"`
	SyncOnSuccess    *bool                  `json:"sync_on_success,omitempty"`
	VerificationMode *string                `json:"verification_mode,omitempty"`
	RecoveryPolicy   *string                `json:"recovery_policy,omitempty"`
	BindingRef       *string                `json:"binding_ref,omitempty"`
	Config           map[string]interface{} `json:"config,omitempty"`
}

// ListSecretRotationTargets returns all deployment targets for the given secret,
// ordered by priority ASC.
func (c *Client) ListSecretRotationTargets(secretID string) ([]SecretRotationTarget, error) {
	req, err := c.newRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/secrets/%s/rotation/targets", secretID), nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Targets []SecretRotationTarget `json:"targets"`
	}
	if err := c.Do(req, &resp); err != nil {
		return nil, err
	}
	return resp.Targets, nil
}

// GetSecretRotationTarget returns a single deployment target by its ID.
// Because no direct GET-by-ID endpoint exists, this iterates the list and
// returns the first match. When the target is not found it returns an
// *APIError with StatusCode 404 so callers can use IsNotFoundError().
func (c *Client) GetSecretRotationTarget(secretID, targetID string) (*SecretRotationTarget, error) {
	targets, err := c.ListSecretRotationTargets(secretID)
	if err != nil {
		return nil, err
	}
	for i := range targets {
		if targets[i].ID == targetID {
			return &targets[i], nil
		}
	}
	return nil, &APIError{
		StatusCode: 404,
		Message:    fmt.Sprintf("rotation target %s not found for secret %s", targetID, secretID),
	}
}

// CreateSecretRotationTarget adds a new deployment target to the secret's pipeline.
func (c *Client) CreateSecretRotationTarget(secretID string, payload *CreateSecretRotationTargetRequest) (*SecretRotationTarget, error) {
	req, err := c.newRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/secrets/%s/rotation/targets", secretID), payload)
	if err != nil {
		return nil, err
	}
	var result SecretRotationTarget
	if err := c.Do(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateSecretRotationTarget partially updates an existing deployment target.
func (c *Client) UpdateSecretRotationTarget(secretID, targetID string, payload *UpdateSecretRotationTargetRequest) (*SecretRotationTarget, error) {
	req, err := c.newRequest(http.MethodPut,
		fmt.Sprintf("/api/v1/secrets/%s/rotation/targets/%s", secretID, targetID), payload)
	if err != nil {
		return nil, err
	}
	var result SecretRotationTarget
	if err := c.Do(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteSecretRotationTarget removes a deployment target from the secret's pipeline.
func (c *Client) DeleteSecretRotationTarget(secretID, targetID string) error {
	req, err := c.newRequest(http.MethodDelete,
		fmt.Sprintf("/api/v1/secrets/%s/rotation/targets/%s", secretID, targetID), nil)
	if err != nil {
		return err
	}
	return c.Do(req, nil)
}

// SyncSecretRotationTarget queues an immediate manual sync for a deployment target.
func (c *Client) SyncSecretRotationTarget(secretID, targetID string) (*SecretRotationTarget, error) {
	req, err := c.newRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/secrets/%s/rotation/targets/%s/sync", secretID, targetID), nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Target  SecretRotationTarget `json:"target"`
		Message string               `json:"message"`
	}
	if err := c.Do(req, &resp); err != nil {
		return nil, err
	}
	return &resp.Target, nil
}
