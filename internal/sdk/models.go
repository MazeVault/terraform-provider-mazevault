package mazevault

import "time"

// LoginRequest represents the login credentials
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	AccessToken string       `json:"access_token"`
	User        UserResponse `json:"user"`
}

// UserResponse represents the user information
type UserResponse struct {
	Email            string `json:"email"`
	FullName         string `json:"full_name"`
	MustChangePasswd bool   `json:"must_change_passwd"`
	Role             string `json:"role"`
}

// Project represents a project
type Project struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	Type               string    `json:"type"`
	Environment        string    `json:"environment"`
	DefaultEnvironment string    `json:"default_environment"`
	OrganizationID     string    `json:"organization_id"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// CreateProjectRequest represents the request to create a project
type CreateProjectRequest struct {
	Name           string `json:"name"`
	Type           string `json:"type,omitempty"`
	Environment    string `json:"environment,omitempty"`
	OrganizationID string `json:"organization_id,omitempty"`
}

// Organization represents an organization
type Organization struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateOrganizationRequest represents the request to create an organization
type CreateOrganizationRequest struct {
	Name string `json:"name"`
}

// Secret represents a secret (decrypted)
type Secret struct {
	ID          string            `json:"id"`
	ProjectID   string            `json:"project_id"`
	Key         string            `json:"key"`
	Value       string            `json:"value"`
	Environment string            `json:"environment"`
	TTLHours    int               `json:"ttl_hours"`
	Metadata    map[string]string `json:"metadata"`
	Version     int               `json:"version"`
	Status      string            `json:"status"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// RotationConfig represents rotation settings
type RotationConfig struct {
	Enabled              bool     `json:"enabled"`
	Schedule             string   `json:"schedule"`
	RotationIntervalDays int      `json:"rotation_interval_days"`
	NotificationEmails   []string `json:"notification_emails"`
}

