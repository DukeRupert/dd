// pocketbase/client.go
package pocketbase

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
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

// sendRequest sends an authenticated request and returns the response body
func (c *Client) sendRequest(method, collection, endpoint string, params QueryParams) ([]byte, error) {
    // Build URL with query parameters
    u, err := url.Parse(c.BaseURL + endpoint)
    if err != nil {
        c.Logger.Error().Err(err).Str("collection", collection).Msg("Failed to parse URL")
        return nil, err
    }
    
    if params != (QueryParams{}) {
        u.RawQuery = params.ToURLValues().Encode()
    }
    
    // Create request
    req, err := http.NewRequest(method, u.String(), nil)
    if err != nil {
        c.Logger.Error().Err(err).Str("collection", collection).Msg("Failed to create request")
        return nil, err
    }
    
    // Set authentication header if authenticated
    if c.IsAuthenticated() {
        req.Header.Set("Authorization", c.authToken)
    }
    
    // Set content type
    req.Header.Set("Content-Type", "application/json")
    
    // Execute request
    c.Logger.Debug().Str("collection", collection).Msg("Sending request")
    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        c.Logger.Error().Err(err).Str("collection", collection).Msg("Request failed")
        return nil, err
    }
    defer resp.Body.Close()
    
    // Read response body
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        c.Logger.Error().Err(err).Msg("Failed to read response body")
        return nil, err
    }
    
    // Check for error status codes
    if resp.StatusCode >= 400 {
        c.Logger.Error().Int("status_code", resp.StatusCode).Str("response", string(body)).Msg("Request failed")
        return nil, errors.New("request failed with status " + resp.Status + ": " + string(body))
    }
    
    return body, nil
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
