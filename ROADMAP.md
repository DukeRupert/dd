# Development Roadmap

This document outlines the remaining tasks to complete the vinyl record collection application.

## Status Legend
- ✅ Complete
- 🚧 In Progress
- ⏳ Not Started

---

## 1. Authentication & Security

### 1.1 Core Authentication ✅
- ✅ User signup (HTML + API)
- ✅ User login (HTML + API)
- ✅ Session management with cookies
- ✅ JWT token generation
- ✅ Password hashing with bcrypt

### 1.2 Security Enhancements ⏳
- ⏳ Move JWT secret to environment variable (currently hardcoded in `jwt.go:12`)
- ⏳ Enable `Secure` flag for session cookies in production (currently `false` in `session.go:69`)
- ⏳ Enable API authentication middleware (currently disabled in `routes.go:142`)
- ⏳ Implement CSRF protection (middleware exists but not enabled)
- ⏳ Add rate limiting configuration per environment
- ⏳ Implement password reset functionality (route exists at `/forgot-password`)

### 1.3 Dashboard ⏳
- ⏳ Implement `handleDashboard` - show user's collection stats
  - Total records count
  - Recently added records
  - Recently played records
  - Collection by artist/location breakdown

---

## 2. Artist Management

### 2.1 HTML Handlers
- ✅ List all artists (`handleGetArtistsPage`)
- ✅ View artist detail page with records (`handleGetArtist`)
- ✅ Create new artist (`handleGetArtistNewForm`, `handlePostArtist`)
- 🚧 Edit artist (`handleGetArtistEditForm`, `handlePutArtist`)
  - Template exists but handlers are stubs
- 🚧 Delete artist (`handleDeleteArtist`)
  - Consider cascade behavior for records

### 2.2 API Handlers
- ✅ GET `/api/v1/artists` - list artists
- ✅ POST `/api/v1/artists` - create artist
- ✅ GET `/api/v1/artists/{id}` - get single artist
- 🚧 PUT `/api/v1/artists/{id}` - update artist
- 🚧 DELETE `/api/v1/artists/{id}` - delete artist
- ⏳ GET `/api/v1/artists/{id}/records` - get artist's records

### 2.3 Features to Add
- ⏳ Search artists by name (query already exists: `SearchArtistsByName`)
- ⏳ Pagination for artist list (query exists: `ListArtistsWithPagination`)
- ⏳ Sort options (alphabetical, by record count, etc.)

---

## 3. Record Management

### 3.1 HTML Handlers
- ✅ List all records (`handleGetRecordsPage`)
- ⏳ View record detail page (`handleGetRecord`)
- ⏳ Create new record (`handleGetRecordNewForm`, `handlePostRecord`)
  - Need to load artists and locations for dropdowns
  - Validate artist_id and location_id references
- ⏳ Edit record (`handleGetRecordEditForm`, `handlePutRecord`)
- ⏳ Delete record (`handleDeleteRecord`)
- ⏳ Track playback (`handlePostRecordPlay`)
  - Increment play count and update last_played_at

### 3.2 API Handlers
- ✅ GET `/api/v1/records` - list records
- ⏳ POST `/api/v1/records` - create record
  - Validate all fields
  - Handle optional fields (album_title, release_year, catalog_number, condition, notes)
- ⏳ GET `/api/v1/records/{id}` - get single record
- ⏳ PUT `/api/v1/records/{id}` - update record
- ⏳ DELETE `/api/v1/records/{id}` - delete record
- ⏳ POST `/api/v1/records/{id}/play` - track playback
- ⏳ GET `/api/v1/records/recent` - recently played (query: `GetRecentlyPlayedRecords`)
- ⏳ GET `/api/v1/records/popular` - most played (query: `GetMostPlayedRecords`)

### 3.3 Features to Add
- ⏳ Search records by title (query exists: `SearchRecordsByTitle`)
- ⏳ Search records by album (query exists: `SearchRecordsByAlbum`)
- ⏳ Filter by artist (query exists: `GetRecordsByArtist`)
- ⏳ Filter by location (query exists: `GetRecordsByLocation`)
- ⏳ Filter by release year (query exists: `GetRecordsByReleaseYear`)
- ⏳ Filter by condition (query exists: `GetRecordsByCondition`)
- ⏳ Pagination (query exists: `ListRecordsWithPagination`)
- ⏳ Sort options (title, artist, year, play count, last played, etc.)
- ⏳ Batch operations (move multiple records to location)
- ⏳ Record condition tracking history

