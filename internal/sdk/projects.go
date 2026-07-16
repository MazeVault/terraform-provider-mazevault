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

// CreateProject creates a new project with the given name, type, and organization.
// projectType is one of "secret" or "certificate" (defaults to "secret" server-side
// when empty).
func (c *Client) CreateProject(name, projectType, organizationID string) (*Project, error) {
	reqBody := CreateProjectRequest{
		Name:           name,
		Type:           projectType,
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

// UpdateProject updates a project's mutable fields via PUT /api/v1/projects/:id.
// Only the name is mutable; project type is immutable server-side.
func (c *Client) UpdateProject(id, name string) (*Project, error) {
	reqBody := UpdateProjectRequest{
		Name: name,
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
