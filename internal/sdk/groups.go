package mazevault

import (
	"fmt"
	"net/http"
	"time"
)

// ConsistencyGroup represents a group of secrets monitored for cross-environment consistency.
type ConsistencyGroup struct {
	ID           string    `json:"id"`
	ProjectID    string    `json:"project_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	SecretKeys   []string  `json:"secret_keys"`
	Environments []string  `json:"environments"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateConsistencyGroupRequest represents the request body for creating a consistency group.
type CreateConsistencyGroupRequest struct {
	Name         string   `json:"name"`
	Description  string   `json:"description,omitempty"`
	SecretKeys   []string `json:"secret_keys"`
	Environments []string `json:"environments"`
}

// CreateConsistencyGroup creates a new consistency group under a project.
func (c *Client) CreateConsistencyGroup(projectID string, req *CreateConsistencyGroupRequest) (*ConsistencyGroup, error) {
	httpReq, err := c.newRequest(http.MethodPost, fmt.Sprintf("/api/v1/projects/%s/consistency-groups", projectID), req)
	if err != nil {
		return nil, err
	}
	var result ConsistencyGroup
	if err := c.Do(httpReq, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetConsistencyGroup retrieves a consistency group by ID.
func (c *Client) GetConsistencyGroup(id string) (*ConsistencyGroup, error) {
	httpReq, err := c.newRequest(http.MethodGet, fmt.Sprintf("/api/v1/consistency-groups/%s", id), nil)
	if err != nil {
		return nil, err
	}
	var result ConsistencyGroup
	if err := c.Do(httpReq, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateConsistencyGroup updates an existing consistency group.
func (c *Client) UpdateConsistencyGroup(id string, req *CreateConsistencyGroupRequest) (*ConsistencyGroup, error) {
	httpReq, err := c.newRequest(http.MethodPut, fmt.Sprintf("/api/v1/consistency-groups/%s", id), req)
	if err != nil {
		return nil, err
	}
	var result ConsistencyGroup
	if err := c.Do(httpReq, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteConsistencyGroup deletes a consistency group by ID.
func (c *Client) DeleteConsistencyGroup(id string) error {
	httpReq, err := c.newRequest(http.MethodDelete, fmt.Sprintf("/api/v1/consistency-groups/%s", id), nil)
	if err != nil {
		return err
	}
	return c.Do(httpReq, nil)
}

// IntegrationGroup represents a named group of integrations mapped to environments.
// Backend routes for integration groups are surfaced as project-scoped sub-resources.
type IntegrationGroup struct {
	ID          string               `json:"id"`
	ProjectID   string               `json:"project_id"`
	Name        string               `json:"name"`
	Description string               `json:"description,omitempty"`
	Mappings    []IntegrationMapping `json:"mappings,omitempty"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

// IntegrationMapping represents a single integration → environment mapping within a group.
type IntegrationMapping struct {
	IntegrationID string `json:"integration_id"`
	Environment   string `json:"environment"`
	Priority      int    `json:"priority,omitempty"`
}

// CreateIntegrationGroupRequest represents the request body for creating an integration group.
type CreateIntegrationGroupRequest struct {
	Name        string               `json:"name"`
	Description string               `json:"description,omitempty"`
	Mappings    []IntegrationMapping `json:"mappings,omitempty"`
}

// CreateIntegrationGroup creates a new integration group under a project.
func (c *Client) CreateIntegrationGroup(projectID string, req *CreateIntegrationGroupRequest) (*IntegrationGroup, error) {
	httpReq, err := c.newRequest(http.MethodPost, fmt.Sprintf("/api/v1/projects/%s/integration-groups", projectID), req)
	if err != nil {
		return nil, err
	}
	var result IntegrationGroup
	if err := c.Do(httpReq, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetIntegrationGroup retrieves an integration group by ID.
func (c *Client) GetIntegrationGroup(id string) (*IntegrationGroup, error) {
	httpReq, err := c.newRequest(http.MethodGet, fmt.Sprintf("/api/v1/integration-groups/%s", id), nil)
	if err != nil {
		return nil, err
	}
	var result IntegrationGroup
	if err := c.Do(httpReq, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateIntegrationGroup updates an existing integration group.
func (c *Client) UpdateIntegrationGroup(id string, req *CreateIntegrationGroupRequest) (*IntegrationGroup, error) {
	httpReq, err := c.newRequest(http.MethodPut, fmt.Sprintf("/api/v1/integration-groups/%s", id), req)
	if err != nil {
		return nil, err
	}
	var result IntegrationGroup
	if err := c.Do(httpReq, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteIntegrationGroup deletes an integration group by ID.
func (c *Client) DeleteIntegrationGroup(id string) error {
	httpReq, err := c.newRequest(http.MethodDelete, fmt.Sprintf("/api/v1/integration-groups/%s", id), nil)
	if err != nil {
		return err
	}
	return c.Do(httpReq, nil)
}

// ConsistencyStatus represents the aggregated consistency state for a project.
type ConsistencyStatus struct {
	Status       string   `json:"status"`
	MissingCount int      `json:"missing_count"`
	Issues       []string `json:"issues"`
}

// GetConsistencyStatus returns the aggregated consistency status for a project by
// listing all consistency groups and deriving a status from the response.
func (c *Client) GetConsistencyStatus(projectID string) (*ConsistencyStatus, error) {
	httpReq, err := c.newRequest(http.MethodGet, fmt.Sprintf("/api/v1/projects/%s/consistency/status", projectID), nil)
	if err != nil {
		return nil, err
	}
	var groups []ConsistencyGroup
	if err := c.Do(httpReq, &groups); err != nil {
		return nil, err
	}
	status := &ConsistencyStatus{Status: "healthy", Issues: []string{}}
	status.MissingCount = len(groups)
	if len(groups) > 0 {
		status.Status = "warning"
		for _, g := range groups {
			status.Issues = append(status.Issues, fmt.Sprintf("consistency group %q has %d secret keys across %d environments", g.Name, len(g.SecretKeys), len(g.Environments)))
		}
	}
	return status, nil
}