### 3.4 Record Fields to Handle
- **Required**: title
- **References**: artist_id, current_location_id, home_location_id
- **Optional**: album_title, release_year, catalog_number, condition, notes
- **Auto-managed**: play_count, last_played_at, created_at, updated_at

---

## 4. Location Management

### 4.1 HTML Handlers
- ✅ List all locations (`handleGetLocationsPage`)
- ⏳ View location detail page with records (`handleGetLocation`)
- ⏳ Create new location (`handleGetLocationNewForm`, `handlePostLocation`)
- ⏳ Edit location (`handleGetLocationEditForm`, `handlePutLocation`)
- ⏳ Delete location (`handleDeleteLocation`)
  - Consider what happens to records at deleted location
- ⏳ Set default location (`handlePostLocationSetDefault`)
  - Use `SetDefaultLocation` query

### 4.2 API Handlers
- ✅ GET `/api/v1/locations` - list locations
- ⏳ POST `/api/v1/locations` - create location
- ⏳ GET `/api/v1/locations/{id}` - get single location
- ⏳ PUT `/api/v1/locations/{id}` - update location
- ⏳ DELETE `/api/v1/locations/{id}` - delete location
- ⏳ POST `/api/v1/locations/{id}/set-default` - set as default
- ⏳ GET `/api/v1/locations/{id}/records` - get records at location

### 4.3 Features to Add
- ⏳ Search locations by name (query exists: `SearchLocationsByName`)
- ⏳ Show record count per location
- ⏳ Highlight default location in UI
- ⏳ Prevent deletion of location with records (or cascade to null)

---

## 5. User Profile Management

### 5.1 HTML Handlers
- ⏳ View profile (`handleGetProfile`)
  - Show user info (email, username, role)
  - Show account stats (member since, total records, etc.)
- ⏳ Edit profile (`handleGetProfileEditForm`, `handlePutProfile`)
  - Allow updating email, username
  - Validate email uniqueness
- ⏳ Change password (`handleGetPasswordForm`, `handlePutPassword`)
  - Verify current password
  - Validate new password strength
  - Hash and update password

### 5.2 API Handlers
- ⏳ GET `/api/v1/me` - get current user
- ⏳ GET `/api/v1/users/{id}` - get user by ID (own profile or admin)
- ⏳ PUT `/api/v1/users/{id}` - update user (own profile or admin)
- ⏳ DELETE `/api/v1/users/{id}` - delete user (admin only)
- ⏳ PUT `/api/v1/users/{id}/password` - change password

### 5.3 Features to Add
- ⏳ Permission checks (users can only edit own profile unless admin)
- ⏳ Prevent self-deletion
- ⏳ Session management (view active sessions, revoke sessions)
- ⏳ API token management (create, list, revoke API tokens)

---

## 6. Statistics & Analytics

### 6.1 API Handler
- ⏳ GET `/api/v1/stats` - get collection statistics
  - Total records, artists, locations
  - Most played records
  - Recently added records
  - Play count trends
  - Collection by condition breakdown

### 6.2 Dashboard Integration
- ⏳ Display stats on dashboard page
- ⏳ Charts/graphs for visual representation (optional)

---

## 7. Templates & UI

### 7.1 Missing Page Templates ⏳
- ⏳ `templates/pages/dashboard.html` - user dashboard
- ⏳ `templates/pages/artist-detail.html` - artist detail view
- ⏳ `templates/pages/record-detail.html` - record detail view
- ⏳ `templates/pages/location-detail.html` - location detail view
- ⏳ `templates/pages/profile.html` - user profile view

### 7.2 Missing Form Templates ⏳
- ⏳ `templates/partials/update-artist-form.html` - edit artist
- ⏳ `templates/partials/create-record-form.html` - new record
- ⏳ `templates/partials/update-record-form.html` - edit record
- ⏳ `templates/partials/create-location-form.html` - new location
- ⏳ `templates/partials/update-location-form.html` - edit location
- ⏳ `templates/partials/profile-edit-form.html` - edit profile
- ⏳ `templates/partials/password-form.html` - change password

### 7.3 Missing Partial Templates ⏳
- ⏳ `templates/partials/records-row.html` - record table row
- ⏳ `templates/partials/record-play-count.html` - play count display
- ⏳ `templates/partials/locations-row.html` - location table row
- ⏳ `templates/partials/locations-list.html` - locations list

