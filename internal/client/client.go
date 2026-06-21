package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

// Client wraps the resty client for MazeVault API interactions
type Client struct {
	Resty   *resty.Client
	BaseURL string
	Token   string
}

// Config holds the configuration for the API client
type Config struct {
	BaseURL      string
	Token        string
	ClientID     string
	ClientSecret string
	Timeout      time.Duration
	SkipTLS      bool
}

// NewClient creates a new API client
func NewClient(cfg *Config) (*Client, error) {
	client := resty.New()
	client.SetBaseURL(cfg.BaseURL)
	client.SetTimeout(cfg.Timeout)
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: cfg.SkipTLS})

	c := &Client{
		Resty:   client,
		BaseURL: cfg.BaseURL,
		Token:   cfg.Token,
	}

	// If token is provided directly, use it
	if cfg.Token != "" {
		client.SetAuthToken(cfg.Token)
	} else if cfg.ClientID != "" && cfg.ClientSecret != "" {
		// If client credentials are provided, authenticate to get a token
		// This is a simplified implementation - in a real scenario we'd handle token expiration/refresh
		token, err := c.authenticate(cfg.ClientID, cfg.ClientSecret)
		if err != nil {
			return nil, fmt.Errorf("authentication failed: %w", err)
		}
		c.Token = token
		client.SetAuthToken(token)
	}

	return c, nil
}

// authenticate performs the login request using client credentials
func (c *Client) authenticate(clientID, clientSecret string) (string, error) {
	var result struct {
		AccessToken string `json:"access_token"`
	}

	// Assuming there's an endpoint for service account login or similar
	// For now, we'll use the standard login endpoint but this might need adjustment
	// based on the actual backend implementation for service accounts
	resp, err := c.Resty.R().
		SetBody(map[string]string{
			"client_id":     clientID,
			"client_secret": clientSecret,
			"grant_type":    "client_credentials",
		}).
		SetResult(&result).
		Post("/api/v1/auth/token")

	if err != nil {
		return "", err
	}

	if resp.IsError() {
		return "", fmt.Errorf("status: %s, body: %s", resp.Status(), resp.String())
	}

	return result.AccessToken, nil
}

// Organization represents a MazeVault organization
type Organization struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateOrganization creates a new organization
func (c *Client) CreateOrganization(ctx context.Context, org *Organization) error {
	resp, err := c.Resty.R().
		SetContext(ctx).
		SetBody(org).
		SetResult(org).
		Post("/api/v1/organizations")

	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("failed to create organization: %s", resp.Status())
	}

	return nil
}

// ReadOrganization retrieves an organization by ID
func (c *Client) ReadOrganization(ctx context.Context, id string) (*Organization, error) {
	var org Organization
	resp, err := c.Resty.R().
		SetContext(ctx).
		SetResult(&org).
		Get(fmt.Sprintf("/api/v1/organizations/%s", id))

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() == http.StatusNotFound {
		return nil, nil
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to read organization: %s", resp.Status())
	}

	return &org, nil
}

// UpdateOrganization updates an existing organization
func (c *Client) UpdateOrganization(ctx context.Context, org *Organization) error {
	resp, err := c.Resty.R().
		SetContext(ctx).
		SetBody(org).
		SetResult(org).
		Put(fmt.Sprintf("/api/v1/organizations/%s", org.ID))

	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("failed to update organization: %s", resp.Status())
	}

	return nil
}

// DeleteOrganization deletes an organization
func (c *Client) DeleteOrganization(ctx context.Context, id string) error {
	resp, err := c.Resty.R().
		SetContext(ctx).
		Delete(fmt.Sprintf("/api/v1/organizations/%s", id))

	if err != nil {
		return err
	}

	if resp.IsError() && resp.StatusCode() != http.StatusNotFound {
		return fmt.Errorf("failed to delete organization: %s", resp.Status())
	}

	return nil
}

// Project represents a MazeVault project
type Project struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	Name           string    `json:"name"`
	Environment    string    `json:"environment"`
	CreatedAt      time.Time `json:"created_at"`
}

// CreateProject creates a new project
func (c *Client) CreateProject(ctx context.Context, project *Project) error {
	resp, err := c.Resty.R().
		SetContext(ctx).
		SetBody(project).
		SetResult(project).
		Post("/api/v1/projects")

	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("failed to create project: %s", resp.Status())
	}

	return nil
}

