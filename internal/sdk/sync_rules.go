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

// ListSyncRules returns all sync rules for a project.
func (c *Client) ListSyncRules(projectID string) ([]SyncRule, error) {
	path := fmt.Sprintf("/api/v1/projects/%s/sync-rules", projectID)
	r, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var resp ListSyncRulesResponse
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return resp.Rules, nil
}

// CreateSyncRule creates a new sync rule for a project.
func (c *Client) CreateSyncRule(req *CreateSyncRuleRequest) (*SyncRule, error) {
	path := fmt.Sprintf("/api/v1/projects/%s/sync-rules", req.ProjectID)
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
// The backend has no dedicated GET-by-ID endpoint; this function uses ListSyncRules
// and filters by ID (O(n) in the number of rules per project).
func (c *Client) GetSyncRule(projectID, ruleID string) (*SyncRule, error) {
	rules, err := c.ListSyncRules(projectID)
	if err != nil {
		return nil, err
	}
	for i := range rules {
		if rules[i].ID == ruleID {
			return &rules[i], nil
		}
	}
	return nil, nil
}

// UpdateSyncRule updates a sync rule.
// Backend route: PUT /api/v1/projects/sync-rules/:ruleId (no project ID in URL).
func (c *Client) UpdateSyncRule(projectID, ruleID string, req *CreateSyncRuleRequest) (*SyncRule, error) {
	path := fmt.Sprintf("/api/v1/projects/sync-rules/%s", ruleID)
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
// Backend route: DELETE /api/v1/projects/sync-rules/:ruleId (no project ID in URL).
func (c *Client) DeleteSyncRule(projectID, ruleID string) error {
	path := fmt.Sprintf("/api/v1/projects/sync-rules/%s", ruleID)
	r, err := c.newRequest(http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	return c.Do(r, nil)
}