### 7.4 HTMX Enhancements ⏳
- ⏳ Inline editing for artist/record/location names
- ⏳ Modal dialogs for forms
- ⏳ Optimistic UI updates
- ⏳ Loading states and error handling
- ⏳ Toast notifications for success/error messages

---

## 8. Testing

### 8.1 Existing Tests
- ✅ `TestHandleLogin` - login functionality
- ✅ `TestHandleAPILogin` - API login
- ✅ `TestHandleAPISignup` - API signup

### 8.2 Tests to Add ⏳
- ⏳ Artist CRUD operations
- ⏳ Record CRUD operations (especially with artist/location references)
- ⏳ Location CRUD operations (especially default location logic)
- ⏳ User profile update
- ⏳ Password change
- ⏳ Authentication middleware
- ⏳ Validation edge cases
- ⏳ Database constraint violations (e.g., duplicate artist names)
- ⏳ Playback tracking
- ⏳ Search and filter functionality

---

## 9. Database & Queries

### 9.1 Missing SQL Queries ⏳
Check `data/sql/queries/locations.sql` and `data/sql/queries/users.sql` for:
- ⏳ Location queries (CreateLocation, UpdateLocation, DeleteLocation, etc.)
- ⏳ User profile update queries
- ⏳ User password update queries
- ⏳ Session management queries (list user sessions, revoke sessions)

### 9.2 Data Integrity ⏳
- ⏳ Handle foreign key constraints properly (ON DELETE SET NULL vs CASCADE)
- ⏳ Validate that default location logic works (only one location can be default)
- ⏳ Test concurrent playback tracking (play_count increments)

---

## 10. Deployment & Configuration

### 10.1 Environment Variables ⏳
- ⏳ Create `.env.example` file with:
  - `DATABASE_PATH` - SQLite database file path
  - `JWT_SECRET` - Secret key for JWT signing
  - `SERVER_HOST` - Server host
  - `SERVER_PORT` - Server port
  - `LOG_LEVEL` - Logging level (debug, info, warn, error)
  - `ENVIRONMENT` - Environment (dev, prod)
  - `SESSION_DURATION` - Session cookie duration
  - `COOKIE_SECURE` - Enable secure cookies (true for production)

### 10.2 Production Readiness ⏳
- ⏳ Dockerfile for containerization
- ⏳ Docker Compose for local development
- ⏳ Database backup strategy
- ⏳ Logging configuration
- ⏳ Health check endpoint
- ⏳ Graceful shutdown handling
- ⏳ Static file serving configuration
- ⏳ HTTPS/TLS configuration

---

## 11. Nice-to-Have Features

### 11.1 Advanced Features ⏳
- ⏳ Record images/cover art upload
- ⏳ Bulk import from CSV
- ⏳ Export collection to CSV/JSON
- ⏳ Barcode scanning for catalog numbers
- ⏳ Discogs API integration for metadata
- ⏳ Collection value estimation
- ⏳ Wishlist/Want list functionality
- ⏳ Loan tracking (who borrowed what record)
- ⏳ Listening history/stats over time
- ⏳ Genre/tag management
- ⏳ Multi-user collections (shared ownership)

### 11.2 UI/UX Improvements ⏳
- ⏳ Dark mode toggle
- ⏳ Responsive design refinements
- ⏳ Keyboard shortcuts
- ⏳ Advanced search with multiple filters
- ⏳ Saved searches/filters
- ⏳ Grid/list view toggle for records
- ⏳ Print-friendly collection list

---

## Priority Order

### Phase 1: Core CRUD Operations (Essential)
1. Complete Record CRUD (create, view, edit, delete)
2. Complete Artist edit/delete handlers
3. Complete Location CRUD operations
4. Create all missing templates

### Phase 2: User Management
1. User profile view and edit
2. Password change functionality
3. Enable authentication middleware for protected routes

### Phase 3: Enhanced Features
1. Search and filtering
2. Pagination
3. Playback tracking
4. Statistics dashboard

### Phase 4: Security & Polish
1. Move secrets to environment variables
2. Enable CSRF protection
3. Add comprehensive tests
4. Production configuration

### Phase 5: Advanced Features (Optional)
1. Cover art uploads
2. Import/export functionality
3. Third-party integrations
4. Advanced analytics