// ReadProject retrieves a project by ID
func (c *Client) ReadProject(ctx context.Context, id string) (*Project, error) {
	var project Project
	resp, err := c.Resty.R().
		SetContext(ctx).
		SetResult(&project).
		Get(fmt.Sprintf("/api/v1/projects/%s", id))

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() == http.StatusNotFound {
		return nil, nil
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to read project: %s", resp.Status())
	}

	return &project, nil
}

// UpdateProject updates an existing project
func (c *Client) UpdateProject(ctx context.Context, project *Project) error {
	resp, err := c.Resty.R().
		SetContext(ctx).
		SetBody(project).
		SetResult(project).
		Put(fmt.Sprintf("/api/v1/projects/%s", project.ID))

	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("failed to update project: %s", resp.Status())
	}

	return nil
}

// DeleteProject deletes a project
func (c *Client) DeleteProject(ctx context.Context, id string) error {
	resp, err := c.Resty.R().
		SetContext(ctx).
		Delete(fmt.Sprintf("/api/v1/projects/%s", id))

	if err != nil {
		return err
	}

	if resp.IsError() && resp.StatusCode() != http.StatusNotFound {
		return fmt.Errorf("failed to delete project: %s", resp.Status())
	}

	return nil
}

// Secret represents a MazeVault secret
type Secret struct {
	ID          string            `json:"id"`
	ProjectID   string            `json:"project_id"`
	Key         string            `json:"key"`
	Value       string            `json:"value,omitempty"`
	Environment string            `json:"environment"`
	TTLHours    int               `json:"ttl_hours,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Version     int               `json:"version"`
	Status      string            `json:"status"`
	CreatedAt   time.Time         `json:"created_at"`
}

// CreateSecret creates a new secret
func (c *Client) CreateSecret(ctx context.Context, secret *Secret) error {
	resp, err := c.Resty.R().
		SetContext(ctx).
		SetBody(secret).
		SetResult(secret).
		Post("/api/v1/secrets")

	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("failed to create secret: %s", resp.Status())
	}

	return nil
}

// ReadSecret retrieves a secret by ID
func (c *Client) ReadSecret(ctx context.Context, id string) (*Secret, error) {
	var secret Secret
	resp, err := c.Resty.R().
		SetContext(ctx).
		SetResult(&secret).
		Get(fmt.Sprintf("/api/v1/secrets/%s", id))

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() == http.StatusNotFound {
		return nil, nil
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to read secret: %s", resp.Status())
	}

	return &secret, nil
}

// UpdateSecret updates an existing secret
func (c *Client) UpdateSecret(ctx context.Context, secret *Secret) error {
	resp, err := c.Resty.R().
		SetContext(ctx).
		SetBody(secret).
		SetResult(secret).
		Put(fmt.Sprintf("/api/v1/secrets/%s", secret.ID))

	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("failed to update secret: %s", resp.Status())
	}

	return nil
}

// DeleteSecret deletes a secret
func (c *Client) DeleteSecret(ctx context.Context, id string) error {
	resp, err := c.Resty.R().
		SetContext(ctx).
		Delete(fmt.Sprintf("/api/v1/secrets/%s", id))

	if err != nil {
		return err
	}

	if resp.IsError() && resp.StatusCode() != http.StatusNotFound {
		return fmt.Errorf("failed to delete secret: %s", resp.Status())
	}

	return nil
}

// APIToken represents an API token
type APIToken struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Token     string    `json:"token,omitempty"`
	ExpiresAt time.Time `json:"expires_at"`
	Scopes    []string  `json:"scopes"`
}

// Certificate represents a certificate
type Certificate struct {
	ID                      string    `json:"id"`
	CommonName              string    `json:"common_name"`
	SerialNumber            string    `json:"serial_number"`
	CertificatePEM          string    `json:"certificate_pem"`
	PrivateKeyPEM           string    `json:"private_key_pem,omitempty"`
	Status                  string    `json:"status"`
	ExpiresAt               time.Time `json:"expiry_date"`
	OrganizationCAAccountID string    `json:"organization_ca_account_id,omitempty"`
}

// CreateCertificateRequest represents request to create certificate
type CreateCertificateRequest struct {
	CommonName string `json:"common_name"`
	TTL        string `json:"ttl"`
	KeySize    int    `json:"key_size"`
	TemplateID string `json:"template_id,omitempty"`
	CSRPEM     string `json:"csr_pem,omitempty"`
}

// CreateCertificate creates a new certificate via the CSR endpoint.
// The caller is responsible for generating the CSR (e.g., via the tls_private_key Terraform resource).
// TemplateID is required; CSRPEM is the PEM-encoded certificate signing request.
func (c *Client) CreateCertificate(ctx context.Context, req *CreateCertificateRequest) (*Certificate, error) {
	type submitCSRBody struct {
		TemplateID    string `json:"template_id,omitempty"`
		CSRPEM        string `json:"csr_pem,omitempty"`
		RequestedCN   string `json:"requested_cn"`
		CSRSourceMode string `json:"csr_source_mode,omitempty"`
	}
	body := submitCSRBody{
		TemplateID:    req.TemplateID,
		CSRPEM:        req.CSRPEM,
		RequestedCN:   req.CommonName,
		CSRSourceMode: "uploaded",
	}
	var cert Certificate
	resp, err := c.Resty.R().
		SetContext(ctx).
		SetBody(body).
		SetResult(&cert).
		Post("/api/v1/certificates/csr")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to create certificate: %s", resp.Status())
	}
	return &cert, nil
}

// ReadCertificate retrieves a certificate
func (c *Client) ReadCertificate(ctx context.Context, id string) (*Certificate, error) {
	var cert Certificate
	resp, err := c.Resty.R().
		SetContext(ctx).
		SetResult(&cert).
		Get(fmt.Sprintf("/api/v1/certificates/%s", id))

	if err != nil {
		return nil, err
	}
	if resp.StatusCode() == http.StatusNotFound {
		return nil, nil
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to read certificate: %s", resp.Status())
	}
	return &cert, nil
}

// RevokeCertificate revokes a certificate
func (c *Client) RevokeCertificate(ctx context.Context, id string) error {
	resp, err := c.Resty.R().
		SetContext(ctx).
		Post(fmt.Sprintf("/api/v1/certificates/%s/revoke", id))

	if err != nil {
		return err
	}
	if resp.IsError() && resp.StatusCode() != http.StatusNotFound {
		return fmt.Errorf("failed to revoke certificate: %s", resp.Status())
	}
	return nil
}

// ListProjectCertificates lists certificates for a project
func (c *Client) ListProjectCertificates(ctx context.Context, projectID string) ([]Certificate, error) {
	var certs []Certificate
	resp, err := c.Resty.R().
		SetContext(ctx).
		SetResult(&certs).
		Get(fmt.Sprintf("/api/v1/projects/%s/certificates", projectID))

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to list project certificates: %s", resp.Status())
	}
	return certs, nil
}

// CertificateRequest represents a certificate signing request
type CertificateRequest struct {
	ID              string                 `json:"id"`
	ProjectID       *string                `json:"project_id,omitempty"`
	TemplateID      *string                `json:"template_id,omitempty"`
	CSRPEM          string                 `json:"csr_pem"`
	RequestedCN     string                 `json:"requested_cn"`
	RequestedSAN    map[string]interface{} `json:"requested_san"`
	Status          string                 `json:"status"`
	RejectionReason string                 `json:"rejection_reason,omitempty"`
	IssuedCertID    *string                `json:"issued_cert_id,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
}

