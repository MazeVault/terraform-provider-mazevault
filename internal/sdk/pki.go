package mazevault

import (
	"fmt"
	"net/http"
)

// ========================
// Project CAs (internal PKI)
// ========================

// ProjectCA represents a project-level internal Certificate Authority
type ProjectCA struct {
	ID         string  `json:"id"`
	ProjectID  *string `json:"project_id,omitempty"`
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	Status     string  `json:"status"`
	ValidUntil string  `json:"valid_until"`
}

// InitializeProjectCARequest is the request body for initializing a project CA
type InitializeProjectCARequest struct {
	Name       string `json:"name"`
	ValidYears int    `json:"valid_years"`
	KeySize    int    `json:"key_size"`
}

// InitializeProjectCA initializes an internal root CA for a project
func (c *Client) InitializeProjectCA(projectID string, req *InitializeProjectCARequest) (*ProjectCA, error) {
	r, err := c.newRequest(http.MethodPost, "/api/v1/projects/"+projectID+"/ca/initialize", req)
	if err != nil {
		return nil, err
	}
	var resp ProjectCA
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetProjectCA retrieves a specific CA within a project
func (c *Client) GetProjectCA(projectID, caID string) (*ProjectCA, error) {
	r, err := c.newRequest(http.MethodGet, fmt.Sprintf("/api/v1/projects/%s/cas/%s", projectID, caID), nil)
	if err != nil {
		return nil, err
	}
	var resp ProjectCA
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListProjectCAs lists all CAs for a project
func (c *Client) ListProjectCAs(projectID string) ([]ProjectCA, error) {
	r, err := c.newRequest(http.MethodGet, "/api/v1/projects/"+projectID+"/cas", nil)
	if err != nil {
		return nil, err
	}
	var resp []ProjectCA
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// ========================
// Certificate Templates
// ========================

// CertificateTemplate represents a certificate template
type CertificateTemplate struct {
	ID             string   `json:"id"`
	ProjectID      *string  `json:"project_id,omitempty"`
	Name           string   `json:"name"`
	Type           string   `json:"type"`
	ValidityPeriod string   `json:"validity_period"`
	KeyUsage       []string `json:"key_usage"`
}

// CreateCertificateTemplateRequest is the request body for creating a template
type CreateCertificateTemplateRequest struct {
	Name           string   `json:"name"`
	Type           string   `json:"type"`
	ValidityPeriod string   `json:"validity_period"`
	KeyUsage       []string `json:"key_usage,omitempty"`
}

// CreateCertificateTemplate creates a certificate template
func (c *Client) CreateCertificateTemplate(req *CreateCertificateTemplateRequest) (*CertificateTemplate, error) {
	r, err := c.newRequest(http.MethodPost, "/api/v1/certificate-templates", req)
	if err != nil {
		return nil, err
	}
	var resp CertificateTemplate
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateProjectCertificateTemplate creates a certificate template scoped to a project
func (c *Client) CreateProjectCertificateTemplate(projectID string, req *CreateCertificateTemplateRequest) (*CertificateTemplate, error) {
	r, err := c.newRequest(http.MethodPost, "/api/v1/projects/"+projectID+"/certificate-templates", req)
	if err != nil {
		return nil, err
	}
	var resp CertificateTemplate
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCertificateTemplate retrieves a certificate template by ID
func (c *Client) GetCertificateTemplate(id string) (*CertificateTemplate, error) {
	r, err := c.newRequest(http.MethodGet, "/api/v1/certificate-templates/"+id, nil)
	if err != nil {
		return nil, err
	}
	var resp CertificateTemplate
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateCertificateTemplate updates a certificate template
func (c *Client) UpdateCertificateTemplate(id string, req *CreateCertificateTemplateRequest) (*CertificateTemplate, error) {
	r, err := c.newRequest(http.MethodPut, "/api/v1/certificate-templates/"+id, req)
	if err != nil {
		return nil, err
	}
	var resp CertificateTemplate
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteCertificateTemplate deletes a certificate template
func (c *Client) DeleteCertificateTemplate(id string) error {
	r, err := c.newRequest(http.MethodDelete, "/api/v1/certificate-templates/"+id, nil)
	if err != nil {
		return err
	}
	return c.Do(r, nil)
}

// ListProjectCertificateTemplates lists templates for a project
func (c *Client) ListProjectCertificateTemplates(projectID string) ([]CertificateTemplate, error) {
	r, err := c.newRequest(http.MethodGet, "/api/v1/projects/"+projectID+"/certificate-templates", nil)
	if err != nil {
		return nil, err
	}
	var resp []CertificateTemplate
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// ========================
// Certificates
// ========================

// Certificate represents an issued certificate
type Certificate struct {
	ID             string  `json:"id"`
	CommonName     string  `json:"common_name"`
	SerialNumber   string  `json:"serial_number"`
	CertificatePEM string  `json:"certificate_pem"`
	PrivateKeyPEM  string  `json:"private_key_pem,omitempty"`
	Status         string  `json:"status"`
	ExpiresAt      string  `json:"expiry_date"`
	ProjectID      *string `json:"project_id,omitempty"`
	TemplateID     *string `json:"template_id,omitempty"`
}

// SubmitCSRRequest is the request body for submitting a CSR
type SubmitCSRRequest struct {
	TemplateID    string `json:"template_id,omitempty"`
	CSRPEM        string `json:"csr_pem,omitempty"`
	RequestedCN   string `json:"requested_cn"`
	CSRSourceMode string `json:"csr_source_mode,omitempty"`
}

// SubmitCSR submits a CSR and returns the issued certificate
func (c *Client) SubmitCSR(req *SubmitCSRRequest) (*Certificate, error) {
	r, err := c.newRequest(http.MethodPost, "/api/v1/certificates/csr", req)
	if err != nil {
		return nil, err
	}
	var resp Certificate
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCertificate retrieves a certificate by ID
func (c *Client) GetCertificate(id string) (*Certificate, error) {
	r, err := c.newRequest(http.MethodGet, "/api/v1/certificates/"+id, nil)
	if err != nil {
		return nil, err
	}
	var resp Certificate
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// RevokeCertificate revokes a certificate by ID
func (c *Client) RevokeCertificate(id string) error {
	r, err := c.newRequest(http.MethodPost, "/api/v1/certificates/"+id+"/revoke", nil)
	if err != nil {
		return err
	}
	return c.Do(r, nil)
}

// ListProjectCertificates lists certificates for a project
func (c *Client) ListProjectCertificates(projectID string) ([]Certificate, error) {
	r, err := c.newRequest(http.MethodGet, "/api/v1/projects/"+projectID+"/certificates", nil)
	if err != nil {
		return nil, err
	}
	var resp []Certificate
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// ========================
// RBAC: Roles & Group Mappings
// ========================

// GroupMapping maps an identity provider group to a MazeVault role
type GroupMapping struct {
	ID        string `json:"id"`
	GroupName string `json:"group_name"`
	RoleID    string `json:"role_id"`
}

// CreateGroupMappingRequest is the request body for creating a group mapping
type CreateGroupMappingRequest struct {
	GroupName string `json:"group_name"`
	RoleID    string `json:"role_id"`
}

// CreateGroupMapping creates a group → role mapping
func (c *Client) CreateGroupMapping(req *CreateGroupMappingRequest) (*GroupMapping, error) {
	r, err := c.newRequest(http.MethodPost, "/api/v1/rbac/mappings", req)
	if err != nil {
		return nil, err
	}
	var resp GroupMapping
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListGroupMappings lists all group → role mappings
func (c *Client) ListGroupMappings() ([]GroupMapping, error) {
	r, err := c.newRequest(http.MethodGet, "/api/v1/rbac/mappings", nil)
	if err != nil {
		return nil, err
	}
	var resp []GroupMapping
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetGroupMapping retrieves a group mapping by searching the list
func (c *Client) GetGroupMapping(id string) (*GroupMapping, error) {
	all, err := c.ListGroupMappings()
	if err != nil {
		return nil, err
	}
	for i, m := range all {
		if m.ID == id {
			return &all[i], nil
		}
	}
	return nil, nil
}

// DeleteGroupMapping deletes a group mapping
func (c *Client) DeleteGroupMapping(id string) error {
	r, err := c.newRequest(http.MethodDelete, "/api/v1/rbac/mappings/"+id, nil)
	if err != nil {
		return err
	}
	return c.Do(r, nil)
}

// GetRole retrieves a role by scanning the list
func (c *Client) GetRole(id string) (*Role, error) {
	all, err := c.ListRoles()
	if err != nil {
		return nil, err
	}
	for i, ro := range all {
		if ro.ID == id {
			return &all[i], nil
		}
	}
	return nil, nil
}

// ========================
// Certificate Signing Requests (project-level)
// ========================

// CertificateRequest represents a CSR submitted to MazeVault
type CertificateRequest struct {
	ID              string  `json:"id"`
	ProjectID       *string `json:"project_id,omitempty"`
	TemplateID      *string `json:"template_id,omitempty"`
	CSRPEM          string  `json:"csr_pem"`
	RequestedCN     string  `json:"requested_cn"`
	Status          string  `json:"status"`
	RejectionReason string  `json:"rejection_reason,omitempty"`
	IssuedCertID    *string `json:"issued_cert_id,omitempty"`
	CreatedAt       string  `json:"created_at"`
}

// ListProjectCSRs lists certificate signing requests for a project
func (c *Client) ListProjectCSRs(projectID string) ([]CertificateRequest, error) {
	r, err := c.newRequest(http.MethodGet, "/api/v1/projects/"+projectID+"/certificate-requests", nil)
	if err != nil {
		return nil, err
	}
	var resp []CertificateRequest
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}
