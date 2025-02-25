package models

import "time"

// LoginRequest represents the login form data
type LoginRequest struct {
	Email    string `json:"email" form:"email"`
	Password string `json:"password" form:"password"`
	Remember bool   `json:"remember_me" form:"remember-me"`
}

// UserRecord represents a user record from PocketBase
type UserRecord struct {
	CollectionId   string    `json:"collectionId"`
	CollectionName string    `json:"collectionName"`
	Id             string    `json:"id"`
	Email          string    `json:"email"`
	EmailVisibility bool      `json:"emailVisibility"`
	Verified       bool      `json:"verified"`
	Name           string    `json:"name"`
	Avatar         string    `json:"avatar"`
	Created        time.Time `json:"created"`
	Updated        time.Time `json:"updated"`
}

// LoginResponse represents the response sent to the client after login
type LoginResponse struct {
	Token  string     `json:"token"`
	Record UserRecord `json:"record"`
}

// AuthResponse represents the response from PocketBase authentication
type AuthResponse struct {
	Token  string     `json:"token"`
	Record UserRecord `json:"record"`
}