package handler

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

// TestCreateArtistRequest_Validation tests application-layer validation
func TestCreateArtistRequest_Validation(t *testing.T) {
	validate := validator.New()

	type CreateArtistRequest struct {
		Name string `validate:"required,min=2,max=100"`
	}

	tests := []struct {
		name      string
		request   CreateArtistRequest
		wantError bool
	}{
		{"valid name", CreateArtistRequest{Name: "The Beatles"}, false},
		{"minimum length (2 chars)", CreateArtistRequest{Name: "U2"}, false},
		{"one character", CreateArtistRequest{Name: "X"}, true},
		{"empty string", CreateArtistRequest{Name: ""}, true},
		{"whitespace only", CreateArtistRequest{Name: "   "}, false}, // validator counts whitespace
		{"maximum length (100 chars)", CreateArtistRequest{Name: string(make([]byte, 100))}, false},
		{"exceeds maximum (101 chars)", CreateArtistRequest{Name: string(make([]byte, 101))}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.request)
			if (err != nil) != tt.wantError {
				if err != nil {
					validationErrs := err.(validator.ValidationErrors)
					t.Errorf("Validation error = %v (field: %s, tag: %s), wantError %v",
						err, validationErrs[0].Field(), validationErrs[0].Tag(), tt.wantError)
				} else {
					t.Errorf("Validation error = nil, wantError %v", tt.wantError)
				}
			}
		})
	}
}

// TestSignupRequest_Validation tests signup validation rules
func TestSignupRequest_Validation(t *testing.T) {
	validate := validator.New()

	type SignupRequest struct {
		Email    string `validate:"required,email"`
		Username string `validate:"required,min=3,max=50"`
		Password string `validate:"required,min=8"`
	}

	tests := []struct {
		name      string
		request   SignupRequest
		wantError bool
	}{
		{
			"valid signup",
			SignupRequest{Email: "test@example.com", Username: "testuser", Password: "password123"},
			false,
		},
		{
			"invalid email",
			SignupRequest{Email: "notanemail", Username: "testuser", Password: "password123"},
			true,
		},
		{
			"empty email",
			SignupRequest{Email: "", Username: "testuser", Password: "password123"},
			true,
		},
		{
			"username too short",
			SignupRequest{Email: "test@example.com", Username: "ab", Password: "password123"},
			true,
		},
		{
			"username minimum length",
			SignupRequest{Email: "test@example.com", Username: "abc", Password: "password123"},
			false,
		},
		{
			"username too long",
			SignupRequest{Email: "test@example.com", Username: string(make([]byte, 51)), Password: "password123"},
			true,
		},
		{
			"password too short",
			SignupRequest{Email: "test@example.com", Username: "testuser", Password: "short"},
			true,
		},
		{
			"password minimum length",
			SignupRequest{Email: "test@example.com", Username: "testuser", Password: "12345678"},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.request)
			if (err != nil) != tt.wantError {
				if err != nil {
					validationErrs := err.(validator.ValidationErrors)
					t.Errorf("Validation error = %v (field: %s, tag: %s), wantError %v",
						err, validationErrs[0].Field(), validationErrs[0].Tag(), tt.wantError)
				} else {
					t.Errorf("Validation error = nil, wantError %v", tt.wantError)
				}
			}
		})
	}
}

// TestLoginRequest_Validation tests login validation rules
func TestLoginRequest_Validation(t *testing.T) {
	validate := validator.New()

	type LoginRequest struct {
		Email    string `validate:"required,email"`
		Password string `validate:"required"`
	}

	tests := []struct {
		name      string
		request   LoginRequest
		wantError bool
	}{
		{
			"valid login",
			LoginRequest{Email: "test@example.com", Password: "password"},
			false,
		},
		{
			"invalid email",
			LoginRequest{Email: "notanemail", Password: "password"},
			true,
		},
		{
			"empty email",
			LoginRequest{Email: "", Password: "password"},
			true,
		},
		{
			"empty password",
			LoginRequest{Email: "test@example.com", Password: ""},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.request)
			if (err != nil) != tt.wantError {
				if err != nil {
					validationErrs := err.(validator.ValidationErrors)
					t.Errorf("Validation error = %v (field: %s, tag: %s), wantError %v",
						err, validationErrs[0].Field(), validationErrs[0].Tag(), tt.wantError)
				} else {
					t.Errorf("Validation error = nil, wantError %v", tt.wantError)
				}
			}
		})
	}
}