// ListProjectCSRs lists CSRs for a project
func (c *Client) ListProjectCSRs(ctx context.Context, projectID string) ([]CertificateRequest, error) {
	var csrs []CertificateRequest
	resp, err := c.Resty.R().
		SetContext(ctx).
		SetResult(&csrs).
		Get(fmt.Sprintf("/api/v1/projects/%s/certificate-requests", projectID))

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to list project CSRs: %s", resp.Status())
	}
	return csrs, nil
}

// CA represents a Certificate Authority
type CA struct {
	ID         string  `json:"id"`
	ProjectID  *string `json:"project_id,omitempty"`
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	Status     string  `json:"status"`
	ValidUntil string  `json:"valid_until"` // Using string for simplicity in parsing
}

// CreateCARequest represents request to initialize CA
type CreateCARequest struct {
	Name       string `json:"name"`
	ValidYears int    `json:"valid_years"`
	KeySize    int    `json:"key_size"`
}

// CreateCA initializes a new Root CA
func (c *Client) CreateCA(ctx context.Context, req *CreateCARequest) (*CA, error) {
	// The backend endpoint is /ca/initialize which initializes the SINGLE root CA or a new one?
	// The gap analysis says "Initialize Root CA".
	// If it's a singleton or global operation, it might be tricky for Terraform resource which usually implies multiple instances.
	// But let's assume it creates a CA record.
	var ca CA
	resp, err := c.Resty.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&ca).
		Post("/api/v1/ca/initialize")

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to initialize CA: %s", resp.Status())
	}
	return &ca, nil
}