// CreateSecretRequest represents the request to create a secret
type CreateSecretRequest struct {
	Key         string            `json:"key"`
	Value       string            `json:"value"`
	ProjectID   string            `json:"project_id"`
	Environment string            `json:"environment,omitempty"`
	TTLHours    int               `json:"ttl_hours,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Rotation    *RotationConfig   `json:"rotation,omitempty"`
}

// CreateSecretResponse represents the response after creating a secret
type CreateSecretResponse struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Version int    `json:"version"`
}

// APIToken represents an API token
type APIToken struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Token     string    `json:"token,omitempty"` // Only returned on creation
	Scopes    []string  `json:"scopes"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateAPITokenRequest represents the request to create an API token
type CreateAPITokenRequest struct {
	Name     string   `json:"name"`
	Scopes   []string `json:"scopes"`
	Duration string   `json:"duration"`
}

// ErrorResponse represents a standard API error
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// ServiceIdentity represents a machine/service account identity
type ServiceIdentity struct {
	ID          string    `json:"id"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description"`
	OwnerEmail  string    `json:"owner_email"`
	ClientID    string    `json:"client_id"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateServiceIdentityRequest represents the request to create a service identity
type CreateServiceIdentityRequest struct {
	DisplayName string `json:"display_name"`
	Description string `json:"description,omitempty"`
	OwnerEmail  string `json:"owner_email"`
}

// CreateServiceIdentityResponse is returned on creation with the client secret
type CreateServiceIdentityResponse struct {
	ServiceIdentity
	ClientSecret string `json:"client_secret"`
}

// Environment represents an org environment (production, staging, dev, etc.)
type Environment struct {
	ID                     string    `json:"id"`
	OrganizationID         string    `json:"organization_id"`
	Name                   string    `json:"name"`
	Slug                   string    `json:"slug"`
	IsProduction           bool      `json:"is_production"`
	IncidentAutoEscalation bool      `json:"incident_auto_escalation"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// CreateEnvironmentRequest represents the request to add an environment
type CreateEnvironmentRequest struct {
	Name                   string `json:"name"`
	Slug                   string `json:"slug,omitempty"`
	IsProduction           bool   `json:"is_production,omitempty"`
	IncidentAutoEscalation bool   `json:"incident_auto_escalation,omitempty"`
}

// UpdateEnvironmentRequest is used to patch an environment
type UpdateEnvironmentRequest struct {
	Name                   string `json:"name,omitempty"`
	IsProduction           *bool  `json:"is_production,omitempty"`
	IncidentAutoEscalation *bool  `json:"incident_auto_escalation,omitempty"`
}

// RenewalPolicy represents a certificate renewal policy
type RenewalPolicy struct {
	ID               string    `json:"id"`
	OrganizationID   string    `json:"organization_id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	LeadDays         int       `json:"lead_days"`
	AutoApprove      bool      `json:"auto_approve"`
	NotifyDaysBefore []int     `json:"notify_days_before"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// CreateRenewalPolicyRequest is the request body for creating a renewal policy
type CreateRenewalPolicyRequest struct {
	Name             string `json:"name"`
	Description      string `json:"description,omitempty"`
	LeadDays         int    `json:"lead_days"`
	AutoApprove      bool   `json:"auto_approve,omitempty"`
	NotifyDaysBefore []int  `json:"notify_days_before,omitempty"`
}

// IdentityProvider represents a SAML/LDAP/OIDC/SCIM identity provider
type IdentityProvider struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Name         string                 `json:"name"`
	Config       map[string]interface{} `json:"config"`
	SyncSchedule string                 `json:"sync_schedule"`
	Status       string                 `json:"status"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// CreateIdentityProviderRequest is the request body for creating an IDP
type CreateIdentityProviderRequest struct {
	Type         string                 `json:"type"`
	Name         string                 `json:"name"`
	Config       map[string]interface{} `json:"config,omitempty"`
	SyncSchedule string                 `json:"sync_schedule,omitempty"`
}

// Keytab represents a Kerberos keytab
type Keytab struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Principal   string     `json:"principal"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ImportKeytabRequest is the request body for importing a keytab
type ImportKeytabRequest struct {
	Name        string `json:"name"`
	Principal   string `json:"principal"`
	Description string `json:"description,omitempty"`
	KeytabB64   string `json:"keytab_base64"`
}

// UpdateKeytabRequest is the request body for updating a keytab
type UpdateKeytabRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// ApprovalPolicy represents a project-level approval policy for secret/cert access
type ApprovalPolicy struct {
	ID                string    `json:"id"`
	ProjectID         string    `json:"project_id"`
	Name              string    `json:"name"`
	Environments      []string  `json:"environments"`
	RequiredApprovals int       `json:"required_approvals"`
	ApproversGroupID  string    `json:"approvers_group_id,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// CreateApprovalPolicyRequest is the request body for creating an approval policy
type CreateApprovalPolicyRequest struct {
	Name              string   `json:"name"`
	Environments      []string `json:"environments,omitempty"`
	RequiredApprovals int      `json:"required_approvals"`
	ApproversGroupID  string   `json:"approvers_group_id,omitempty"`
}

// User represents a MazeVault user account
type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}

// CreateUserRequest is the request body for creating a user
type CreateUserRequest struct {
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Password string `json:"password"`
	Role     string `json:"role,omitempty"`
}

// UserRole represents a role assignment for a user
type UserRole struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	RoleID      string `json:"role_id"`
	ProjectID   string `json:"project_id,omitempty"`
	Environment string `json:"environment,omitempty"`
}

// AssignRoleRequest assigns a role to a user
type AssignRoleRequest struct {
	RoleID      string `json:"role_id"`
	ProjectID   string `json:"project_id,omitempty"`
	Environment string `json:"environment,omitempty"`
}

// Role represents an RBAC role
type Role struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	DisplayName  string   `json:"display_name"`
	Description  string   `json:"description"`
	Permissions  []string `json:"permissions"`
	IsSystemRole bool     `json:"is_system_role"`
}

// CreateRoleRequest is the request body for creating a role
type CreateRoleRequest struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name,omitempty"`
	Description string   `json:"description,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

// CAAccount represents an organization-level CA account (DigiCert, Venafi, ADCS, ACME…)
type CAAccount struct {
	ID             string                 `json:"id"`
	OrganizationID string                 `json:"organization_id"`
	Name           string                 `json:"name"`
	ProviderType   string                 `json:"provider_type"`
	Config         map[string]interface{} `json:"config"`
	Status         string                 `json:"status"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// CreateCAAccountRequest is the request body for connecting a CA account
type CreateCAAccountRequest struct {
	Name         string                 `json:"name"`
	ProviderType string                 `json:"provider_type"`
	Config       map[string]interface{} `json:"config,omitempty"`
}

// ConfigTemplate represents an org-level config management template
type ConfigTemplate struct {
	ID             string                 `json:"id"`
	OrganizationID string                 `json:"organization_id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	Format         string                 `json:"format"`
	Template       string                 `json:"template"`
	Variables      map[string]interface{} `json:"variables"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// CreateConfigTemplateRequest is the request body for creating a config template
type CreateConfigTemplateRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Format      string                 `json:"format,omitempty"`
	Template    string                 `json:"template"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
}

// Deployment represents an agent deployment package
type Deployment struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	OrganizationID string     `json:"organization_id"`
	OSType         string     `json:"os_type"`
	AgentMode      string     `json:"agent_mode"`
	DeploymentType string     `json:"deployment_type"`
	GatewayURL     string     `json:"gateway_url"`
	Token          string     `json:"token,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	AutoUpdate     bool       `json:"auto_update"`
	CreatedAt      time.Time  `json:"created_at"`
}

// CreateDeploymentRequest is the request body for creating a deployment
type CreateDeploymentRequest struct {
	Name           string `json:"name"`
	OSType         string `json:"os_type"`
	AgentMode      string `json:"agent_mode,omitempty"`
	DeploymentType string `json:"deployment_type,omitempty"`
	GatewayURL     string `json:"gateway_url,omitempty"`
	AutoUpdate     bool   `json:"auto_update,omitempty"`
}

// SharedSecret represents a time-limited one-way secret share
type SharedSecret struct {
	ID             string    `json:"id"`
	ContentType    string    `json:"content_type"`
	RecipientEmail string    `json:"recipient_email"`
	ExpiresAt      time.Time `json:"expires_at"`
	MaxViews       int       `json:"max_views"`
	ViewCount      int       `json:"view_count"`
	CreatedAt      time.Time `json:"created_at"`
}

// CreateSharedSecretRequest is the request body for sharing a secret
type CreateSharedSecretRequest struct {
	SecretID       string `json:"source_id,omitempty"`
	ContentType    string `json:"content_type,omitempty"`
	RecipientEmail string `json:"recipient_email,omitempty"`
	TTLHours       int    `json:"ttl_hours,omitempty"`
	MaxViews       int    `json:"max_views,omitempty"`
	Password       string `json:"password,omitempty"`
}

// CreateSharedSecretResponse includes the share URL returned on creation
type CreateSharedSecretResponse struct {
	SharedSecret
	ShareURL string `json:"share_url"`
}

// ProjectSettings represents configurable project-level settings
type ProjectSettings struct {
	ProjectID            string   `json:"project_id"`
	RetentionDays        int      `json:"retention_days"`
	NotificationWebhooks []string `json:"notification_webhooks"`
	SlackChannel         string   `json:"slack_channel"`
	Tags                 []string `json:"tags"`
	OwnerEmail           string   `json:"owner_email"`
	SyncEnabled          *bool    `json:"sync_enabled,omitempty"`
	SyncIntervalMinutes  *int     `json:"sync_interval_minutes,omitempty"`
}

// UpdateProjectSettingsRequest is the request body for updating project settings
type UpdateProjectSettingsRequest struct {
	RetentionDays        int      `json:"retention_days,omitempty"`
	NotificationWebhooks []string `json:"notification_webhooks,omitempty"`
	SlackChannel         string   `json:"slack_channel,omitempty"`
	Tags                 []string `json:"tags,omitempty"`
	OwnerEmail           string   `json:"owner_email,omitempty"`
	SyncEnabled          *bool    `json:"sync_enabled,omitempty"`
	SyncIntervalMinutes  *int     `json:"sync_interval_minutes,omitempty"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID         string                 `json:"id"`
	UserID     string                 `json:"user_id"`
	ProjectID  string                 `json:"project_id,omitempty"`
	Action     string                 `json:"action"`
	EntityType string                 `json:"entity_type"`
	EntityID   string                 `json:"entity_id"`
	IPAddress  string                 `json:"ip_address"`
	Severity   string                 `json:"severity"`
	Details    string                 `json:"details"`
	CreatedAt  time.Time              `json:"created_at"`
	EventData  map[string]interface{} `json:"event_data,omitempty"`
}

// ListAuditLogsResponse wraps the paginated audit log response
type ListAuditLogsResponse struct {
	Logs  []AuditLog `json:"logs"`
	Total int        `json:"total"`
}

// RotationExecution represents a rotation execution record
type RotationExecution struct {
	ID          string     `json:"id"`
	ConfigID    string     `json:"config_id"`
	Status      string     `json:"status"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Error       string     `json:"error,omitempty"`
}

// RenewalQueueItem represents an item in the certificate renewal queue
type RenewalQueueItem struct {
	ID            string     `json:"id"`
	CertificateID string     `json:"certificate_id"`
	Status        string     `json:"status"`
	RequestedAt   time.Time  `json:"requested_at"`
	ApprovedAt    *time.Time `json:"approved_at,omitempty"`
}
