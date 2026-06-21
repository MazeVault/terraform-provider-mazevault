package mazevault

import (
	"fmt"
	"net/http"
)

// ListUsers lists all users in the system
func (c *Client) ListUsers() ([]User, error) {
	r, err := c.newRequest(http.MethodGet, "/api/v1/rbac/users", nil)
	if err != nil {
		return nil, err
	}
	var resp []User
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// CreateUser creates a new user account
func (c *Client) CreateUser(req *CreateUserRequest) (*User, error) {
	r, err := c.newRequest(http.MethodPost, "/api/v1/rbac/users", req)
	if err != nil {
		return nil, err
	}
	var resp User
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteUser deletes a user account
func (c *Client) DeleteUser(id string) error {
	r, err := c.newRequest(http.MethodDelete, "/api/v1/rbac/users/"+id, nil)
	if err != nil {
		return err
	}
	return c.Do(r, nil)
}

// AssignRole assigns a role to a user
func (c *Client) AssignRole(userID string, req *AssignRoleRequest) (*UserRole, error) {
	r, err := c.newRequest(http.MethodPost, fmt.Sprintf("/api/v1/rbac/users/%s/roles", userID), req)
	if err != nil {
		return nil, err
	}
	var resp UserRole
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetUserRoles returns the role assignments for a user
func (c *Client) GetUserRoles(userID string) ([]UserRole, error) {
	r, err := c.newRequest(http.MethodGet, fmt.Sprintf("/api/v1/rbac/users/%s/roles", userID), nil)
	if err != nil {
		return nil, err
	}
	var resp []UserRole
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// RemoveRoleAssignment removes a specific role assignment from a user
func (c *Client) RemoveRoleAssignment(userID, assignmentID string) error {
	r, err := c.newRequest(http.MethodDelete, fmt.Sprintf("/api/v1/rbac/users/%s/roles/%s", userID, assignmentID), nil)
	if err != nil {
		return err
	}
	return c.Do(r, nil)
}

// ListRoles lists all roles in the system
func (c *Client) ListRoles() ([]Role, error) {
	r, err := c.newRequest(http.MethodGet, "/api/v1/rbac/roles", nil)
	if err != nil {
		return nil, err
	}
	var resp []Role
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// CreateRole creates a new RBAC role
func (c *Client) CreateRole(req *CreateRoleRequest) (*Role, error) {
	r, err := c.newRequest(http.MethodPost, "/api/v1/rbac/roles", req)
	if err != nil {
		return nil, err
	}
	var resp Role
	if err := c.Do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
