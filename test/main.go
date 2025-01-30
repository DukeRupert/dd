package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"
)

// ValidationFunc is a function type for custom field validation
type ValidationFunc func(interface{}) error

// TestCase represents a single HTTP request test case
type TestCase struct {
	Name             string
	Method           string
	URL              string
	Headers          map[string]string
	Body             interface{}
	ExpectedStatus   int
	ExpectedBody     interface{}
	CustomValidators map[string]ValidationFunc
}

// TestResult stores the result of a test case execution
type TestResult struct {
	TestCase TestCase
	Passed   bool
	Error    string
}

// TestSuite represents a collection of test cases
type TestSuite struct {
	BaseURL    string
	TestCases  []TestCase
	Results    []TestResult
	httpClient *http.Client
}

// NewTestSuite creates a new test suite with the given base URL
func NewTestSuite(baseURL string) *TestSuite {
	return &TestSuite{
		BaseURL: baseURL,
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

// AddTest adds a new test case to the suite
func (ts *TestSuite) AddTest(tc TestCase) {
	ts.TestCases = append(ts.TestCases, tc)
}

// RunTests executes all test cases in the suite
func (ts *TestSuite) RunTests() {
	ts.Results = make([]TestResult, 0)

	for _, tc := range ts.TestCases {
		result := ts.runTest(tc)
		ts.Results = append(ts.Results, result)
	}
}

// Parse dot notation path to access nested fields
func getNestedValue(data map[string]interface{}, path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts[:len(parts)-1] {
		next, ok := current[part].(map[string]interface{})
		if !ok {
			return nil, false
		}
		current = next
	}

	val, exists := current[parts[len(parts)-1]]
	return val, exists
}

func validateResponse(actual, expected interface{}, validators map[string]ValidationFunc) error {
	actualMap, actualOk := actual.(map[string]interface{})
	expectedMap, expectedOk := expected.(map[string]interface{})

	if !actualOk || !expectedOk {
		return errors.New("both actual and expected must be objects")
	}

	// First check custom validators
	for path, validator := range validators {
		value, exists := getNestedValue(actualMap, path)
		if !exists {
			return fmt.Errorf("missing field for validation: %s", path)
		}
		if err := validator(value); err != nil {
			return fmt.Errorf("validation failed for field %s: %v", path, err)
		}
	}

	// Then check the rest of the fields
	for key, expectedValue := range expectedMap {
		// Skip fields that have custom validators
		if _, hasValidator := validators[key]; hasValidator {
			continue
		}

		actualValue, exists := actualMap[key]
		if !exists {
			return fmt.Errorf("missing field: %s", key)
		}

		// For nested objects, recurse
		if nestedExpected, isObject := expectedValue.(map[string]interface{}); isObject {
			if nestedActual, ok := actualValue.(map[string]interface{}); ok {
				// Create new validators map for nested fields
				nestedValidators := make(map[string]ValidationFunc)
				for path, validator := range validators {
					if strings.HasPrefix(path, key+".") {
						nestedPath := strings.TrimPrefix(path, key+".")
						nestedValidators[nestedPath] = validator
					}
				}
				if err := validateResponse(nestedActual, nestedExpected, nestedValidators); err != nil {
					return err
				}
				continue
			}
		}

		// For non-validated fields, require exact match
		if _, hasNestedValidator := validators[key]; !hasNestedValidator {
			expectedJSON, _ := json.Marshal(expectedValue)
			actualJSON, _ := json.Marshal(actualValue)
			if string(expectedJSON) != string(actualJSON) {
				return fmt.Errorf("field %s mismatch. Expected: %s, Got: %s", key, string(expectedJSON), string(actualJSON))
			}
		}
	}

	return nil
}

func validateID(value interface{}) error {
	// Check if it's a number (float64 due to JSON parsing)
	id, ok := value.(float64)
	if !ok {
		return fmt.Errorf("expected number, got %T", value)
	}

	// Check if it's a positive integer
	if id <= 0 || math.Floor(id) != id {
		return fmt.Errorf("expected positive integer, got %v", id)
	}

	return nil
}

func validateTimestamp(value interface{}) error {
	timestamp, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string timestamp, got %T", value)
	}

	_, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return fmt.Errorf("invalid timestamp format: %v", err)
	}

	return nil
}

