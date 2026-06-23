package mazevault

import "net/http"

// ListAuditLogs retrieves audit log entries with optional pagination
func (c *Client) ListAuditLogs(limit, offset int) (*ListAuditLogsResponse, error) {
	path := "/api/v1/audit-logs/"
	if limit > 0 {
		path += "?limit=" + itoa(limit) + "&offset=" + itoa(offset)
	}
	r, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var logs []AuditLog
	if err := c.Do(r, &logs); err != nil {
		return nil, err
	}
	return &ListAuditLogsResponse{Logs: logs, Total: len(logs)}, nil
}

// ListProjectAuditLogs retrieves audit log entries for a specific project
func (c *Client) ListProjectAuditLogs(projectID string, limit, offset int) (*ListAuditLogsResponse, error) {
	path := "/api/v1/projects/" + projectID + "/audit-logs"
	if limit > 0 {
		path += "?limit=" + itoa(limit) + "&offset=" + itoa(offset)
	}
	r, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var logs []AuditLog
	if err := c.Do(r, &logs); err != nil {
		return nil, err
	}
	return &ListAuditLogsResponse{Logs: logs, Total: len(logs)}, nil
}

// ListRotationExecutions retrieves rotation execution history
func (c *Client) ListRotationExecutions() ([]RotationExecution, error) {
	r, err := c.newRequest(http.MethodGet, "/api/v1/rotation/executions", nil)
	if err != nil {
		return nil, err
	}
	var resp []RotationExecution
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetRenewalQueue retrieves the certificate renewal queue for an organization
func (c *Client) GetRenewalQueue(organizationID string) ([]RenewalQueueItem, error) {
	r, err := c.newRequest(http.MethodGet, "/api/v1/renewal-queue/?organization_id="+organizationID, nil)
	if err != nil {
		return nil, err
	}
	var resp []RenewalQueueItem
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// CreateSharedSecret creates a one-time secret share link
func (c *Client) CreateSharedSecret(req *CreateSharedSecretRequest) (*CreateSharedSecretResponse, error) {
	r, err := c.newRequest(http.MethodPost, "/api/v1/share/", req)
	if err != nil {
		return nil, err
	}
	var resp CreateSharedSecretResponse
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetSharedSecretMetadata retrieves metadata (not the value) for a shared secret
func (c *Client) GetSharedSecretMetadata(id string) (*SharedSecret, error) {
	r, err := c.newRequest(http.MethodGet, "/api/v1/share/"+id, nil)
	if err != nil {
		return nil, err
	}
	var resp SharedSecret
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// itoa converts int to string without importing strconv
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	buf := make([]byte, 20)
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
