package mazevault

import (
	"fmt"
	"net/http"
)

// SyncRule represents a synchronization rule between MazeVault and an external provider.
type SyncRule struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	IntegrationID        string `json:"integration_id"`
	ProjectID            string `json:"project_id"`
	TargetEnvironment    string `json:"target_environment,omitempty"`
	SourcePath           string `json:"source_path,omitempty"`
	KeyTransformTemplate string `json:"key_transform_template,omitempty"`
	ConflictStrategy     string `json:"conflict_strategy,omitempty"`
	SyncDirection        string `json:"sync_direction,omitempty"`
	SyncMode             string `json:"sync_mode,omitempty"`
	CreatedAt            string `json:"created_at"`
	UpdatedAt            string `json:"updated_at"`
}

// CreateSyncRuleRequest is the request body for creating a sync rule.
type CreateSyncRuleRequest struct {
	Name                 string `json:"name"`
	IntegrationID        string `json:"integration_id"`
	ProjectID            string `json:"project_id"`
	TargetEnvironment    string `json:"target_environment,omitempty"`
	SourcePath           string `json:"source_path,omitempty"`
	KeyTransformTemplate string `json:"key_transform_template,omitempty"`
	ConflictStrategy     string `json:"conflict_strategy,omitempty"`
	SyncDirection        string `json:"sync_direction,omitempty"`
	SyncMode             string `json:"sync_mode,omitempty"`
}

// ListSyncRulesResponse is returned by the list endpoint.
type ListSyncRulesResponse struct {
	Rules []SyncRule `json:"rules"`
}

// CreateSyncRule creates a new sync rule for a project.
func (c *Client) CreateSyncRule(req *CreateSyncRuleRequest) (*SyncRule, error) {
	path := fmt.Sprintf("/projects/%s/sync-rules", req.ProjectID)
	r, err := c.newRequest(http.MethodPost, path, req)
	if err != nil {
		return nil, err
	}
	var resp SyncRule
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetSyncRule retrieves a sync rule by project ID and rule ID.
func (c *Client) GetSyncRule(projectID, ruleID string) (*SyncRule, error) {
	path := fmt.Sprintf("/projects/%s/sync-rules/%s", projectID, ruleID)
	r, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var resp SyncRule
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateSyncRule updates a sync rule.
func (c *Client) UpdateSyncRule(projectID, ruleID string, req *CreateSyncRuleRequest) (*SyncRule, error) {
	path := fmt.Sprintf("/projects/%s/sync-rules/%s", projectID, ruleID)
	r, err := c.newRequest(http.MethodPut, path, req)
	if err != nil {
		return nil, err
	}
	var resp SyncRule
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteSyncRule deletes a sync rule.
func (c *Client) DeleteSyncRule(projectID, ruleID string) error {
	path := fmt.Sprintf("/projects/%s/sync-rules/%s", projectID, ruleID)
	r, err := c.newRequest(http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	return c.Do(r, nil)
}
