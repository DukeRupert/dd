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

// TestCreateRecordRequest_Validation tests create record validation rules
func TestCreateRecordRequest_Validation(t *testing.T) {
	validate := validator.New()

	type CreateRecordRequest struct {
		Title             string `validate:"required,min=1,max=200"`
		ArtistID          int64  `validate:"omitempty,min=1"`
		AlbumTitle        string `validate:"max=200"`
		ReleaseYear       int32  `validate:"omitempty,min=1900,max=2100"`
		CurrentLocationID int64  `validate:"omitempty,min=1"`
		HomeLocationID    int64  `validate:"omitempty,min=1"`
		CatalogNumber     string `validate:"max=100"`
		Condition         string `validate:"omitempty,oneof=Mint 'Near Mint' 'Very Good' Good Fair Poor"`
		Notes             string `validate:"max=1000"`
	}

	tests := []struct {
		name      string
		request   CreateRecordRequest
		wantError bool
	}{
		{
			"valid record with all fields",
			CreateRecordRequest{
				Title:             "Dark Side of the Moon",
				ArtistID:          1,
				AlbumTitle:        "Dark Side of the Moon",
				ReleaseYear:       1973,
				CurrentLocationID: 1,
				HomeLocationID:    1,
				CatalogNumber:     "SHVL 804",
				Condition:         "Mint",
				Notes:             "Original pressing",
			},
			false,
		},
		{
			"valid record with minimal fields",
			CreateRecordRequest{Title: "Unknown Record"},
			false,
		},
		{
			"empty title",
			CreateRecordRequest{Title: ""},
			true,
		},
		{
			"title too long (201 chars)",
			CreateRecordRequest{Title: string(make([]byte, 201))},
			true,
		},
		{
			"title maximum length (200 chars)",
			CreateRecordRequest{Title: string(make([]byte, 200))},
			false,
		},
		{
			"invalid artist ID (zero)",
			CreateRecordRequest{Title: "Test", ArtistID: 0},
			true,
		},
		{
			"invalid artist ID (negative)",
			CreateRecordRequest{Title: "Test", ArtistID: -1},
			true,
		},
				{
			"valid artist ID",
			CreateRecordRequest{Title: "Test", ArtistID: 1},
			false,
		},
		{
			"album title too long (201 chars)",
			CreateRecordRequest{Title: "Test", AlbumTitle: string(make([]byte, 201))},
			true,
		},
		{
			"album title maximum length (200 chars)",
			CreateRecordRequest{Title: "Test", AlbumTitle: string(make([]byte, 200))},
			false,
		},
		{
			"release year too early (1899)",
			CreateRecordRequest{Title: "Test", ReleaseYear: 1899},
			true,
		},
		{
			"release year minimum (1900)",
			CreateRecordRequest{Title: "Test", ReleaseYear: 1900},
			false,
		},
		{
			"release year too late (2101)",
			CreateRecordRequest{Title: "Test", ReleaseYear: 2101},
			true,
		},
		{
			"release year maximum (2100)",
			CreateRecordRequest{Title: "Test", ReleaseYear: 2100},
			false,
		},
		{
			"invalid current location ID (zero)",
			CreateRecordRequest{Title: "Test", CurrentLocationID: 0},
			true,
		},
		{
			"invalid home location ID (negative)",
			CreateRecordRequest{Title: "Test", HomeLocationID: -1},
			true,
		},
		{
			"valid location IDs",
			CreateRecordRequest{Title: "Test", CurrentLocationID: 1, HomeLocationID: 2},
			false,
		},
		{
			"catalog number too long (101 chars)",
			CreateRecordRequest{Title: "Test", CatalogNumber: string(make([]byte, 101))},
			true,
		},
		{
			"catalog number maximum length (100 chars)",
			CreateRecordRequest{Title: "Test", CatalogNumber: string(make([]byte, 100))},
			false,
		},
		{
			"valid condition: Mint",
			CreateRecordRequest{Title: "Test", Condition: "Mint"},
			false,
		},
		{
			"valid condition: Near Mint",
			CreateRecordRequest{Title: "Test", Condition: "Near Mint"},
			false,
		},
		{
			"valid condition: Very Good",
			CreateRecordRequest{Title: "Test", Condition: "Very Good"},
			false,
		},
		{
			"valid condition: Good",
			CreateRecordRequest{Title: "Test", Condition: "Good"},
			false,
		},
		{
			"valid condition: Fair",
			CreateRecordRequest{Title: "Test", Condition: "Fair"},
			false,
		},
		{
			"valid condition: Poor",
			CreateRecordRequest{Title: "Test", Condition: "Poor"},
			false,
		},
		{
			"invalid condition",
			CreateRecordRequest{Title: "Test", Condition: "Excellent"},
			true,
		},
		{
			"notes too long (1001 chars)",
			CreateRecordRequest{Title: "Test", Notes: string(make([]byte, 1001))},
			true,
		},
		{
			"notes maximum length (1000 chars)",
			CreateRecordRequest{Title: "Test", Notes: string(make([]byte, 1000))},
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

// TestUpdateRecordRequest_Validation tests update record validation rules
func TestUpdateRecordRequest_Validation(t *testing.T) {
	validate := validator.New()

	type UpdateRecordRequest struct {
		Title             string `validate:"required,min=1,max=200"`
		ArtistID          int64  `validate:"omitempty,min=1"`
		AlbumTitle        string `validate:"max=200"`
		ReleaseYear       int32  `validate:"omitempty,min=1900,max=2100"`
		CurrentLocationID int64  `validate:"omitempty,min=1"`
		HomeLocationID    int64  `validate:"omitempty,min=1"`
		CatalogNumber     string `validate:"max=100"`
		Condition         string `validate:"omitempty,oneof=Mint 'Near Mint' 'Very Good' Good Fair Poor"`
		Notes             string `validate:"max=1000"`
	}

	tests := []struct {
		name      string
		request   UpdateRecordRequest
		wantError bool
	}{
		{
			"valid update with all fields",
			UpdateRecordRequest{
				Title:             "Updated Title",
				ArtistID:          2,
				AlbumTitle:        "Updated Album",
				ReleaseYear:       2020,
				CurrentLocationID: 3,
				HomeLocationID:    4,
				CatalogNumber:     "NEW-123",
				Condition:         "Good",
				Notes:             "Updated notes",
			},
			false,
		},
		{
			"valid update with minimal fields",
			UpdateRecordRequest{Title: "Updated Title"},
			false,
		},
		{
			"empty title",
			UpdateRecordRequest{Title: ""},
			true,
		},
		{
			"clear optional fields",
			UpdateRecordRequest{
				Title:         "Valid Title",
				ArtistID:      0,
				AlbumTitle:    "",
				CatalogNumber: "",
				Notes:         "",
			},
			true, // ArtistID=0 fails validation
		},
		{
			"update condition to Fair",
			UpdateRecordRequest{Title: "Test", Condition: "Fair"},
			false,
		},
		{
			"update condition to Poor",
			UpdateRecordRequest{Title: "Test", Condition: "Poor"},
			false,
		},
		{
			"invalid condition",
			UpdateRecordRequest{Title: "Test", Condition: "Damaged"},
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

// TestUpdateRecordLocationRequest_Validation tests location update validation
func TestUpdateRecordLocationRequest_Validation(t *testing.T) {
	validate := validator.New()

	type UpdateRecordLocationRequest struct {
		CurrentLocationID int64 `validate:"required,min=1"`
	}

	tests := []struct {
		name      string
		request   UpdateRecordLocationRequest
		wantError bool
	}{
		{"valid location ID", UpdateRecordLocationRequest{CurrentLocationID: 1}, false},
		{"large location ID", UpdateRecordLocationRequest{CurrentLocationID: 999999}, false},
		{"zero location ID", UpdateRecordLocationRequest{CurrentLocationID: 0}, true},
		{"negative location ID", UpdateRecordLocationRequest{CurrentLocationID: -1}, true},
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

// TestUpdateRecordConditionRequest_Validation tests condition update validation
func TestUpdateRecordConditionRequest_Validation(t *testing.T) {
	validate := validator.New()

	type UpdateRecordConditionRequest struct {
		Condition string `validate:"required,oneof=Mint 'Near Mint' 'Very Good' Good Fair Poor"`
	}

	tests := []struct {
		name      string
		request   UpdateRecordConditionRequest
		wantError bool
	}{
		{"Mint", UpdateRecordConditionRequest{Condition: "Mint"}, false},
		{"Near Mint", UpdateRecordConditionRequest{Condition: "Near Mint"}, false},
		{"Very Good", UpdateRecordConditionRequest{Condition: "Very Good"}, false},
		{"Good", UpdateRecordConditionRequest{Condition: "Good"}, false},
		{"Fair", UpdateRecordConditionRequest{Condition: "Fair"}, false},
		{"Poor", UpdateRecordConditionRequest{Condition: "Poor"}, false},
		{"empty condition", UpdateRecordConditionRequest{Condition: ""}, true},
		{"invalid condition", UpdateRecordConditionRequest{Condition: "Excellent"}, true},
		{"lowercase", UpdateRecordConditionRequest{Condition: "mint"}, true}, // case-sensitive
		{"wrong casing", UpdateRecordConditionRequest{Condition: "MINT"}, true},
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