// CreateProjectCA initializes a new Root CA for a project
func (c *Client) CreateProjectCA(ctx context.Context, projectID string, req *CreateCARequest) (*CA, error) {
	var ca CA
	resp, err := c.Resty.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&ca).
		Post(fmt.Sprintf("/api/v1/projects/%s/ca/initialize", projectID))

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to initialize project CA: %s", resp.Status())
	}
	return &ca, nil
}

// ListProjectCAs lists CAs for a project
func (c *Client) ListProjectCAs(ctx context.Context, projectID string) ([]CA, error) {
	var cas []CA
	resp, err := c.Resty.R().
		SetContext(ctx).
		SetResult(&cas).
		Get(fmt.Sprintf("/api/v1/projects/%s/cas", projectID))

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to list project CAs: %s", resp.Status())
	}
	return cas, nil
}

// GetProjectCA retrieves a specific CA for a project
func (c *Client) GetProjectCA(ctx context.Context, projectID, caID string) (*CA, error) {
	var ca CA
	resp, err := c.Resty.R().
		SetContext(ctx).
		SetResult(&ca).
		Get(fmt.Sprintf("/api/v1/projects/%s/cas/%s", projectID, caID))

	if err != nil {
		return nil, err
	}
	if resp.StatusCode() == http.StatusNotFound {
		return nil, nil
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to get project CA: %s", resp.Status())
	}
	return &ca, nil
}

// ReadCA retrieves a CA
func (c *Client) ReadCA(ctx context.Context, id string) (*CA, error) {
	var ca CA
	resp, err := c.Resty.R().
		SetContext(ctx).
		SetResult(&ca).
		Get(fmt.Sprintf("/api/v1/ca/%s", id))

	if err != nil {
		return nil, err
	}
	if resp.StatusCode() == http.StatusNotFound {
		return nil, nil
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to read CA: %s", resp.Status())
	}
	return &ca, nil
}

// DeleteCA deletes a CA (Not supported by backend usually, but needed for TF)
func (c *Client) DeleteCA(ctx context.Context, id string) error {
	// Placeholder: Backend doesn't seem to have Delete CA.
	// We just return nil to allow Terraform to remove it from state.
	return nil
}

// CertificateTemplate represents a certificate template
type CertificateTemplate struct {
	ID             string   `json:"id"`
	ProjectID      *string  `json:"project_id,omitempty"`
	Name           string   `json:"name"`
	Type           string   `json:"type"`
	ValidityPeriod string   `json:"validity_period"`
	KeyUsage       []string `json:"key_usage"`
}

// CreateTemplateRequest represents request to create template
type CreateTemplateRequest struct {
	Name           string   `json:"name"`
	Type           string   `json:"type"`
	ValidityPeriod string   `json:"validity_period"`
	KeyUsage       []string `json:"key_usage"`
}

// CreateTemplate creates a new certificate template
func (c *Client) CreateTemplate(ctx context.Context, req *CreateTemplateRequest) (*CertificateTemplate, error) {
	var tmpl CertificateTemplate
	resp, err := c.Resty.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&tmpl).
		Post("/api/v1/certificate-templates")

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to create template: %s", resp.Status())
	}
	return &tmpl, nil
}

// CreateProjectTemplate creates a new certificate template for a project
func (c *Client) CreateProjectTemplate(ctx context.Context, projectID string, req *CreateTemplateRequest) (*CertificateTemplate, error) {
	var tmpl CertificateTemplate
	resp, err := c.Resty.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&tmpl).
		Post(fmt.Sprintf("/api/v1/projects/%s/certificate-templates", projectID))

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to create project template: %s", resp.Status())
	}
	return &tmpl, nil
}

// ListProjectCertificateTemplates lists certificate templates for a project
func (c *Client) ListProjectCertificateTemplates(ctx context.Context, projectID string) ([]CertificateTemplate, error) {
	var tmpls []CertificateTemplate
	resp, err := c.Resty.R().
		SetContext(ctx).
		SetResult(&tmpls).
		Get(fmt.Sprintf("/api/v1/projects/%s/certificate-templates", projectID))

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to list project templates: %s", resp.Status())
	}
	return tmpls, nil
}

// GetProjectCertificateTemplate retrieves a specific template for a project
func (c *Client) GetProjectCertificateTemplate(ctx context.Context, projectID, templateID string) (*CertificateTemplate, error) {
	var tmpl CertificateTemplate
	resp, err := c.Resty.R().
		SetContext(ctx).
		SetResult(&tmpl).
		Get(fmt.Sprintf("/api/v1/projects/%s/certificate-templates/%s", projectID, templateID))

	if err != nil {
		return nil, err
	}
	if resp.StatusCode() == http.StatusNotFound {
		return nil, nil
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to get project template: %s", resp.Status())
	}
	return &tmpl, nil
}