func validateRequestID(value interface{}) error {
	id, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string request ID, got %T", value)
	}
	if len(id) == 0 {
		return fmt.Errorf("request ID cannot be empty")
	}
	return nil
}

// runTest executes a single test case
func (ts *TestSuite) runTest(tc TestCase) TestResult {
	result := TestResult{
		TestCase: tc,
		Passed:   false,
	}

	// Prepare request body
	var bodyReader io.Reader
	if tc.Body != nil {
		jsonBody, err := json.Marshal(tc.Body)
		if err != nil {
			result.Error = fmt.Sprintf("Failed to marshal request body: %v", err)
			return result
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	// Create request
	req, err := http.NewRequest(tc.Method, ts.BaseURL+tc.URL, bodyReader)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to create request: %v", err)
		return result
	}

	// Add headers
	for key, value := range tc.Headers {
		req.Header.Add(key, value)
	}
	if tc.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Execute request
	resp, err := ts.httpClient.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("Request failed: %v", err)
		return result
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != tc.ExpectedStatus {
		result.Error = fmt.Sprintf("Expected status %d, got %d", tc.ExpectedStatus, resp.StatusCode)
		return result
	}

	// Check response body if expected body is provided
	if tc.ExpectedBody != nil {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			result.Error = fmt.Sprintf("Failed to read response body: %v", err)
			return result
		}

		var actualBody interface{}
		if err := json.Unmarshal(body, &actualBody); err != nil {
			result.Error = fmt.Sprintf("Failed to unmarshal response body: %v", err)
			return result
		}

		// If there are custom validators, use them instead of exact matching for specific fields
		if len(tc.CustomValidators) > 0 {
			if err := validateResponse(actualBody, tc.ExpectedBody, tc.CustomValidators); err != nil {
				result.Error = err.Error()
				return result
			}
		} else {
			expectedJSON, _ := json.Marshal(tc.ExpectedBody)
			actualJSON, _ := json.Marshal(actualBody)

			if string(expectedJSON) != string(actualJSON) {
				result.Error = fmt.Sprintf("Response body mismatch.\nExpected: %s\nGot: %s",
					string(expectedJSON), string(actualJSON))
				return result
			}
		}
	}

	result.Passed = true
	return result
}

// PrintResults prints the test results to stdout
func (ts *TestSuite) PrintResults() {
	fmt.Println("\nTest Results:")
	fmt.Println("=============")

	totalTests := len(ts.Results)
	passedTests := 0

	for _, result := range ts.Results {
		if result.Passed {
			passedTests++
			fmt.Printf("✅ %s: PASSED\n", result.TestCase.Name)
		} else {
			fmt.Printf("❌ %s: FAILED\n", result.TestCase.Name)
			fmt.Printf("   Error: %s\n", result.Error)
		}
	}

	fmt.Printf("\nSummary: %d/%d tests passed\n", passedTests, totalTests)
}

