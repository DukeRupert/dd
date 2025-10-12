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