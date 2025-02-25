package pocketbase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client interface defines methods that a Pocketbase client should implement
type IClient interface {
    AuthWithPassword(ctx context.Context, email, password string) (*AuthResponse, error)
    SetAuthToken(token string)
    GetAuthToken() string
    ClearAuth()
    IsAuthenticated() bool
}

// Ensure the concrete Client struct implements the IClient interface
var _ IClient = (*Client)(nil)

// Client represents the Pocketbase API client
type Client struct {
	baseURL    string
	httpClient *http.Client
	authToken  string
}

// AuthResponse represents the response from auth-with-password endpoint
type AuthResponse struct {
	Token  string       `json:"token"`
	Record RecordDetail `json:"record"`
}

// RecordDetail represents user record details
type RecordDetail struct {
	CollectionID   string `json:"collectionId"`
	CollectionName string `json:"collectionName"`
	ID             string `json:"id"`
	Email          string `json:"email"`
	EmailVisibility bool   `json:"emailVisibility"`
	Verified       bool   `json:"verified"`
	Name           string `json:"name"`
	Avatar         string `json:"avatar"`
	Created        string `json:"created"`
	Updated        string `json:"updated"`
}

// AuthRequest represents the login request payload
type AuthRequest struct {
	Identity string `json:"identity"`
	Password string `json:"password"`
}

// Error represents the error response format
type Error struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

// NewClient creates a new Pocketbase client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SetAuthToken sets the auth token for the client
func (c *Client) SetAuthToken(token string) {
	c.authToken = token
}

// GetAuthToken returns the current auth token
func (c *Client) GetAuthToken() string {
	return c.authToken
}

// ClearAuth clears the authentication token
func (c *Client) ClearAuth() {
	c.authToken = ""
}

// IsAuthenticated checks if the client has an auth token
func (c *Client) IsAuthenticated() bool {
	return c.authToken != ""
}

// AuthWithPassword authenticates a user with email and password
func (c *Client) AuthWithPassword(ctx context.Context, email, password string) (*AuthResponse, error) {
	url := fmt.Sprintf("%s/api/collections/users/auth-with-password", c.baseURL)
	
	authReq := AuthRequest{
		Identity: email,
		Password: password,
	}
	
	reqBody, err := json.Marshal(authReq)
	if err != nil {
		return nil, fmt.Errorf("marshaling auth request: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("creating auth request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing auth request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		var pbError Error
		if err := json.NewDecoder(resp.Body).Decode(&pbError); err != nil {
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("authentication failed: %s", pbError.Message)
	}
	
	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, fmt.Errorf("decoding auth response: %w", err)
	}
	
	// Set the token in the client for future requests
	c.SetAuthToken(authResp.Token)
	
	return &authResp, nil
}

// AuthenticatedRequest is a helper method for making authenticated requests
func (c *Client) AuthenticatedRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	if !c.IsAuthenticated() {
		return nil, fmt.Errorf("client is not authenticated")
	}
	
	url := fmt.Sprintf("%s%s", c.baseURL, path)
	
	var reqBody *bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		reqBody = bytes.NewBuffer(b)
	}
	
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.authToken)
	
	return c.httpClient.Do(req)
}