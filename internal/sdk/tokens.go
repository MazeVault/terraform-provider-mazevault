package mazevault

import (
	"fmt"
	"net/http"
	"time"
)

// ManagedToken represents an externally-issued API token tracked by MazeVault.
type ManagedToken struct {
	ID                   string    `json:"id"`
	OrganizationID       string    `json:"organization_id"`
	ProjectID            string    `json:"project_id,omitempty"`
	Name                 string    `json:"name"`
	Description          string    `json:"description,omitempty"`
	Provider             string    `json:"provider"`
	StorageMode          string    `json:"storage_mode"`
	ExternalProviderType string    `json:"external_provider_type,omitempty"`
	ExpiresAt            time.Time `json:"expires_at"`
	LeadTimeDays         int       `json:"lead_time_days"`
	Status               string    `json:"status"`
	IsExpired            bool      `json:"is_expired"`
	DaysUntilExpiry      int       `json:"days_until_expiry"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// CreateManagedTokenRequest is the request body for creating a managed token.
type CreateManagedTokenRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	Provider     string `json:"provider,omitempty"`
	Value        string `json:"value"`
	ProjectID    string `json:"project_id,omitempty"`
	ExpiresAt    string `json:"expires_at"`
	LeadTimeDays *int   `json:"lead_time_days,omitempty"`
}

// UpdateManagedTokenRequest is the request body for updating a managed token's metadata.
type UpdateManagedTokenRequest struct {
	Name         *string `json:"name,omitempty"`
	Description  *string `json:"description,omitempty"`
	Provider     *string `json:"provider,omitempty"`
	ExpiresAt    *string `json:"expires_at,omitempty"`
	LeadTimeDays *int    `json:"lead_time_days,omitempty"`
	Status       *string `json:"status,omitempty"`
}

// ListManagedTokens lists managed tokens for the organization.
func (c *Client) ListManagedTokens() ([]ManagedToken, error) {
	r, err := c.newRequest(http.MethodGet, "/api/v1/tokens", nil)
	if err != nil {
		return nil, err
	}
	// The tokens endpoint returns a paginated envelope.
	var resp struct {
		Data  []ManagedToken `json:"data"`
		Items []ManagedToken `json:"items"`
	}
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) > 0 {
		return resp.Data, nil
	}
	return resp.Items, nil
}

// CreateManagedToken creates a new managed token.
func (c *Client) CreateManagedToken(req *CreateManagedTokenRequest) (*ManagedToken, error) {
	r, err := c.newRequest(http.MethodPost, "/api/v1/tokens", req)
	if err != nil {
		return nil, err
	}
	var resp ManagedToken
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetManagedToken retrieves a managed token by ID.
func (c *Client) GetManagedToken(id string) (*ManagedToken, error) {
	r, err := c.newRequest(http.MethodGet, "/api/v1/tokens/"+id, nil)
	if err != nil {
		return nil, err
	}
	var resp ManagedToken
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateManagedToken updates a managed token's metadata.
func (c *Client) UpdateManagedToken(id string, req *UpdateManagedTokenRequest) (*ManagedToken, error) {
	r, err := c.newRequest(http.MethodPut, fmt.Sprintf("/api/v1/tokens/%s", id), req)
	if err != nil {
		return nil, err
	}
	var resp ManagedToken
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteManagedToken deletes a managed token.
func (c *Client) DeleteManagedToken(id string) error {
	r, err := c.newRequest(http.MethodDelete, "/api/v1/tokens/"+id, nil)
	if err != nil {
		return err
	}
	return c.Do(r, nil)
}
