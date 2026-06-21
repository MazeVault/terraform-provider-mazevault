package mazevault

import (
	"fmt"
	"net/http"
)

// Login authenticates with the API and sets the client token
func (c *Client) Login(email, password string) (*LoginResponse, error) {
	reqBody := LoginRequest{
		Email:    email,
		Password: password,
	}

	req, err := c.newRequest(http.MethodPost, "/auth/login", reqBody)
	if err != nil {
		return nil, err
	}

	var resp LoginResponse
	if err := c.Do(req, &resp); err != nil {
		return nil, err
	}

	c.SetToken(resp.AccessToken)
	return &resp, nil
}

// RefreshToken refreshes the access token
func (c *Client) RefreshToken() (*LoginResponse, error) {
	req, err := c.newRequest(http.MethodPost, "/auth/refresh", nil)
	if err != nil {
		return nil, err
	}

	var resp LoginResponse
	if err := c.Do(req, &resp); err != nil {
		return nil, err
	}

	c.SetToken(resp.AccessToken)
	return &resp, nil
}

// clientCredentialsRequest is the body for OAuth2 client-credentials grant
type clientCredentialsRequest struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type clientCredentialsResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// ClientCredentials exchanges a client_id / client_secret for an access token and
// stores it on the client for all subsequent calls.
func (c *Client) ClientCredentials(clientID, clientSecret string) error {
	body := clientCredentialsRequest{
		GrantType:    "client_credentials",
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
	req, err := c.newRequest(http.MethodPost, "/api/v1/auth/token", body)
	if err != nil {
		return err
	}
	var resp clientCredentialsResponse
	if err := c.Do(req, &resp); err != nil {
		return fmt.Errorf("client_credentials grant failed: %w", err)
	}
	c.SetToken(resp.AccessToken)
	return nil
}
