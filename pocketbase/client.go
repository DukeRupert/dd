// pocketbase/client.go
package pocketbase

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

// Client represents a PocketBase API client
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Logger     *zerolog.Logger
	authToken  string
}

// AuthResponse represents the authentication response structure
type AuthResponse struct {
	Record struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
		// Add other fields if needed
	} `json:"record"`
	Token string `json:"token"`
}

// NewClient creates a new PocketBase client
func NewClient(baseURL string, logger *zerolog.Logger) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: time.Second * 10, // Default timeout of 10 seconds
		},
		Logger: logger,
	}
}

// Authenticate performs authentication and stores the token
func (c *Client) Authenticate(identity, password string) error {
	endpoint := "/api/collections/users/auth-with-password"

	c.Logger.Info().Msg("Authenticating user")

	// Create the JSON payload
	requestBody, err := json.Marshal(map[string]string{
		"identity": identity,
		"password": password,
	})
	if err != nil {
		c.Logger.Error().Err(err).Msg("Failed to marshal authentication request")
		return err
	}

	// Create the request
	req, err := http.NewRequest("POST", c.BaseURL+endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		c.Logger.Error().Err(err).Msg("Failed to create authentication request")
		return err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	c.Logger.Debug().Msg("Sending authentication request")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		c.Logger.Error().Err(err).Msg("Authentication request failed")
		return err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Logger.Error().Err(err).Msg("Failed to read authentication response")
		return err
	}

	// Check for successful status code
	if resp.StatusCode != http.StatusOK {
		c.Logger.Error().Int("status_code", resp.StatusCode).Str("response", string(body)).Msg("Authentication failed")
		return errors.New("authentication failed: " + string(body))
	}

	// Parse JSON response
	var authResponse AuthResponse
	err = json.Unmarshal(body, &authResponse)
	if err != nil {
		c.Logger.Error().Err(err).Msg("Failed to unmarshal authentication response")
		return err
	}

	// Store the token
	c.authToken = authResponse.Token
	c.Logger.Info().Msg("Authentication successful")
	return nil
}

// IsAuthenticated checks if the client has an auth token
func (c *Client) IsAuthenticated() bool {
	return c.authToken != ""
}

// Get performs an authenticated GET request
func (c *Client) Get(endpoint string) ([]byte, error) {
	return c.Request("GET", endpoint, nil)
}

// Post performs an authenticated POST request
func (c *Client) Post(endpoint string, data interface{}) ([]byte, error) {
	requestBody, err := json.Marshal(data)
	if err != nil {
		c.Logger.Error().Err(err).Msg("Failed to marshal request body")
		return nil, err
	}
	return c.Request("POST", endpoint, requestBody)
}

// Request performs an authenticated request with the given method
func (c *Client) Request(method, endpoint string, body []byte) ([]byte, error) {
	// Ensure we have a token
	if !c.IsAuthenticated() {
		c.Logger.Warn().Msg("Attempted to make request without authentication")
		return nil, errors.New("client not authenticated")
	}

	// Create request
	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequest(method, c.BaseURL+endpoint, bytes.NewBuffer(body))
	} else {
		req, err = http.NewRequest(method, c.BaseURL+endpoint, nil)
	}

	if err != nil {
		c.Logger.Error().Err(err).Str("method", method).Str("endpoint", endpoint).Msg("Failed to create request")
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.authToken)

	// Send the request
	c.Logger.Debug().Str("method", method).Str("endpoint", endpoint).Msg("Sending request")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		c.Logger.Error().Err(err).Str("method", method).Str("endpoint", endpoint).Msg("Request failed")
		return nil, err
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Logger.Error().Err(err).Msg("Failed to read response body")
		return nil, err
	}

	// Check for error status codes
	if resp.StatusCode >= 400 {
		c.Logger.Error().Int("status_code", resp.StatusCode).Str("response", string(responseBody)).Msg("Request failed")
		return responseBody, errors.New("request failed with status " + resp.Status + ": " + string(responseBody))
	}

	c.Logger.Debug().Str("method", method).Str("endpoint", endpoint).Int("status_code", resp.StatusCode).Msg("Request successful")
	return responseBody, nil
}