// ReadTemplate retrieves a template
func (c *Client) ReadTemplate(ctx context.Context, id string) (*CertificateTemplate, error) {
	var tmpl CertificateTemplate
	resp, err := c.Resty.R().
		SetContext(ctx).
		SetResult(&tmpl).
		Get(fmt.Sprintf("/api/v1/certificate-templates/%s", id))

	if err != nil {
		return nil, err
	}
	if resp.StatusCode() == http.StatusNotFound {
		return nil, nil
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to read template: %s", resp.Status())
	}
	return &tmpl, nil
}

// DeleteTemplate deletes a template (Not supported by backend usually)
func (c *Client) DeleteTemplate(ctx context.Context, id string) error {
	return nil
}

// Role represents an RBAC role
type Role struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

// CreateRoleRequest represents request to create role
type CreateRoleRequest struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

// CreateRole creates a new role
func (c *Client) CreateRole(ctx context.Context, req *CreateRoleRequest) (*Role, error) {
	var role Role
	resp, err := c.Resty.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&role).
		Post("/api/v1/rbac/roles")

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to create role: %s", resp.Status())
	}
	return &role, nil
}

// ReadRole retrieves a role by ID (by listing all and filtering)
func (c *Client) ReadRole(ctx context.Context, id string) (*Role, error) {
	var roles []Role
	resp, err := c.Resty.R().
		SetContext(ctx).
		SetResult(&roles).
		Get("/api/v1/rbac/roles")

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to list roles: %s", resp.Status())
	}

	for _, r := range roles {
		if r.ID == id {
			return &r, nil
		}
	}
	return nil, nil // Not found
}

// GroupMapping represents a group mapping
type GroupMapping struct {
	ID        string `json:"id"`
	GroupName string `json:"group_name"`
	RoleID    string `json:"role_id"`
}

// CreateGroupMappingRequest represents request to create group mapping
type CreateGroupMappingRequest struct {
	GroupName string `json:"group_name"`
	RoleID    string `json:"role_id"`
}

// CreateGroupMapping creates a new group mapping
func (c *Client) CreateGroupMapping(ctx context.Context, req *CreateGroupMappingRequest) (*GroupMapping, error) {
	var mapping GroupMapping
	resp, err := c.Resty.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&mapping).
		Post("/api/v1/rbac/mappings")

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to create group mapping: %s", resp.Status())
	}
	return &mapping, nil
}

// ReadGroupMapping retrieves a group mapping by ID
func (c *Client) ReadGroupMapping(ctx context.Context, id string) (*GroupMapping, error) {
	var mappings []GroupMapping
	resp, err := c.Resty.R().
		SetContext(ctx).
		SetResult(&mappings).
		Get("/api/v1/rbac/mappings")

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to list group mappings: %s", resp.Status())
	}

	for _, m := range mappings {
		if m.ID == id {
			return &m, nil
		}
	}
	return nil, nil
}

// DeleteGroupMapping deletes a group mapping
func (c *Client) DeleteGroupMapping(ctx context.Context, id string) error {
	resp, err := c.Resty.R().
		SetContext(ctx).
		Delete(fmt.Sprintf("/api/v1/rbac/mappings/%s", id))

	if err != nil {
		return err
	}
	if resp.IsError() && resp.StatusCode() != http.StatusNotFound {
		return fmt.Errorf("failed to delete group mapping: %s", resp.Status())
	}
	return nil
}

// CreateAPIToken creates a new API token
func (c *Client) CreateAPIToken(ctx context.Context, name string, scopes []string, duration string) (*APIToken, error) {
	req := map[string]interface{}{
		"name":     name,
		"scopes":   scopes,
		"duration": duration,
	}

	var token APIToken
	resp, err := c.Resty.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&token).
		Post("/api/v1/users/tokens")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to create api token: %s", resp.Status())
	}

	return &token, nil
}

// RevokeAPIToken revokes an API token
func (c *Client) RevokeAPIToken(ctx context.Context, id string) error {
	resp, err := c.Resty.R().
		SetContext(ctx).
		Delete(fmt.Sprintf("/api/v1/users/tokens/%s", id))

	if err != nil {
		return err
	}

	if resp.IsError() && resp.StatusCode() != http.StatusNotFound {
		return fmt.Errorf("failed to revoke api token: %s", resp.Status())
	}

	return nil
}
