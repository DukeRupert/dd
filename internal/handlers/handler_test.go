package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/dukerupert/dd/internal/handlers"
	"github.com/dukerupert/dd/internal/models"
	pocketbase "github.com/dukerupert/dd/pb"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestLiveAuthentication(t *testing.T) {
	// Skip this test if not in integration test mode
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=true to run.")
	}

	// Get database URL from environment or use default
	dbURL := os.Getenv("POCKETBASE_URL")
	if dbURL == "" {
		dbURL = "http://localhost:8090" // Default Pocketbase URL
	}

	// Create a real client instance
	pbClient := pocketbase.NewClient(dbURL)
	authHandler := handlers.NewAuthHandler(pbClient)

	// Initialize Echo instance
	e := echo.New()

	// Define test credentials - store these in environment variables for security
	testCases := []struct {
		name           string
		email          string
		password       string
		expectedStatus int
	}{
		{
			name:           "valid-credentials",
			email:          os.Getenv("TEST_USER_EMAIL"),    // e.g., "test@example.com"
			password:       os.Getenv("TEST_USER_PASSWORD"), // e.g., "password123"
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid-credentials",
			email:          "test@example.com",
			password:       "wrongpassword",
			expectedStatus: http.StatusUnauthorized,
		},
		// Add more test cases as needed
	}

	// Variable to store a valid token for subsequent tests
	var validToken string

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Skip if credentials not provided (for valid test cases)
			if tc.expectedStatus == http.StatusOK &&
				(os.Getenv("TEST_USER_EMAIL") == "" || os.Getenv("TEST_USER_PASSWORD") == "") {
				t.Skip("Skipping test that requires valid credentials. Set TEST_USER_EMAIL and TEST_USER_PASSWORD environment variables.")
			}

			// Create request
			loginRequest := models.LoginRequest{
				Email:    tc.email,
				Password: tc.password,
			}
			reqBody, _ := json.Marshal(loginRequest)
			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Execute handler
			err := authHandler.Login(c)

			// Add debugging
			// t.Logf("Response status: %d", rec.Code)
			// t.Logf("Response body: %s", rec.Body.String())

			// Assertions
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, rec.Code)

			if tc.expectedStatus == http.StatusOK {
				var response models.LoginResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.Token)
				assert.NotEmpty(t, response.Record.Id)
				assert.Equal(t, tc.email, response.Record.Email)

				// Store the token for later tests
				validToken = response.Token
			}
		})
	}

	// Only run the following tests if we have a valid token
	if validToken != "" {
		// Test auth check with valid token
		t.Run("check-valid-token", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/auth/check", nil)
			req.Header.Set("Authorization", validToken)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := authHandler.CheckAuth(c)

			t.Logf("Check auth response status: %d", rec.Code)
			t.Logf("Check auth response body: %s", rec.Body.String())

			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
		})

		// Test logout
		t.Run("logout", func(t *testing.T) {
			// First set the token in the client (since this might be a different instance)
			pbClient.SetAuthToken(validToken)

			req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := authHandler.Logout(c)

			t.Logf("Logout response status: %d", rec.Code)
			t.Logf("Logout response body: %s", rec.Body.String())

			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)

			// Verify client state after logout
			assert.Empty(t, pbClient.GetAuthToken(), "Token should be cleared after logout")
			assert.False(t, pbClient.IsAuthenticated(), "Client should not be authenticated after logout")
		})

		// Test auth check after logout
		t.Run("check-after-logout", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/auth/check", nil)
			req.Header.Set("Authorization", validToken)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := authHandler.CheckAuth(c)

			t.Logf("Check after logout response status: %d", rec.Code)
			t.Logf("Check after logout response body: %s", rec.Body.String())

			assert.NoError(t, err)
			// This could be 401 or 200 depending on whether your JWT validation actually checks
			// if the token is in a "logged out" state or just checks token format/expiration
			// If using stateless JWT without a blacklist, the token will still be valid
			t.Logf("Note: If using purely stateless JWT, tokens remain valid after logout")
		})
	}
}
