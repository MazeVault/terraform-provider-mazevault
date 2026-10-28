package mazevault

import "net/http"

// ListDeployments lists agent deployment packages
func (c *Client) ListDeployments() ([]Deployment, error) {
	r, err := c.newRequest(http.MethodGet, "/api/v1/deployments/", nil)
	if err != nil {
		return nil, err
	}
	var resp []Deployment
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// CreateDeployment creates a new agent deployment package
func (c *Client) CreateDeployment(req *CreateDeploymentRequest) (*Deployment, error) {
	r, err := c.newRequest(http.MethodPost, "/api/v1/deployments/", req)
	if err != nil {
		return nil, err
	}
	var resp Deployment
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetDeployment retrieves a deployment by scanning the list (no dedicated GET endpoint)
func (c *Client) GetDeployment(id string) (*Deployment, error) {
	all, err := c.ListDeployments()
	if err != nil {
		return nil, err
	}
	for i, d := range all {
		if d.ID == id {
			return &all[i], nil
		}
	}
	return nil, nil
}

// DeleteDeployment soft-deletes a deployment record, invalidating its bootstrap
// token and generated install script. It does NOT affect agents already
// registered from the deployment. Scoped to the owning organization.
func (c *Client) DeleteDeployment(id, organizationID string) error {
	r, err := c.newRequest(http.MethodDelete, "/api/v1/deployments/"+id+"?organization_id="+organizationID, nil)
	if err != nil {
		return err
	}
	return c.Do(r, nil)
}
