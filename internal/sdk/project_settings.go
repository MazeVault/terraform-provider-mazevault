package mazevault

import "net/http"

// GetProjectSettings retrieves configurable settings for a project
func (c *Client) GetProjectSettings(projectID string) (*ProjectSettings, error) {
	r, err := c.newRequest(http.MethodGet, "/api/v1/projects/"+projectID+"/settings", nil)
	if err != nil {
		return nil, err
	}
	var resp ProjectSettings
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateProjectSettings updates project settings
func (c *Client) UpdateProjectSettings(projectID string, req *UpdateProjectSettingsRequest) (*ProjectSettings, error) {
	r, err := c.newRequest(http.MethodPut, "/api/v1/projects/"+projectID+"/settings", req)
	if err != nil {
		return nil, err
	}
	var resp ProjectSettings
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
