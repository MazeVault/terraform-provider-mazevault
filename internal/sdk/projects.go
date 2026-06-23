package mazevault

import "net/http"

// ListProjects retrieves all projects
func (c *Client) ListProjects() ([]Project, error) {
	req, err := c.newRequest(http.MethodGet, "/api/v1/projects", nil)
	if err != nil {
		return nil, err
	}

	var resp []Project
	if err := c.Do(req, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

// GetProject retrieves a project by ID
func (c *Client) GetProject(id string) (*Project, error) {
	req, err := c.newRequest(http.MethodGet, "/api/v1/projects/"+id, nil)
	if err != nil {
		return nil, err
	}

	var resp Project
	if err := c.Do(req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// CreateProject creates a new project with the given name, environment, and organization.
func (c *Client) CreateProject(name, environment, organizationID string) (*Project, error) {
	reqBody := CreateProjectRequest{
		Name:           name,
		Environment:    environment,
		OrganizationID: organizationID,
	}

	req, err := c.newRequest(http.MethodPost, "/api/v1/projects", reqBody)
	if err != nil {
		return nil, err
	}

	var resp Project
	if err := c.Do(req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// UpdateProject updates a project's name, environment, and organization.
func (c *Client) UpdateProject(id, name, environment, organizationID string) (*Project, error) {
	reqBody := CreateProjectRequest{
		Name:           name,
		Environment:    environment,
		OrganizationID: organizationID,
	}

	req, err := c.newRequest(http.MethodPut, "/api/v1/projects/"+id, reqBody)
	if err != nil {
		return nil, err
	}

	var resp Project
	if err := c.Do(req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// DeleteProject deletes a project
func (c *Client) DeleteProject(id string) error {
	req, err := c.newRequest(http.MethodDelete, "/api/v1/projects/"+id, nil)
	if err != nil {
		return err
	}

	return c.Do(req, nil)
}
