# Migration Guide: File Structure Reorganization

This guide explains how to complete the migration to the new file structure.

## Current Status

### Completed ✅
1. Created new directory structure (`internal/*`)
2. Created `internal/config` package with environment variable support
3. Created `internal/auth` package (jwt, session, password utilities)
4. Created `internal/renderer` package (template rendering)
5. Created `internal/middleware` package (all middleware consolidated)
6. Created `internal/handler` package base (Handler struct + helpers)

### Remaining Work 🚧

## Step 1: Migrate Handler Files (Manual)

Each handler file needs to be migrated to use the new Handler struct pattern.

### Old Pattern (artist-handlers.go)
```go
func handleGetArtist(logger *slog.Logger, queries *store.Queries, renderer *TemplateRenderer) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // handler logic
    })
}
```

### New Pattern (internal/handler/artist.go)
```go
// GetArtist handles GET /artists/{id}
func (h *Handler) GetArtist() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // handler logic using h.logger, h.queries, h.renderer
    }
}
```

### Migration Checklist for Each Handler File:

- [ ] `artist-handlers.go` → `internal/handler/artist.go`
  - Convert `handleGetArtistsPage` → `(h *Handler) GetArtists Page()`
  - Convert `handleGetArtist` → `(h *Handler) GetArtist()`
  - Convert `handlePostArtist` → `(h *Handler) PostArtist()`
  - Convert `handlePutArtist` → `(h *Handler) PutArtist()`
  - Convert `handleDeleteArtist` → `(h *Handler) DeleteArtist()`
  - Convert `handleGetArtistNewForm` → `(h *Handler) GetArtistNewForm()`
  - Convert `handleGetArtistEditForm` → `(h *Handler) GetArtistEditForm()`
  - Convert API handlers similarly

- [ ] `record-handlers.go` → `internal/handler/record.go`
  - Same pattern for all record handlers

- [ ] `location-handlers.go` → `internal/handler/location.go`
  - Same pattern for all location handlers

- [ ] `auth-handlers.go` → `internal/handler/auth.go`
  - Use `h.config` for JWT secret and session settings
  - Replace `generateJWT` with `auth.GenerateJWT()`
  - Replace `createSession` with `auth.CreateSession()`
  - Replace `setSessionCookie` with `auth.SetSessionCookie()`
  - Replace `bcrypt.GenerateFromPassword` with `auth.HashPassword()`
  - Replace `bcrypt.CompareHashAndPassword` with `auth.ComparePassword()`

- [ ] `user-handlers.go` → `internal/handler/user.go`
  - Same pattern

- [ ] `stat-handlers.go` → `internal/handler/stats.go`
  - Same pattern

### Helper Function Replacements

Replace global helper calls with Handler methods:

| Old Function | New Method |
|--------------|------------|
| `bind(r, &req)` | `h.bind(r, &req)` |
| `writeJSON(w, data, status)` | `h.writeJSON(w, data, status)` |
| `writeErrorJSON(w, msg, status)` | `h.writeErrorJSON(w, msg, status)` |
| `formatValidationErrorsHTML(errs)` | `h.formatValidationErrorsHTML(errs)` |
| `getValidationErrors(err)` | `h.getValidationErrors(err)` |

## Step 2: Create Router Package

Create `internal/router/router.go`:

```go
package router

import (
	"net/http"

	"github.com/dukerupert/dd/internal/handler"
	"github.com/dukerupert/dd/internal/middleware"
	"github.com/dukerupert/dd/internal/store"
)

// New creates and configures the application router
func New(h *handler.Handler, queries *store.Queries, sessionCookieName string) http.Handler {
	mux := http.NewServeMux()

	// API routes
	apiMux := http.NewServeMux()
	addAPIRoutes(apiMux, h)

	apiHandler := http.StripPrefix("/api", apiMux)
	apiHandler = middleware.RateLimit(apiHandler, 100)
	apiHandler = middleware.MaxBytes(1 << 20)(apiHandler)
	apiHandler = middleware.Auth(queries, sessionCookieName)(apiHandler)
	apiHandler = middleware.Logging(apiHandler, h.Logger())
	apiHandler = middleware.RequestID(apiHandler)
	mux.Handle("/api/", apiHandler)

	// HTML routes
	htmlMux := http.NewServeMux()
	addHTMLRoutes(htmlMux, h)

	htmlHandler := http.Handler(htmlMux)
	htmlHandler = middleware.RateLimit(htmlHandler, 1000)
	htmlHandler = middleware.MaxBytes(10 << 20)(htmlHandler)
	htmlHandler = middleware.Auth(queries, sessionCookieName)(htmlHandler)
	htmlHandler = middleware.Logging(htmlHandler, h.Logger())
	htmlHandler = middleware.RequestID(htmlHandler)
	mux.Handle("/", htmlHandler)

	return mux
}

func addHTMLRoutes(mux *http.ServeMux, h *handler.Handler) {
	// Public routes
	mux.HandleFunc("GET /", h.Landing())
	mux.HandleFunc("GET /signup", h.SignupPage())
	mux.HandleFunc("POST /signup", h.Signup())
	mux.HandleFunc("GET /login", h.LoginPage())
	mux.HandleFunc("POST /login", h.Login())
	mux.HandleFunc("POST /logout", h.Logout())

	// Protected routes
	mux.HandleFunc("GET /dashboard", h.Dashboard())

	// Artists
	mux.HandleFunc("GET /artists", h.GetArtistsPage())
	mux.HandleFunc("GET /artists/new", h.GetArtistNewForm())
	mux.HandleFunc("POST /artists", h.PostArtist())
	mux.HandleFunc("GET /artists/{id}", h.GetArtist())
	mux.HandleFunc("GET /artists/{id}/edit", h.GetArtistEditForm())
	mux.HandleFunc("PUT /artists/{id}", h.PutArtist())
	mux.HandleFunc("DELETE /artists/{id}", h.DeleteArtist())

	// Records
	mux.HandleFunc("GET /records", h.GetRecordsPage())
	// ... etc
}

func addAPIRoutes(mux *http.ServeMux, h *handler.Handler) {
	// Public API
	mux.HandleFunc("POST /v1/auth/signup", h.APISignup())
	mux.HandleFunc("POST /v1/auth/login", h.APILogin())

	// Protected API
	mux.HandleFunc("GET /v1/artists", h.APIGetArtists())
	// ... etc
}
```

