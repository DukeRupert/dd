package handler

import (
	"context"
	"database/sql"
	"testing"

	"github.com/dukerupert/dd/data/sql/migrations"
	"github.com/dukerupert/dd/internal/store"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
	_ "modernc.org/sqlite"
)

// setupTestDB creates an in-memory SQLite database with migrations
func setupTestDB(t *testing.T) (*sql.DB, *store.Queries) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	provider, err := goose.NewProvider(database.DialectSQLite3, db, migrations.Embed)
	if err != nil {
		t.Fatalf("Failed to create migration provider: %v", err)
	}

	if _, err := provider.Up(context.Background()); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return db, store.New(db)
}

// TestCreateArtist_DuplicateName tests that duplicate artist names are handled
func TestCreateArtist_DuplicateName(t *testing.T) {
	db, queries := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	artistName := "The Beatles"

	// Create first artist
	artist1, err := queries.CreateArtist(ctx, artistName)
	if err != nil {
		t.Fatalf("Failed to create first artist: %v", err)
	}
	if artist1.Name != artistName {
		t.Errorf("Artist name = %v, want %v", artist1.Name, artistName)
	}

	// Attempt to create duplicate
	_, err = queries.CreateArtist(ctx, artistName)
	if err == nil {
		t.Error("CreateArtist() should fail for duplicate name")
	}
}

// TestGetArtist_InvalidID tests ID validation logic
func TestGetArtist_InvalidID(t *testing.T) {
	db, queries := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	tests := []struct {
		name      string
		artistID  int64
		wantError bool
	}{
		{"nonexistent positive ID", 999, true},
		{"negative ID", -1, true},
		{"zero ID", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := queries.GetArtist(ctx, tt.artistID)
			if (err != nil) != tt.wantError {
				t.Errorf("GetArtist() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestGetArtist_WithRecords tests the join logic for artist with records
func TestGetArtist_WithRecords(t *testing.T) {
	db, queries := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create artist
	artist, err := queries.CreateArtist(ctx, "Pink Floyd")
	if err != nil {
		t.Fatalf("Failed to create artist: %v", err)
	}

	// Create location for records
	location, err := queries.CreateLocation(ctx, store.CreateLocationParams{
		Name:        "Main Collection",
		Description: sql.NullString{String: "Primary storage", Valid: true},
		IsDefault:   sql.NullBool{Bool: true, Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create location: %v", err)
	}

	// Create records for artist
	recordTitles := []string{"Dark Side of the Moon", "The Wall", "Wish You Were Here"}
	for _, title := range recordTitles {
		_, err := queries.CreateRecord(ctx, store.CreateRecordParams{
			Title:             title,
			ArtistID:          sql.NullInt64{Int64: artist.ID, Valid: true},
			CurrentLocationID: sql.NullInt64{Int64: location.ID, Valid: true},
			PlayCount:         sql.NullInt64{Int64: 0, Valid: true},
		})
		if err != nil {
			t.Fatalf("Failed to create record %s: %v", title, err)
		}
	}

	// Get artist's records
	records, err := queries.GetRecordsByArtist(ctx, sql.NullInt64{Int64: artist.ID, Valid: true})
	if err != nil {
		t.Fatalf("GetRecordsByArtist() error = %v", err)
	}

	// Verify record count
	if len(records) != len(recordTitles) {
		t.Errorf("Record count = %d, want %d", len(records), len(recordTitles))
	}

	// Verify all records belong to artist
	for _, record := range records {
		if !record.ArtistID.Valid || record.ArtistID.Int64 != artist.ID {
			t.Errorf("Record %s has artistID = %v, want %d", record.Title, record.ArtistID, artist.ID)
		}
	}
}

// TestGetArtist_NoRecords tests artist with no records
func TestGetArtist_NoRecords(t *testing.T) {
	db, queries := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create artist without records
	artist, err := queries.CreateArtist(ctx, "New Artist")
	if err != nil {
		t.Fatalf("Failed to create artist: %v", err)
	}

	// Get artist's records (should be empty)
	records, err := queries.GetRecordsByArtist(ctx, sql.NullInt64{Int64: artist.ID, Valid: true})
	if err != nil {
		t.Fatalf("GetRecordsByArtist() error = %v", err)
	}

	if len(records) != 0 {
		t.Errorf("Expected 0 records, got %d", len(records))
	}
}

// TestListArtists_Ordering tests alphabetical ordering
func TestListArtists_Ordering(t *testing.T) {
	db, queries := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create artists in non-alphabetical order
	artistNames := []string{"Zeppelin", "Beatles", "Radiohead", "Nirvana"}
	for _, name := range artistNames {
		_, err := queries.CreateArtist(ctx, name)
		if err != nil {
			t.Fatalf("Failed to create artist %s: %v", name, err)
		}
	}

	// Get all artists
	artists, err := queries.ListArtists(ctx)
	if err != nil {
		t.Fatalf("ListArtists() error = %v", err)
	}

	// Verify they're in alphabetical order
	expectedOrder := []string{"Beatles", "Nirvana", "Radiohead", "Zeppelin"}
	if len(artists) != len(expectedOrder) {
		t.Fatalf("Artist count = %d, want %d", len(artists), len(expectedOrder))
	}

	for i, artist := range artists {
		if artist.Name != expectedOrder[i] {
			t.Errorf("Artist[%d] name = %v, want %v", i, artist.Name, expectedOrder[i])
		}
	}
}

// TestCreateArtist_DatabaseConstraints tests SQLite constraint enforcement
func TestCreateArtist_DatabaseConstraints(t *testing.T) {
	db, queries := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	t.Run("NOT NULL constraint on empty name", func(t *testing.T) {
		_, err := queries.CreateArtist(ctx, "")
		if err == nil {
			t.Error("CreateArtist() should fail with empty name (NOT NULL constraint)")
		}
	})

	t.Run("UNIQUE constraint on duplicate name", func(t *testing.T) {
		name := "The Beatles"
		
		// First insert should succeed
		_, err := queries.CreateArtist(ctx, name)
		if err != nil {
			t.Fatalf("First CreateArtist() failed: %v", err)
		}

		// Second insert should fail
		_, err = queries.CreateArtist(ctx, name)
		if err == nil {
			t.Error("CreateArtist() should fail for duplicate name (UNIQUE constraint)")
		}
	})

	t.Run("valid names are accepted", func(t *testing.T) {
		validNames := []string{"U2", "Pink Floyd", "X"}
		for _, name := range validNames {
			_, err := queries.CreateArtist(ctx, name)
			if err != nil {
				t.Errorf("CreateArtist(%q) unexpected error: %v", name, err)
			}
		}
	})
}