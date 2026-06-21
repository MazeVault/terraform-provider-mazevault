package mazevault

import (
	"fmt"
	"net/http"
)

// ListApprovalPolicies lists approval policies for a project
func (c *Client) ListApprovalPolicies(projectID string) ([]ApprovalPolicy, error) {
	r, err := c.newRequest(http.MethodGet, "/api/v1/projects/"+projectID+"/approval-policies", nil)
	if err != nil {
		return nil, err
	}
	var resp []ApprovalPolicy
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// CreateApprovalPolicy creates an approval policy in a project
func (c *Client) CreateApprovalPolicy(projectID string, req *CreateApprovalPolicyRequest) (*ApprovalPolicy, error) {
	r, err := c.newRequest(http.MethodPost, "/api/v1/projects/"+projectID+"/approval-policies", req)
	if err != nil {
		return nil, err
	}
	var resp ApprovalPolicy
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetApprovalPolicy is synthesised by listing and filtering (no dedicated GET endpoint)
func (c *Client) GetApprovalPolicy(projectID, policyID string) (*ApprovalPolicy, error) {
	all, err := c.ListApprovalPolicies(projectID)
	if err != nil {
		return nil, err
	}
	for i, p := range all {
		if p.ID == policyID {
			return &all[i], nil
		}
	}
	return nil, nil
}

// UpdateApprovalPolicy updates an existing approval policy
func (c *Client) UpdateApprovalPolicy(policyID string, req *CreateApprovalPolicyRequest) (*ApprovalPolicy, error) {
	r, err := c.newRequest(http.MethodPut, fmt.Sprintf("/api/v1/projects/approval-policies/%s", policyID), req)
	if err != nil {
		return nil, err
	}
	var resp ApprovalPolicy
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteApprovalPolicy deletes an approval policy
func (c *Client) DeleteApprovalPolicy(policyID string) error {
	r, err := c.newRequest(http.MethodDelete, fmt.Sprintf("/api/v1/projects/approval-policies/%s", policyID), nil)
	if err != nil {
		return err
	}
	return c.Do(r, nil)
}
