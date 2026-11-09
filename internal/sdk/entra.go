package mazevault

import (
	"fmt"
	"net/http"
	"time"
)

// EntraScopeMode values accepted by the project Entra integration scope.
const (
	EntraScopeModeLegacyAll    = "legacy_all"
	EntraScopeModeSelectedApps = "selected_apps"
)

// EntraBindingKind values accepted for selected application bindings.
const (
	EntraBindingKindAppRegistration  = "app_registration"
	EntraBindingKindServicePrincipal = "service_principal"
)

// EntraSelectedAppBinding describes one Entra application selected into a
// project-scoped Entra integration (PUT /entra/:integrationId/scope).
type EntraSelectedAppBinding struct {
	BindingKind string `json:"binding_kind"`
	AppID       string `json:"app_id"`
	ObjectID    string `json:"object_id"`
	TenantID    string `json:"tenant_id,omitempty"`
	DisplayName string `json:"display_name"`
}

// EntraIntegrationScopeRequest replaces the selected application bindings of
// an Entra integration.
type EntraIntegrationScopeRequest struct {
	ScopeMode           string                    `json:"scope_mode"`
	SelectedAppBindings []EntraSelectedAppBinding `json:"selected_app_bindings"`
}

// EntraSelectedAppBindingView is a persisted binding as returned by the backend.
type EntraSelectedAppBindingView struct {
	ID              string `json:"id"`
	ProjectID       string `json:"project_id"`
	IntegrationID   string `json:"integration_id"`
	BindingKind     string `json:"binding_kind"`
	AppID           string `json:"app_id"`
	ObjectID        string `json:"object_id"`
	TenantID        string `json:"tenant_id"`
	DisplayName     string `json:"display_name"`
	SelectionSource string `json:"selection_source"`
	SelectedAt      string `json:"selected_at"`
}

// EntraIntegrationScopeState is the current scope of an Entra integration.
type EntraIntegrationScopeState struct {
	IntegrationID       string                        `json:"integration_id"`
	ProjectID           string                        `json:"project_id"`
	ScopeMode           string                        `json:"scope_mode"`
	SelectedAppBindings []EntraSelectedAppBindingView `json:"selected_app_bindings"`
}

// EntraAppRegistration is an Entra application synchronised into a project.
type EntraAppRegistration struct {
	ID                 string     `json:"id"`
	ProjectID          string     `json:"project_id"`
	IntegrationID      string     `json:"integration_id"`
	AppID              string     `json:"app_id"`
	ObjectID           string     `json:"object_id"`
	TenantID           string     `json:"tenant_id"`
	DisplayName        string     `json:"display_name"`
	Description        string     `json:"description"`
	SignInAudience     string     `json:"sign_in_audience"`
	ManagedByMazeVault bool       `json:"managed_by_mazevault"`
	SyncStatus         string     `json:"sync_status"`
	SyncError          string     `json:"sync_error"`
	LastSyncedAt       *time.Time `json:"last_synced_at"`
}

// EntraCredential is a client secret or certificate credential of an Entra
// application. Secret values are never returned by the API.
type EntraCredential struct {
	ID                string     `json:"id"`
	AppRegistrationID string     `json:"app_registration_id"`
	CredentialType    string     `json:"credential_type"`
	KeyID             string     `json:"key_id"`
	DisplayName       string     `json:"display_name"`
	ValueHint         string     `json:"value_hint"`
	StartDate         time.Time  `json:"start_date"`
	EndDate           time.Time  `json:"end_date"`
	Status            string     `json:"status"`
	IsPrimary         bool       `json:"is_primary"`
	LifecycleState    string     `json:"lifecycle_state"`
	RotationEnabled   bool       `json:"rotation_enabled"`
	LastRotatedAt     *time.Time `json:"last_rotated_at"`
}

// GetEntraIntegrationScope returns the selected application scope of a
// project-scoped Entra integration.
func (c *Client) GetEntraIntegrationScope(integrationID string) (*EntraIntegrationScopeState, error) {
	r, err := c.newRequest(http.MethodGet, fmt.Sprintf("/api/v1/entra/%s/scope", integrationID), nil)
	if err != nil {
		return nil, err
	}
	var resp EntraIntegrationScopeState
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateEntraIntegrationScope replaces the selected application bindings of an
// Entra integration. The backend validates that selected applications are not
// already bound in another project.
func (c *Client) UpdateEntraIntegrationScope(integrationID string, req *EntraIntegrationScopeRequest) (*EntraIntegrationScopeState, error) {
	r, err := c.newRequest(http.MethodPut, fmt.Sprintf("/api/v1/entra/%s/scope", integrationID), req)
	if err != nil {
		return nil, err
	}
	var resp EntraIntegrationScopeState
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SyncEntraAppRegistrations pulls application registrations and their
// credentials for an Entra integration from Microsoft Graph into MazeVault.
func (c *Client) SyncEntraAppRegistrations(integrationID string) error {
	r, err := c.newRequest(http.MethodPost, fmt.Sprintf("/api/v1/entra/%s/sync", integrationID), nil)
	if err != nil {
		return err
	}
	return c.Do(r, nil)
}

// ListEntraAppRegistrations lists synchronised application registrations of an
// Entra integration.
func (c *Client) ListEntraAppRegistrations(integrationID string) ([]EntraAppRegistration, error) {
	r, err := c.newRequest(http.MethodGet, fmt.Sprintf("/api/v1/entra/%s/apps", integrationID), nil)
	if err != nil {
		return nil, err
	}
	var resp []EntraAppRegistration
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// ListEntraAppCredentials lists credentials of a synchronised application
// registration (MazeVault app registration ID, not the Entra app/client ID).
func (c *Client) ListEntraAppCredentials(appRegistrationID string) ([]EntraCredential, error) {
	r, err := c.newRequest(http.MethodGet, fmt.Sprintf("/api/v1/entra/apps/%s/credentials", appRegistrationID), nil)
	if err != nil {
		return nil, err
	}
	var resp []EntraCredential
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}