// TestCreateLocationRequest_Validation tests create location validation rules
func TestCreateLocationRequest_Validation(t *testing.T) {
	validate := validator.New()

	type CreateLocationRequest struct {
		Name        string `validate:"required,min=2,max=100"`
		Description string `validate:"max=500"`
		IsDefault   bool
	}

	tests := []struct {
		name      string
		request   CreateLocationRequest
		wantError bool
	}{
		{
			"valid location",
			CreateLocationRequest{Name: "Main Shelf", Description: "Primary storage", IsDefault: false},
			false,
		},
		{
			"minimum name length",
			CreateLocationRequest{Name: "A1", Description: "", IsDefault: false},
			false,
		},
		{
			"name too short",
			CreateLocationRequest{Name: "X", Description: "", IsDefault: false},
			true,
		},
		{
			"empty name",
			CreateLocationRequest{Name: "", Description: "", IsDefault: false},
			true,
		},
		{
			"maximum name length (100 chars)",
			CreateLocationRequest{Name: string(make([]byte, 100)), Description: "", IsDefault: false},
			false,
		},
		{
			"name exceeds maximum (101 chars)",
			CreateLocationRequest{Name: string(make([]byte, 101)), Description: "", IsDefault: false},
			true,
		},
		{
			"empty description allowed",
			CreateLocationRequest{Name: "Valid Name", Description: "", IsDefault: false},
			false,
		},
		{
			"maximum description length (500 chars)",
			CreateLocationRequest{Name: "Valid Name", Description: string(make([]byte, 500)), IsDefault: false},
			false,
		},
		{
			"description exceeds maximum (501 chars)",
			CreateLocationRequest{Name: "Valid Name", Description: string(make([]byte, 501)), IsDefault: false},
			true,
		},
		{
			"with default flag true",
			CreateLocationRequest{Name: "Main Collection", Description: "Primary", IsDefault: true},
			false,
		},
		{
			"with default flag false",
			CreateLocationRequest{Name: "Secondary Shelf", Description: "Backup", IsDefault: false},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.request)
			if (err != nil) != tt.wantError {
				if err != nil {
					validationErrs := err.(validator.ValidationErrors)
					t.Errorf("Validation error = %v (field: %s, tag: %s), wantError %v",
						err, validationErrs[0].Field(), validationErrs[0].Tag(), tt.wantError)
				} else {
					t.Errorf("Validation error = nil, wantError %v", tt.wantError)
				}
			}
		})
	}
}

// TestUpdateLocationRequest_Validation tests update location validation rules
func TestUpdateLocationRequest_Validation(t *testing.T) {
	validate := validator.New()

	type UpdateLocationRequest struct {
		Name        string `validate:"required,min=2,max=100"`
		Description string `validate:"max=500"`
		IsDefault   bool
	}

	tests := []struct {
		name      string
		request   UpdateLocationRequest
		wantError bool
	}{
		{
			"valid update",
			UpdateLocationRequest{Name: "Updated Shelf", Description: "New description", IsDefault: false},
			false,
		},
		{
			"minimum name length",
			UpdateLocationRequest{Name: "B2", Description: "", IsDefault: false},
			false,
		},
		{
			"name too short",
			UpdateLocationRequest{Name: "Y", Description: "", IsDefault: false},
			true,
		},
		{
			"empty name",
			UpdateLocationRequest{Name: "", Description: "Has description", IsDefault: false},
			true,
		},
		{
			"maximum name length (100 chars)",
			UpdateLocationRequest{Name: string(make([]byte, 100)), Description: "", IsDefault: false},
			false,
		},
		{
			"name exceeds maximum (101 chars)",
			UpdateLocationRequest{Name: string(make([]byte, 101)), Description: "", IsDefault: false},
			true,
		},
		{
			"clear description (empty string)",
			UpdateLocationRequest{Name: "Valid Name", Description: "", IsDefault: false},
			false,
		},
		{
			"maximum description length (500 chars)",
			UpdateLocationRequest{Name: "Valid Name", Description: string(make([]byte, 500)), IsDefault: false},
			false,
		},
		{
			"description exceeds maximum (501 chars)",
			UpdateLocationRequest{Name: "Valid Name", Description: string(make([]byte, 501)), IsDefault: false},
			true,
		},
		{
			"update to default",
			UpdateLocationRequest{Name: "Main Collection", Description: "Now primary", IsDefault: true},
			false,
		},
		{
			"update from default to non-default",
			UpdateLocationRequest{Name: "Secondary", Description: "No longer primary", IsDefault: false},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.request)
			if (err != nil) != tt.wantError {
				if err != nil {
					validationErrs := err.(validator.ValidationErrors)
					t.Errorf("Validation error = %v (field: %s, tag: %s), wantError %v",
						err, validationErrs[0].Field(), validationErrs[0].Tag(), tt.wantError)
				} else {
					t.Errorf("Validation error = nil, wantError %v", tt.wantError)
				}
			}
		})
	}
}

// TestUpdateLocationNameRequest_Validation tests the simpler name-only update
func TestUpdateLocationNameRequest_Validation(t *testing.T) {
	validate := validator.New()

	type UpdateLocationNameRequest struct {
		Name string `validate:"required,min=2,max=100"`
	}

	tests := []struct {
		name      string
		request   UpdateLocationNameRequest
		wantError bool
	}{
		{"valid name", UpdateLocationNameRequest{Name: "New Name"}, false},
		{"minimum length", UpdateLocationNameRequest{Name: "AB"}, false},
		{"too short", UpdateLocationNameRequest{Name: "X"}, true},
		{"empty", UpdateLocationNameRequest{Name: ""}, true},
		{"maximum length", UpdateLocationNameRequest{Name: string(make([]byte, 100))}, false},
		{"exceeds maximum", UpdateLocationNameRequest{Name: string(make([]byte, 101))}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.request)
			if (err != nil) != tt.wantError {
				if err != nil {
					validationErrs := err.(validator.ValidationErrors)
					t.Errorf("Validation error = %v (field: %s, tag: %s), wantError %v",
						err, validationErrs[0].Field(), validationErrs[0].Tag(), tt.wantError)
				} else {
					t.Errorf("Validation error = nil, wantError %v", tt.wantError)
				}
			}
		})
	}
}