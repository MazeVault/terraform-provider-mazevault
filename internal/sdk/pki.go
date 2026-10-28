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
	OCSPURL    string  `json:"ocsp_url"`
	CRLURL     string  `json:"crl_url"`
}

// InitializeProjectCARequest is the request body for initializing a project CA
type InitializeProjectCARequest struct {
	Name       string `json:"name"`
	ValidYears int    `json:"valid_years"`
	KeySize    int    `json:"key_size"`
}

// UpdateProjectCARequest carries the mutable configuration fields for an existing project CA.
// All fields are optional — only non-nil values are applied (patch semantics).
type UpdateProjectCARequest struct {
	OCSPURL             *string `json:"ocsp_url,omitempty"`
	CRLURL              *string `json:"crl_url,omitempty"`
	OCSPResponderCertID *string `json:"ocsp_responder_cert_id,omitempty"`
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
	r, err := c.newRequest(http.MethodGet, fmt.Sprintf("/api/v1/projects/%s/ca/%s", projectID, caID), nil)
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
	r, err := c.newRequest(http.MethodGet, "/api/v1/projects/"+projectID+"/ca", nil)
	if err != nil {
		return nil, err
	}
	var resp []ProjectCA
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// UpdateProjectCA updates the mutable OCSP/CRL configuration of a project CA.
// Only non-nil fields in req are applied (patch semantics).
func (c *Client) UpdateProjectCA(projectID, caID string, req *UpdateProjectCARequest) (*ProjectCA, error) {
	r, err := c.newRequest(http.MethodPut, fmt.Sprintf("/api/v1/projects/%s/ca/%s", projectID, caID), req)
	if err != nil {
		return nil, err
	}
	var resp ProjectCA
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ========================
// Organization Internal CA
// ========================

// InternalCA represents an organization's single internal Certificate Authority.
// It mirrors the backend models.CertificateAuthority response.
type InternalCA struct {
	ID           string `json:"id"`
	AccountID    string `json:"account_id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Status       string `json:"status"`
	SubjectCN    string `json:"subject_cn"`
	SerialNumber string `json:"serial_number"`
	NotBefore    string `json:"not_before"`
	NotAfter     string `json:"not_after"`
	KeyAlgorithm string `json:"key_algorithm"`
	KeySize      int    `json:"key_size"`
	IsRoot       bool   `json:"is_root"`
	CRLURL       string `json:"crl_url"`
	OCSPURL      string `json:"ocsp_url"`
}

// InitializeInternalCARequest is the payload for creating an organization's
// single internal Certificate Authority.
type InitializeInternalCARequest struct {
	CommonName    string `json:"common_name,omitempty"`
	KeySize       int    `json:"key_size,omitempty"`
	ValidityYears int    `json:"validity_years,omitempty"`
	KeyAlgorithm  string `json:"key_algorithm,omitempty"`
}

// InitializeInternalCA creates the single internal Root CA for an organization.
// It fails with a conflict error if one already exists.
func (c *Client) InitializeInternalCA(orgID string, req *InitializeInternalCARequest) (*InternalCA, error) {
	r, err := c.newRequest(http.MethodPost, "/api/v1/organizations/"+orgID+"/internal-ca/initialize", req)
	if err != nil {
		return nil, err
	}
	var resp InternalCA
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetInternalCA returns the organization's internal Certificate Authority, or
// (nil, nil) when none is initialized (the backend responds 204 No Content).
func (c *Client) GetInternalCA(orgID string) (*InternalCA, error) {
	r, err := c.newRequest(http.MethodGet, "/api/v1/organizations/"+orgID+"/internal-ca", nil)
	if err != nil {
		return nil, err
	}
	var resp InternalCA
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	if resp.ID == "" {
		return nil, nil
	}
	return &resp, nil
}

// DeleteInternalCA soft-deletes the organization's internal Certificate
// Authority. The backend refuses (409) when active certificates were issued
// by it.
func (c *Client) DeleteInternalCA(orgID string) error {
	r, err := c.newRequest(http.MethodDelete, "/api/v1/organizations/"+orgID+"/internal-ca", nil)
	if err != nil {
		return err
	}
	return c.Do(r, nil)
}

// ========================
// Certificate Templates
// ========================

// CertificateTemplate represents a certificate template
type CertificateTemplate struct {
	ID             string   `json:"id"`
	OrganizationID string   `json:"organization_id,omitempty"`
	ProjectID      *string  `json:"project_id,omitempty"`
	Name           string   `json:"name"`
	Type           string   `json:"type"`
	ValidityPeriod string   `json:"validity_period"`
	KeyUsage       []string `json:"key_usage"`
	ChallengeType  string   `json:"challenge_type,omitempty"`
}

// CreateCertificateTemplateRequest is the request body for creating a template
type CreateCertificateTemplateRequest struct {
	Name           string   `json:"name"`
	Type           string   `json:"type"`
	ValidityPeriod string   `json:"validity_period"`
	KeyUsage       []string `json:"key_usage,omitempty"`
	ChallengeType  string   `json:"challenge_type,omitempty"`
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

// CreateOrgCertificateTemplate creates an organization-scoped certificate template.
// Certificate templates are org-scoped entities; they are then assigned to
// certificate-management projects via AssignCertificateTemplateToProject.
func (c *Client) CreateOrgCertificateTemplate(organizationID string, req *CreateCertificateTemplateRequest) (*CertificateTemplate, error) {
	r, err := c.newRequest(http.MethodPost, "/api/v1/certificate-templates/?organization_id="+organizationID, req)
	if err != nil {
		return nil, err
	}
	var resp CertificateTemplate
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AssignCertificateTemplateToProject assigns an org template to a project.
func (c *Client) AssignCertificateTemplateToProject(projectID, templateID string) error {
	r, err := c.newRequest(http.MethodPost, "/api/v1/projects/"+projectID+"/templates/"+templateID, map[string]interface{}{})
	if err != nil {
		return err
	}
	return c.Do(r, nil)
}

// UnassignCertificateTemplateFromProject removes a template assignment from a project.
func (c *Client) UnassignCertificateTemplateFromProject(projectID, templateID string) error {
	r, err := c.newRequest(http.MethodDelete, "/api/v1/projects/"+projectID+"/templates/"+templateID, nil)
	if err != nil {
		return err
	}
	return c.Do(r, nil)
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
	ID           string `json:"id"`
	CommonName   string `json:"common_name"`
	SerialNumber string `json:"serial_number"`
	// CertificatePEM is the full PEM bundle (leaf followed by the effective issuer chain).
	// The API field is public_chain_pem; it always reflects the same chain export produces.
	CertificatePEM string `json:"public_chain_pem"`
	// ChainPEM holds the issuer chain only (nearest issuer first, root last).
	ChainPEM string `json:"chain_pem,omitempty"`
	// ChainSource reports where the effective chain came from: "stored", "inventory" or "none".
	ChainSource string `json:"chain_source,omitempty"`
	// ChainComplete is true when the effective chain terminates at a self-signed root.
	ChainComplete           bool    `json:"chain_complete,omitempty"`
	PrivateKeyPEM           string  `json:"private_key_pem,omitempty"`
	Status                  string  `json:"status"`
	ExpiresAt               string  `json:"expiry_date"`
	ProjectID               *string `json:"project_id,omitempty"`
	TemplateID              *string `json:"template_id,omitempty"`
	OrganizationCAAccountID *string `json:"organization_ca_account_id,omitempty"`
}

// SubmitCSRRequest is the request body for submitting a CSR
type SubmitCSRRequest struct {
	TemplateID              string `json:"template_id,omitempty"`
	ProjectID               string `json:"project_id,omitempty"`
	CSRPEM                  string `json:"csr_pem,omitempty"`
	RequestedCN             string `json:"requested_cn"`
	CSRSourceMode           string `json:"csr_source_mode,omitempty"`
	RequestedKeySize        int    `json:"requested_key_size,omitempty"`
	OrganizationCAAccountID string `json:"organization_ca_account_id,omitempty"`

	// Per-request issuance overrides for commercial CAs (request -> template -> CA
	// account precedence). All optional; omitted values inherit the template, then the
	// account defaults.
	DCVMethod             string            `json:"dcv_method,omitempty"`
	CAProductID           string            `json:"ca_product_id,omitempty"`
	OrderYears            *int              `json:"order_years,omitempty"`
	RequestedValidityDays *int              `json:"requested_validity_days,omitempty"`
	TechContact           map[string]string `json:"tech_contact,omitempty"`
	Voucher               string            `json:"voucher,omitempty"`
}

// csrSubmitResponse mirrors the CertificateRequest returned by POST
// /certificates/csr. When the request is auto-approved the certificate is
// issued synchronously and IssuedCertID points at the resulting certificate.
type csrSubmitResponse struct {
	ID           string  `json:"id"`
	Status       string  `json:"status"`
	IssuedCertID *string `json:"issued_cert_id"`
	ProjectID    *string `json:"project_id"`
	TemplateID   *string `json:"template_id"`
}

// SubmitCSR submits a CSR and returns the issued certificate. The backend
// responds with a certificate request; when it is auto-approved and issued,
// this method resolves and returns the actual issued certificate.
func (c *Client) SubmitCSR(req *SubmitCSRRequest) (*Certificate, error) {
	r, err := c.newRequest(http.MethodPost, "/api/v1/certificates/csr", req)
	if err != nil {
		return nil, err
	}
	var reqResp csrSubmitResponse
	if err := c.Do(r, &reqResp); err != nil {
		return nil, err
	}
	if reqResp.IssuedCertID != nil && *reqResp.IssuedCertID != "" {
		return c.GetCertificate(*reqResp.IssuedCertID)
	}
	// Not yet issued (e.g. pending approval) — surface the request state so the
	// caller does not mistake the request ID for a certificate ID.
	return nil, fmt.Errorf("certificate request %s is not issued (status: %s); approval may be required", reqResp.ID, reqResp.Status)
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
	// The revoke endpoint requires a body with an RFC 5280 reason code.
	// 4 = superseded, an appropriate neutral reason for lifecycle-driven revocation.
	body := map[string]interface{}{"reason_code": 4, "comment": "revoked via SDK"}
	r, err := c.newRequest(http.MethodPost, "/api/v1/certificates/"+id+"/revoke", body)
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
