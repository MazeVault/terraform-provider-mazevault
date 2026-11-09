package mazevault

import (
	"fmt"
	"net/http"
	"time"
)

// CertificateRotationTarget is a deployment target of a certificate: where the
// (renewed) certificate is delivered and how the consumer is reloaded. Mirrors
// models.CertificateRotationTarget on the backend.
//
// Target types: secret_manager, kubernetes_secret, database, agent_keystore,
// agent_file, agent_truststore, agent_container_file, weblogic.
type CertificateRotationTarget struct {
	ID                    string                 `json:"id"`
	CertificateID         string                 `json:"certificate_id"`
	TargetType            string                 `json:"target_type"`
	IntegrationID         *string                `json:"integration_id,omitempty"`
	AgentID               *string                `json:"agent_id,omitempty"`
	DatabaseIntegrationID *string                `json:"database_integration_id,omitempty"`
	ContentMode           string                 `json:"content_mode"`
	TargetConfig          map[string]interface{} `json:"target_config"`
	FormatMode            string                 `json:"format_mode"`
	ConflictStrategy      string                 `json:"conflict_strategy"`
	Enabled               bool                   `json:"enabled"`
	SyncOnRenewal         bool                   `json:"sync_on_renewal"`
	Priority              int                    `json:"priority"`
	DeploymentWindow      map[string]interface{} `json:"deployment_window,omitempty"`
	LastSyncAt            *time.Time             `json:"last_sync_at,omitempty"`
	LastSyncStatus        string                 `json:"last_sync_status,omitempty"`
	LastSyncError         string                 `json:"last_sync_error,omitempty"`
}

// CreateCertificateRotationTargetRequest is the body of POST /certificates/{id}/targets.
type CreateCertificateRotationTargetRequest struct {
	TargetType            string                 `json:"target_type"`
	IntegrationID         *string                `json:"integration_id,omitempty"`
	AgentID               *string                `json:"agent_id,omitempty"`
	DatabaseIntegrationID *string                `json:"database_integration_id,omitempty"`
	ContentMode           string                 `json:"content_mode"`
	TargetConfig          map[string]interface{} `json:"target_config"`
	FormatMode            string                 `json:"format_mode,omitempty"`
	ConflictStrategy      string                 `json:"conflict_strategy,omitempty"`
	Enabled               *bool                  `json:"enabled,omitempty"`
	SyncOnRenewal         *bool                  `json:"sync_on_renewal,omitempty"`
	Priority              int                    `json:"priority,omitempty"`
	DeploymentWindow      map[string]interface{} `json:"deployment_window,omitempty"`
}

// UpdateCertificateRotationTargetRequest is the body of PUT /certificates/{id}/targets/{targetId}.
// Partial-update semantics: nil fields are left untouched.
type UpdateCertificateRotationTargetRequest struct {
	ContentMode           string                 `json:"content_mode,omitempty"`
	TargetConfig          map[string]interface{} `json:"target_config,omitempty"`
	FormatMode            string                 `json:"format_mode,omitempty"`
	ConflictStrategy      string                 `json:"conflict_strategy,omitempty"`
	IntegrationID         *string                `json:"integration_id,omitempty"`
	AgentID               *string                `json:"agent_id,omitempty"`
	DatabaseIntegrationID *string                `json:"database_integration_id,omitempty"`
	Enabled               *bool                  `json:"enabled,omitempty"`
	SyncOnRenewal         *bool                  `json:"sync_on_renewal,omitempty"`
	Priority              *int                   `json:"priority,omitempty"`
	// DeploymentWindow is a pointer so an explicit empty object {} can be sent to
	// clear a configured window: a plain map with omitempty would be dropped from
	// the JSON when empty, leaving the stored window untouched. A nil pointer
	// omits the field (no change); a pointer to an empty map clears the window.
	DeploymentWindow *map[string]interface{} `json:"deployment_window,omitempty"`
}

// DeploymentTargetSuggestion is a proposed target derived from discovery evidence
// (GET /certificates/{id}/targets/suggestions).
type DeploymentTargetSuggestion struct {
	Key          string                 `json:"key"`
	Source       string                 `json:"source"`
	Reason       string                 `json:"reason"`
	TargetType   string                 `json:"target_type"`
	AgentID      string                 `json:"agent_id"`
	AgentName    string                 `json:"agent_name,omitempty"`
	Hostname     string                 `json:"hostname,omitempty"`
	ContentMode  string                 `json:"content_mode"`
	TargetConfig map[string]interface{} `json:"target_config"`
	ConsumerType string                 `json:"consumer_type,omitempty"`
	ReloadHint   string                 `json:"reload_hint,omitempty"`
}

// ListCertificateRotationTargets returns the deployment targets of a certificate.
func (c *Client) ListCertificateRotationTargets(certificateID string) ([]CertificateRotationTarget, error) {
	req, err := c.newRequest(http.MethodGet, fmt.Sprintf("/api/v1/certificates/%s/targets", certificateID), nil)
	if err != nil {
		return nil, err
	}
	var out []CertificateRotationTarget
	if err := c.Do(req, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetCertificateRotationTarget returns one deployment target.
func (c *Client) GetCertificateRotationTarget(certificateID, targetID string) (*CertificateRotationTarget, error) {
	req, err := c.newRequest(http.MethodGet, fmt.Sprintf("/api/v1/certificates/%s/targets/%s", certificateID, targetID), nil)
	if err != nil {
		return nil, err
	}
	var out CertificateRotationTarget
	if err := c.Do(req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateCertificateRotationTarget adds a deployment target to a certificate.
func (c *Client) CreateCertificateRotationTarget(certificateID string, payload *CreateCertificateRotationTargetRequest) (*CertificateRotationTarget, error) {
	req, err := c.newRequest(http.MethodPost, fmt.Sprintf("/api/v1/certificates/%s/targets", certificateID), payload)
	if err != nil {
		return nil, err
	}
	var out CertificateRotationTarget
	if err := c.Do(req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateCertificateRotationTarget partially updates a deployment target.
func (c *Client) UpdateCertificateRotationTarget(certificateID, targetID string, payload *UpdateCertificateRotationTargetRequest) (*CertificateRotationTarget, error) {
	req, err := c.newRequest(http.MethodPut, fmt.Sprintf("/api/v1/certificates/%s/targets/%s", certificateID, targetID), payload)
	if err != nil {
		return nil, err
	}
	var out CertificateRotationTarget
	if err := c.Do(req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteCertificateRotationTarget removes a deployment target.
func (c *Client) DeleteCertificateRotationTarget(certificateID, targetID string) error {
	req, err := c.newRequest(http.MethodDelete, fmt.Sprintf("/api/v1/certificates/%s/targets/%s", certificateID, targetID), nil)
	if err != nil {
		return err
	}
	return c.Do(req, nil)
}

// GetCertificateTargetSuggestions returns deployment targets proposed from discovery evidence.
func (c *Client) GetCertificateTargetSuggestions(certificateID string) ([]DeploymentTargetSuggestion, error) {
	req, err := c.newRequest(http.MethodGet, fmt.Sprintf("/api/v1/certificates/%s/targets/suggestions", certificateID), nil)
	if err != nil {
		return nil, err
	}
	var out []DeploymentTargetSuggestion
	if err := c.Do(req, &out); err != nil {
		return nil, err
	}
	return out, nil
}