## Step 3: Update cmd/server/main.go

Create `cmd/server/main.go`:

```go
package main

import (
	"context"
	"database/sql"
	"embed"
	"log"
	"log/slog"
	"net"
	"net/http"
	"strconv"

	"github.com/dukerupert/dd/data/sql/migrations"
	"github.com/dukerupert/dd/internal/config"
	"github.com/dukerupert/dd/internal/handler"
	"github.com/dukerupert/dd/internal/renderer"
	"github.com/dukerupert/dd/internal/router"
	"github.com/dukerupert/dd/internal/store"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
	_ "modernc.org/sqlite"
)

//go:embed ../../templates/*
var templateFS embed.FS

func run() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Setup logger
	logger := slog.New(cfg.Logging.Handler)
	slog.SetDefault(logger)

	// Open database
	db, err := sql.Open("sqlite", cfg.Database.Path)
	if err != nil {
		return err
	}
	defer db.Close()

	// Run migrations
	provider, err := goose.NewProvider(database.DialectSQLite3, db, migrations.Embed)
	if err != nil {
		return err
	}

	ctx := context.Background()
	if _, err := provider.Up(ctx); err != nil {
		return err
	}

	// Create queries
	queries := store.New(db)

	// Create renderer
	r := renderer.New(templateFS)
	if err := r.LoadTemplates(); err != nil {
		return err
	}

	// Create handler
	h := handler.New(logger, queries, r, cfg)

	// Create router
	srv := router.New(h, queries, cfg.Session.CookieName)

	logger.Info("Starting server", slog.String("port", strconv.Itoa(cfg.Server.Port)))

	return http.ListenAndServe(net.JoinHostPort(cfg.Server.Host, strconv.Itoa(cfg.Server.Port)), srv)
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
```

## Step 4: Create .env.example

Create `.env.example`:

```bash
# Server Configuration
SERVER_HOST=localhost
SERVER_PORT=8080
ENVIRONMENT=dev

# Database
DATABASE_PATH=sqlite.db

# Logging
LOG_LEVEL=info

# Authentication
JWT_SECRET=your-secret-key-change-in-production

# Session (handled by config defaults)
```

## Step 5: Update Tests

Update test files to use new imports:

```go
package handler_test

import (
	"testing"

	"github.com/dukerupert/dd/internal/config"
	"github.com/dukerupert/dd/internal/handler"
	"github.com/dukerupert/dd/internal/renderer"
)

// Test helper
func setupTestHandler(t *testing.T) *handler.Handler {
	db, queries := setupTestDB(t)
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	r := renderer.New(templateFS)
	r.LoadTemplates()
	cfg := &config.Config{} // Test config

	return handler.New(logger, queries, r, cfg)
}
```

## Step 6: Update CLAUDE.md

Update documentation to reflect new structure:

```markdown
## Commands

### Build and Run
```bash
# Build
go build -o bin/server ./cmd/server

# Run with default config
./bin/server

# Run with environment variables
SERVER_PORT=3000 LOG_LEVEL=debug ./bin/server
```

## Architecture

### Project Structure
```
dd/
├── cmd/server/          # Application entry point
├── internal/
│   ├── config/         # Configuration management
│   ├── auth/           # Authentication (JWT, sessions, passwords)
│   ├── handler/        # HTTP handlers
│   ├── middleware/     # HTTP middleware
│   ├── renderer/       # Template rendering
│   ├── router/         # Route registration
│   └── store/          # Database layer (sqlc generated)
├── data/sql/           # SQL migrations and queries
├── templates/          # HTML templates
└── static/             # Static assets
```
```

## Step 7: Cleanup

After everything is working:

1. Delete old files from root:
   ```bash
   rm artist-handlers.go auth-handlers.go record-handlers.go location-handlers.go \\
      user-handlers.go stat-handlers.go http.go middleware.go routes.go \\
      renderer.go session.go jwt.go main.go
   ```

2. Keep these files:
   - `go.mod`, `go.sum`
   - `sqlc.yaml`
   - `data/`, `templates/`, `static/`, `internal/store/`
   - Documentation: `CLAUDE.md`, `ROADMAP.md`, `MIGRATION.md`

## Quick Start (Minimal Migration)

If you want to test the new structure quickly:

1. Copy one handler file (e.g., `artist-handlers.go`) to `internal/handler/artist.go`
2. Convert function signatures to methods
3. Create a minimal `cmd/server/main.go`
4. Test with: `go run cmd/server/main.go`
5. Once working, migrate remaining handlers

## Benefits of New Structure

- **Better organization**: Related code grouped together
- **Easier testing**: Dependencies injected via Handler struct
- **Cleaner imports**: No circular dependencies
- **Configuration**: Environment variables + command-line flags
- **Scalability**: Easy to add new domains/features
- **Professional**: Follows Go best practices (Standard Project Layout)