func main() {
	// Create a new test suite
	suite := NewTestSuite("http://localhost:8080")

	// Define a validator for non-empty string tokens
	validateToken := func(value interface{}) error {
		str, ok := value.(string)
		if !ok {
			return errors.New("token must be a string")
		}
		if len(str) == 0 {
			return errors.New("token must not be empty")
		}
		return nil
	}

	// Test user registration
	suite.AddTest(TestCase{
		Name:   "User Registration",
		Method: "POST",
		URL:    "/register",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: map[string]interface{}{
			"email":      "test3@example.com",
			"password":   "securepassword123",
			"first_name": "John",
			"last_name":  "Doe",
		},
		ExpectedStatus: http.StatusCreated,
		ExpectedBody: map[string]interface{}{
			"user": map[string]interface{}{
				"id":         "", // Empty string means we'll validate it's any positive number
				"email":      "test3@example.com",
				"first_name": "John",
				"last_name":  "Doe",
				"created_at": "", // Will be validated by validateTimestamp
			},
			"token": "", // Will be validated by validateToken
		},
		CustomValidators: map[string]ValidationFunc{
			"token":           validateToken,
			"user.created_at": validateTimestamp,
			"user.id":         validateID,
		},
	})
	// Test duplicate registration
	suite.AddTest(TestCase{
		Name:   "Duplicate Registration",
		Method: "POST",
		URL:    "/register",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: map[string]interface{}{
			"email":      "test@example.com",
			"password":   "securepassword123",
			"first_name": "John",
			"last_name":  "Doe",
		},
		ExpectedStatus: http.StatusBadRequest,
		ExpectedBody: map[string]interface{}{
			"status_code": float64(400),
			"error": map[string]interface{}{
				"code":    "BAD_REQUEST",
				"message": "email already registered",
			},
			"request_id": "", // will be validated
			"timestamp":  "", // will be validated
		},
		CustomValidators: map[string]ValidationFunc{
			"request_id": validateRequestID,
			"timestamp":  validateTimestamp,
		},
	})

	// Test user login
	suite.AddTest(TestCase{
		Name:   "User Login",
		Method: "POST",
		URL:    "/login",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: map[string]interface{}{
			"email":    "test@example.com",
			"password": "securepassword123",
		},
		ExpectedStatus: http.StatusOK,
		ExpectedBody: map[string]interface{}{
			"refresh_token": "", // will be validated by CustomValidators
			"token":         "", // will be validated by CustomValidators
			"user": map[string]interface{}{
				"email":      "test@example.com",
				"first_name": "John",
				"last_name":  "Doe",
				"created_at": "", // will be validated by CustomValidators
			},
		},
		CustomValidators: map[string]ValidationFunc{
			"token":           validateToken,
			"refresh_token":   validateToken,
			"user.created_at": validateTimestamp,
		},
	})

	// Test login with wrong password
	suite.AddTest(TestCase{
		Name:   "Login Wrong Password",
		Method: "POST",
		URL:    "/login",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: map[string]interface{}{
			"email":    "test@example.com",
			"password": "wrongpassword",
		},
		ExpectedStatus: http.StatusUnauthorized,
		ExpectedBody: map[string]interface{}{
			"status_code": float64(401),
			"error": map[string]interface{}{
				"code":    "UNAUTHORIZED",
				"message": "invalid credentials",
			},
			"request_id": "", // will be validated
			"timestamp":  "", // will be validated
		},
		CustomValidators: map[string]ValidationFunc{
			"request_id": validateRequestID,
			"timestamp":  validateTimestamp,
		},
	})

	// Test refresh token
	suite.AddTest(TestCase{
		Name:   "Refresh Token",
		Method: "POST",
		URL:    "/refresh",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: map[string]interface{}{
			"refresh_token": "valid-refresh-token", // You'll need to extract this from the login response
		},
		ExpectedStatus: http.StatusOK,
		ExpectedBody: map[string]interface{}{
			"token":         "", // will be validated by CustomValidators
			"refresh_token": "", // will be validated by CustomValidators
		},
		CustomValidators: map[string]ValidationFunc{
			"token":         validateToken,
			"refresh_token": validateToken,
		},
	})

	// Test delete user (with auth)
	suite.AddTest(TestCase{
		Name:   "Delete User",
		Method: "DELETE",
		URL:    "/dev/users",
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer your-admin-token-here", // You'll need to add authorization
		},
		Body: map[string]interface{}{
			"email": "test@example.com",
		},
		ExpectedStatus: http.StatusNoContent,
		ExpectedBody:   nil,
	})

	// Run all tests
	suite.RunTests()

	// Print results
	suite.PrintResults()
}
