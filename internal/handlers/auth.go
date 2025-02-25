package handlers

import (
	"net/http"
	
	"github.com/labstack/echo/v4"
	"github.com/dukerupert/dd/pb"
)

type AuthHandler struct {
	pbClient pocketbase.IClient
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token  string                 `json:"token"`
	User   pocketbase.RecordDetail `json:"user"`
}

func NewAuthHandler(pbClient pocketbase.IClient) *AuthHandler {
	return &AuthHandler{
		pbClient: pbClient,
	}
}

// Login handles user authentication
func (h *AuthHandler) Login(c echo.Context) error {
    var req LoginRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "Invalid request format",
        })
    }
    
    // Validate required fields
    if req.Email == "" || req.Password == "" {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "Email and password are required",
        })
    }
    
    // Authenticate with Pocketbase
    authResp, err := h.pbClient.AuthWithPassword(c.Request().Context(), req.Email, req.Password)
    if err != nil {
        return c.JSON(http.StatusUnauthorized, map[string]string{
            "error": "Authentication failed",
        })
    }
    
    // Important: This line is probably missing in your implementation
    // The test expects SetAuthToken to be called
    h.pbClient.SetAuthToken(authResp.Token)
    
    // Return token and user info
    return c.JSON(http.StatusOK, LoginResponse{
        Token: authResp.Token,
        User:  authResp.Record,
    })
}

// Middleware to check if user is authenticated
func (h *AuthHandler) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Request().Header.Get("Authorization")
		if token == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Missing authorization token",
			})
		}
		
		// In a real app, you might want to validate the JWT token here
		// For simplicity, we're just checking if it exists
		
		// Set the token in the context for later use
		c.Set("auth_token", token)
		
		return next(c)
	}
}

// Logout handles user logout
func (h *AuthHandler) Logout(c echo.Context) error {
	// Clear the token on the server side
	h.pbClient.ClearAuth()
	
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}

// CheckAuth validates if the user is authenticated
func (h *AuthHandler) CheckAuth(c echo.Context) error {
	token := c.Request().Header.Get("Authorization")
	if token == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Missing authorization token",
		})
	}
	
	// Set token and check if it's valid
	h.pbClient.SetAuthToken(token)
	if !h.pbClient.IsAuthenticated() {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid or expired token",
		})
	}
	
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Token is valid",
	})
}