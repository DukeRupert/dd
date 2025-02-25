package handlers

import (
	"fmt"
	"net/http"

	"github.com/dukerupert/dd/internal/models"

	pocketbase "github.com/dukerupert/dd/pb"
	"github.com/labstack/echo/v4"
)

// Use the LoginRequest from models package
type LoginRequest = models.LoginRequest

type AuthHandler struct {
	pbClient pocketbase.IClient
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
		// For API requests, maintain proper status code
		if c.Request().Header.Get("Accept") == "application/json" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid request format",
			})
		}
		return h.LoginError(c, "Invalid request format")
	}

	// Debug statements
	fmt.Printf("Auth request - Email: %s, Password length: %d\n", req.Email, len(req.Password))
	fmt.Printf("Attempting authentication with PocketBase...\n")

	// Log the request headers to see what's being sent
	fmt.Println("Request headers:")
	for k, v := range c.Request().Header {
		fmt.Printf("  %s: %v\n", k, v)
	}

	// Authenticate with Pocketbase
	authResp, err := h.pbClient.AuthWithPassword(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		// For API requests, maintain proper status code
		if c.Request().Header.Get("Accept") == "application/json" {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Invalid email or password",
			})
		}
		return h.LoginError(c, "Invalid email or password")
	}

	// Use the token directly from the typed response
	h.pbClient.SetAuthToken(authResp.Token)

	// Return token and user info in case of API requests
	if c.Request().Header.Get("Accept") == "application/json" {
		return c.JSON(http.StatusOK, models.LoginResponse{
			Token:  authResp.Token,
			Record: authResp.Record,
		})
	}

	// Return success HTML for HTMX
	return h.LoginSuccess(c)
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

// LoginSuccess renders a success message or redirects after successful login
func (h *AuthHandler) LoginSuccess(c echo.Context) error {
	// In a real app, you might redirect to a dashboard
	// For this example, we'll just render a success message
	return c.HTML(http.StatusOK, `
		<div id="login-form" class="space-y-6">
			<div class="bg-green-50 p-4 rounded-md">
				<div class="flex">
					<div class="flex-shrink-0">
						<svg class="h-5 w-5 text-green-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
							<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
						</svg>
					</div>
					<div class="ml-3">
						<h3 class="text-sm font-medium text-green-800">Login successful</h3>
						<div class="mt-2 text-sm text-green-700">
							<p>You are now logged in. Redirecting...</p>
						</div>
					</div>
				</div>
			</div>
			<script>
				setTimeout(function() {
					window.location.href = '/dashboard';
				}, 2000);
			</script>
		</div>
	`)
}

// LoginError renders an error message when login fails
func (h *AuthHandler) LoginError(c echo.Context, errorMsg string) error {
	return c.HTML(http.StatusUnauthorized, fmt.Sprintf(`
		<form class="space-y-6" hx-post="/auth/login" hx-trigger="submit" hx-swap="outerHTML" hx-target="#login-form" id="login-form">
			<div class="bg-red-50 p-4 rounded-md">
				<div class="flex">
					<div class="flex-shrink-0">
						<svg class="h-5 w-5 text-red-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
							<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
						</svg>
					</div>
					<div class="ml-3">
						<h3 class="text-sm font-medium text-red-800">Login failed</h3>
						<div class="mt-2 text-sm text-red-700">
							<p>%s</p>
						</div>
					</div>
				</div>
			</div>
			
			<div>
				<label for="email" class="block text-sm/6 font-medium text-gray-900">Email address</label>
				<div class="mt-2">
					<input type="email" name="email" id="email" autocomplete="email" required class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline-2 focus:-outline-offset-2 focus:outline-indigo-600 sm:text-sm/6" />
				</div>
			</div>

			<div>
				<label for="password" class="block text-sm/6 font-medium text-gray-900">Password</label>
				<div class="mt-2">
					<input type="password" name="password" id="password" autocomplete="current-password" required class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline-2 focus:-outline-offset-2 focus:outline-indigo-600 sm:text-sm/6" />
				</div>
			</div>

			<div class="flex items-center justify-between">
				<div class="flex gap-3">
					<div class="flex h-6 shrink-0 items-center">
						<div class="group grid size-4 grid-cols-1">
							<input id="remember-me" name="remember-me" type="checkbox" class="col-start-1 row-start-1 appearance-none rounded-sm border border-gray-300 bg-white checked:border-indigo-600 checked:bg-indigo-600 indeterminate:border-indigo-600 indeterminate:bg-indigo-600 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 disabled:border-gray-300 disabled:bg-gray-100 disabled:checked:bg-gray-100 forced-colors:appearance-auto" />
							<svg class="pointer-events-none col-start-1 row-start-1 size-3.5 self-center justify-self-center stroke-white group-has-disabled:stroke-gray-950/25" viewBox="0 0 14 14" fill="none">
								<path class="opacity-0 group-has-checked:opacity-100" d="M3 8L6 11L11 3.5" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
								<path class="opacity-0 group-has-indeterminate:opacity-100" d="M3 7H11" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
							</svg>
						</div>
					</div>
					<label for="remember-me" class="block text-sm/6 text-gray-900">Remember me</label>
				</div>

				<div class="text-sm/6">
					<a href="#" class="font-semibold text-indigo-600 hover:text-indigo-500">Forgot password?</a>
				</div>
			</div>

			<div>
				<button type="submit" class="flex w-full justify-center rounded-md bg-indigo-600 px-3 py-1.5 text-sm/6 font-semibold text-white shadow-xs hover:bg-indigo-500 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600">
					<span class="htmx-indicator inline-block mr-2">
						<svg class="animate-spin h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
							<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
							<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
						</svg>
					</span>
					Sign in
				</button>
			</div>
		</form>
	`, errorMsg))
}
