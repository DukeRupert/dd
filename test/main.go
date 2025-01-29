package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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

// validateResponse recursively compares the actual response with the expected response
// using custom validators where specified
func validateResponse(actual, expected interface{}, validators map[string]ValidationFunc) error {
	actualMap, actualOk := actual.(map[string]interface{})
	expectedMap, expectedOk := expected.(map[string]interface{})

	if !actualOk || !expectedOk {
		return errors.New("both actual and expected must be objects")
	}

	for key, expectedValue := range expectedMap {
		actualValue, exists := actualMap[key]
		if !exists {
			return fmt.Errorf("missing field: %s", key)
		}

		// If there's a custom validator for this field, use it
		if validator, hasValidator := validators[key]; hasValidator {
			if err := validator(actualValue); err != nil {
				return fmt.Errorf("validation failed for field %s: %v", key, err)
			}
			continue
		}

		// For nested objects, recurse
		if nestedExpected, isObject := expectedValue.(map[string]interface{}); isObject {
			if nestedActual, ok := actualValue.(map[string]interface{}); ok {
				if err := validateResponse(nestedActual, nestedExpected, validators); err != nil {
					return err
				}
				continue
			}
		}

		// For non-validated fields, require exact match
		expectedJSON, _ := json.Marshal(expectedValue)
		actualJSON, _ := json.Marshal(actualValue)
		if string(expectedJSON) != string(actualJSON) {
			return fmt.Errorf("field %s mismatch. Expected: %s, Got: %s", key, string(expectedJSON), string(actualJSON))
		}
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

	// Add login test case
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
			"refresh_token": "", // actual value will be validated by CustomValidators
			"token":         "", // actual value will be validated by CustomValidators
			"user": map[string]interface{}{
				"id":         1,
				"email":      "test@example.com",
				"first_name": "John",
				"last_name":  "Doe",
				"created_at": "2025-01-29T20:36:29Z",
			},
		},
		CustomValidators: map[string]ValidationFunc{
			"token":         validateToken,
			"refresh_token": validateToken,
		},
	})

	// Add refresh token test case
	suite.AddTest(TestCase{
		Name:   "Refresh Token",
		Method: "POST",
		URL:    "/refresh",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: map[string]interface{}{
			"refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIiwiZXhwIjoxNzM4Nzg5MjMwLCJuYmYiOjE3MzgxODQ0MzAsImlhdCI6MTczODE4NDQzMH0.Y-kCRV_JNSjYYJCGmJTDVBjUqaNcUh_y8YRpDPRQLNc",
		},
		ExpectedStatus: http.StatusOK,
		ExpectedBody: map[string]interface{}{
			"token":         "", // actual value will be validated by CustomValidators
			"refresh_token": "", // actual value will be validated by CustomValidators
		},
		CustomValidators: map[string]ValidationFunc{
			"token":         validateToken,
			"refresh_token": validateToken,
		},
	})

	// Run all tests
	suite.RunTests()

	// Print results
	suite.PrintResults()
}