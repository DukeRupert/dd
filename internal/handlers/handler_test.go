package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/dukerupert/dd/internal/handlers"
	"github.com/dukerupert/dd/pb"
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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Skip if credentials not provided (for valid test cases)
			if tc.expectedStatus == http.StatusOK &&
				(os.Getenv("TEST_USER_EMAIL") == "" || os.Getenv("TEST_USER_PASSWORD") == "") {
				t.Skip("Skipping test that requires valid credentials. Set TEST_USER_EMAIL and TEST_USER_PASSWORD environment variables.")
			}

			// Create request
			loginRequest := handlers.LoginRequest{
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
			t.Logf("Response status: %d", rec.Code)
			t.Logf("Response body: %s", rec.Body.String())

			// Assertions
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, rec.Code)

			if tc.expectedStatus == http.StatusOK {
				var response handlers.LoginResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.Token)
				assert.NotEmpty(t, response.User.ID)
				assert.Equal(t, tc.email, response.User.Email)

				// You can add more assertions here about the returned user
				// For example, check if specific roles or permissions are set
			}
		})
	}
}
