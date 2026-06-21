package mazevault

import (
	"fmt"
	"net/http"
)

// ListRenewalPolicies lists renewal policies for an organization
func (c *Client) ListRenewalPolicies(organizationID string) ([]RenewalPolicy, error) {
	r, err := c.newRequest(http.MethodGet, "/api/v1/organizations/"+organizationID+"/renewal-policies/", nil)
	if err != nil {
		return nil, err
	}
	var resp []RenewalPolicy
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// CreateRenewalPolicy creates a new renewal policy
func (c *Client) CreateRenewalPolicy(organizationID string, req *CreateRenewalPolicyRequest) (*RenewalPolicy, error) {
	r, err := c.newRequest(http.MethodPost, "/api/v1/organizations/"+organizationID+"/renewal-policies/", req)
	if err != nil {
		return nil, err
	}
	var resp RenewalPolicy
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetRenewalPolicy retrieves a single renewal policy
func (c *Client) GetRenewalPolicy(organizationID, policyID string) (*RenewalPolicy, error) {
	r, err := c.newRequest(http.MethodGet, fmt.Sprintf("/api/v1/organizations/%s/renewal-policies/%s", organizationID, policyID), nil)
	if err != nil {
		return nil, err
	}
	var resp RenewalPolicy
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateRenewalPolicy updates a renewal policy
func (c *Client) UpdateRenewalPolicy(organizationID, policyID string, req *CreateRenewalPolicyRequest) (*RenewalPolicy, error) {
	r, err := c.newRequest(http.MethodPut, fmt.Sprintf("/api/v1/organizations/%s/renewal-policies/%s", organizationID, policyID), req)
	if err != nil {
		return nil, err
	}
	var resp RenewalPolicy
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteRenewalPolicy deletes a renewal policy
func (c *Client) DeleteRenewalPolicy(organizationID, policyID string) error {
	r, err := c.newRequest(http.MethodDelete, fmt.Sprintf("/api/v1/organizations/%s/renewal-policies/%s", organizationID, policyID), nil)
	if err != nil {
		return err
	}
	return c.Do(r, nil)
}
