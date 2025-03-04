// pocketbase/models.go
package pocketbase

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"time"
)

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

type User struct {
	AuthProviders    []interface{} `json:"authProviders"`
	UsernamePassword bool          `json:"usernamePassword"`
	EmailPassword    bool          `json:"emailPassword"`
	OnlyVerified     bool          `json:"onlyVerified"`
}

// BaseModel contains fields common to all PocketBase records
type BaseModel struct {
	ID             string         `json:"id"`
	CollectionName string         `json:"collectionName"`
	CollectionID   string         `json:"collectionId"`
	Created        PocketBaseTime `json:"created"`
	Updated        PocketBaseTime `json:"updated"`
}

// ListResult represents a paginated list of records
type ListResult[T any] struct {
	Page       int `json:"page"`
	PerPage    int `json:"perPage"`
	TotalItems int `json:"totalItems"`
	TotalPages int `json:"totalPages"`
	Items      []T `json:"items"`
}

// QueryParams represents query parameters for list endpoints
type QueryParams struct {
	Page      int
	PerPage   int
	Sort      string
	Filter    string
	Expand    string
	Fields    string
	SkipTotal bool
}

// ToURLValues converts QueryParams to URL query values
func (q QueryParams) ToURLValues() url.Values {
	values := url.Values{}

	if q.Page > 0 {
		values.Set("page", strconv.Itoa(q.Page))
	}

	if q.PerPage > 0 {
		values.Set("perPage", strconv.Itoa(q.PerPage))
	}

	if q.Sort != "" {
		values.Set("sort", q.Sort)
	}

	if q.Filter != "" {
		values.Set("filter", q.Filter)
	}

	if q.Expand != "" {
		values.Set("expand", q.Expand)
	}

	if q.Fields != "" {
		values.Set("fields", q.Fields)
	}

	if q.SkipTotal {
		values.Set("skipTotal", "true")
	}

	return values
}

// PocketBaseTime is a custom time type to handle PocketBase's time format
type PocketBaseTime time.Time

// UnmarshalJSON implements json.Unmarshaler
func (pbt *PocketBaseTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "null" || s == "" {
		*pbt = PocketBaseTime(time.Time{})
		return nil
	}

	// PocketBase format: "2025-02-24 03:27:41.065Z"
	t, err := time.Parse("2006-01-02 15:04:05.999Z", s)
	if err != nil {
		// Try alternative formats if the first one fails
		t, err = time.Parse("2006-01-02 15:04:05.999", s)
		if err != nil {
			return err
		}
	}

	*pbt = PocketBaseTime(t)
	return nil
}

// MarshalJSON implements json.Marshaler
func (pbt PocketBaseTime) MarshalJSON() ([]byte, error) {
	t := time.Time(pbt)
	if t.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(t.Format("2006-01-02 15:04:05.999Z"))
}

// String returns the time as a formatted string
func (pbt PocketBaseTime) String() string {
	t := time.Time(pbt)
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05.999Z")
}

// Time returns the underlying time.Time
func (pbt PocketBaseTime) Time() time.Time {
	return time.Time(pbt)
}

type Album struct {
	BaseModel
	Title         string         `json:"title"`
	ArtistID      string         `json:"artist_id"`
	LocationID    string         `json:"location_id"`
	ReleaseYear   int            `json:"release_year"`
	Genre         string         `json:"genre"`
	Condition     string         `json:"condition"`
	PurchaseDate  PocketBaseTime `json:"purchase_date,omitempty"`
	PurchasePrice float64        `json:"purchase_price,omitempty"`
	Notes         string         `json:"notes,omitempty"`
	// Add expand field for relations if needed
	Expand struct {
		Artist   *Artist   `json:"artist_id,omitempty"`
		Location *Location `json:"location_id,omitempty"`
	} `json:"expand,omitempty"`
}

// Artist represents an artist record
type Artist struct {
	BaseModel
	Name string `json:"name"`
	// Add other fields as needed
}

// Location represents a location record
type Location struct {
	BaseModel
	Name string `json:"name"`
	// Add other fields as needed
}
