package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/dukerupert/dd/internal/store"
)

type contextKey string

const (
	RequestIDKey contextKey = "requestID"
	UserIDKey    contextKey = "userID"
	StartTimeKey contextKey = "startTime"
)

// RequestID adds a unique request ID to context
func RequestID(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request ID already exists (from load balancer/proxy)
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			// Generate random ID - use crypto/rand for better entropy
			bytes := make([]byte, 16) // 16 bytes = 32 hex chars (better uniqueness)
			if _, err := rand.Read(bytes); err != nil {
				// Fallback to timestamp-based ID if random fails
				requestID = fmt.Sprintf("%d", time.Now().UnixNano())
			} else {
				requestID = hex.EncodeToString(bytes)
			}
		}

		// Add to context
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		r = r.WithContext(ctx)

		// Add to response headers for debugging
		w.Header().Set("X-Request-ID", requestID)

		h.ServeHTTP(w, r)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += int64(n)
	return n, err
}

// Logging adds structured logging with request context
func Logging(h http.Handler, l *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		requestID := GetRequestID(r.Context())
		userID, _ := GetUserID(r.Context())

		// Wrap response writer to capture status code
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // default
		}

		// Create contextualized logger
		logger := l.With(
			slog.String("request_id", requestID),
			slog.String("user_id", userID),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("user_agent", r.UserAgent()),
		)

		// Log incoming request at debug level
		logger.Debug("Incoming request")

		// Recover from panics
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered",
					slog.Any("error", err),
					slog.String("stack", string(debug.Stack())),
				)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		h.ServeHTTP(rw, r)

		// Log response with additional context
		elapsed := time.Since(start)
		logLevel := slog.LevelInfo

		// Use different log levels based on status code
		if rw.statusCode >= 500 {
			logLevel = slog.LevelError
		} else if rw.statusCode >= 400 {
			logLevel = slog.LevelWarn
		}

		logger.Log(r.Context(), logLevel, "Request completed",
			slog.Int("status", rw.statusCode),
			slog.Duration("duration", elapsed),
			slog.Int64("bytes", rw.written),
		)
	})
}

// MaxBytes limits request body size
func MaxBytes(maxBytes int64) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			h.ServeHTTP(w, r)
		})
	}
}

// Auth extracts and validates user authentication
// This is a permissive middleware - it extracts user info but doesn't require auth
func Auth(queries *store.Queries, sessionCookieName string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			userID := "anonymous"

			// Try to extract session from cookie
			cookie, err := r.Cookie(sessionCookieName)
			if err == nil && cookie.Value != "" {
				// Validate session token from database
				session, err := queries.GetSessionByToken(ctx, cookie.Value)
				if err == nil && session.ExpiresAt.After(time.Now()) {
					userID = session.UserID

					// Optional: Refresh session expiry on activity
					// queries.UpdateSessionExpiry(ctx, cookie.Value, time.Now().Add(24*time.Hour))
				}
			}

			// Fallback: Check Authorization header for API tokens
			if userID == "anonymous" {
				authHeader := r.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					token := strings.TrimPrefix(authHeader, "Bearer ")

					// Validate API token
					apiToken, err := queries.GetAPITokenByToken(ctx, token)
					if err == nil && apiToken.ExpiresAt.After(time.Now()) {
						userID = apiToken.UserID
					}
				}
			}

			// Add user to context
			ctx = context.WithValue(ctx, UserIDKey, userID)
			r = r.WithContext(ctx)

			h.ServeHTTP(w, r)
		})
	}
}

// RequireAuth enforces authentication - returns 401 if not authenticated
func RequireAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())

		if !ok || userID == "anonymous" || userID == "" {
			// Check if HTMX request
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/login")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Regular request - redirect to login
			http.Redirect(w, r, "/login?redirect="+r.URL.Path, http.StatusSeeOther)
			return
		}

		h.ServeHTTP(w, r)
	})
}

// RequireAPIAuth enforces authentication for API routes - returns JSON error
func RequireAPIAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())

		if !ok || userID == "anonymous" || userID == "" {
			writeErrorJSON(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		h.ServeHTTP(w, r)
	})
}

// RequireRole checks if user has required role
func RequireRole(queries *store.Queries, role string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			userID, ok := GetUserID(r.Context())

			if !ok || userID == "anonymous" || userID == "" {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Check user role from database
			user, err := queries.GetUserByID(ctx, userID)
			if err != nil || user.Role != role {
				http.Error(w, "Forbidden - insufficient permissions", http.StatusForbidden)
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}

// CSRF validates CSRF tokens for state-changing requests
func CSRF(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF for safe methods
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
			h.ServeHTTP(w, r)
			return
		}

		// Get CSRF token from header (HTMX) or form
		csrfToken := r.Header.Get("X-CSRF-Token")
		if csrfToken == "" {
			csrfToken = r.FormValue("csrf_token")
		}

		// Get expected token from cookie
		cookie, err := r.Cookie("csrf_token")
		if err != nil || cookie.Value == "" || cookie.Value != csrfToken {
			if r.Header.Get("HX-Request") == "true" {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("Invalid CSRF token"))
				return
			}
			http.Error(w, "Invalid CSRF token", http.StatusForbidden)
			return
		}

		h.ServeHTTP(w, r)
	})
}

// Helper functions with better error handling
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return "unknown"
}

func GetUserID(ctx context.Context) (string, bool) {
	if id, ok := ctx.Value(UserIDKey).(string); ok && id != "" {
		return id, true
	}
	return "", false
}

func IsAuthenticated(ctx context.Context) bool {
	userID, ok := GetUserID(ctx)
	return ok && userID != "" && userID != "anonymous"
}

// RateLimiter tracks request counts per IP address
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     int           // requests per window
	window   time.Duration // time window
}

// visitor tracks requests for a single IP
type visitor struct {
	lastSeen time.Time
	count    int
	resetAt  time.Time
}

// NewRateLimiter creates a new rate limiter
// rate: number of requests allowed per window
// window: time window (e.g., time.Minute)
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
	}

	// Cleanup stale visitors every window duration
	go rl.cleanupVisitors()

	return rl
}

// Allow checks if a request from this IP should be allowed
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	v, exists := rl.visitors[ip]
	if !exists {
		// First request from this IP
		rl.visitors[ip] = &visitor{
			lastSeen: now,
			count:    1,
			resetAt:  now.Add(rl.window),
		}
		return true
	}

	// Reset count if window has passed
	if now.After(v.resetAt) {
		v.count = 1
		v.resetAt = now.Add(rl.window)
		v.lastSeen = now
		return true
	}

	// Check if under rate limit
	if v.count < rl.rate {
		v.count++
		v.lastSeen = now
		return true
	}

	// Rate limit exceeded
	v.lastSeen = now
	return false
}

// cleanupVisitors removes stale entries to prevent memory leak
func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, v := range rl.visitors {
			// Remove visitors not seen for 3x the window duration
			if now.Sub(v.lastSeen) > rl.window*3 {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimit returns a middleware that rate limits requests
func RateLimit(next http.Handler, requestsPerMinute int) http.Handler {
	limiter := NewRateLimiter(requestsPerMinute, time.Minute)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)

		// Add rate limit headers
		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", requestsPerMinute))

		if !limiter.Allow(ip) {
			w.Header().Set("Retry-After", "60")
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getIP extracts the real IP address from the request
func getIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies/load balancers)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		return forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// writeErrorJSON writes a JSON error response
func writeErrorJSON(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"error":   http.StatusText(statusCode),
		"message": message,
		"code":    statusCode,
	}

	json.NewEncoder(w).Encode(response)
}
