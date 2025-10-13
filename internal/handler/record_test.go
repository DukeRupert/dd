package handler

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/dukerupert/dd/internal/store"
)

// TestCreateRecord_Success tests successful record creation
func TestCreateRecord_Success(t *testing.T) {
	db, queries := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create artist
	artist, err := queries.CreateArtist(ctx, "Pink Floyd")
	if err != nil {
		t.Fatalf("Failed to create artist: %v", err)
	}

	// Create location
	location, err := queries.CreateLocation(ctx, store.CreateLocationParams{
		Name:        "Main Shelf",
		Description: sql.NullString{String: "Test", Valid: true},
		IsDefault:   sql.NullBool{Bool: true, Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create location: %v", err)
	}

	// Create record with all fields
	record, err := queries.CreateRecord(ctx, store.CreateRecordParams{
		Title:             "Dark Side of the Moon",
		ArtistID:          sql.NullInt64{Int64: artist.ID, Valid: true},
		AlbumTitle:        sql.NullString{String: "Dark Side of the Moon", Valid: true},
		ReleaseYear:       sql.NullInt64{Int64: 1973, Valid: true},
		CurrentLocationID: sql.NullInt64{Int64: location.ID, Valid: true},
		HomeLocationID:    sql.NullInt64{Int64: location.ID, Valid: true},
		CatalogNumber:     sql.NullString{String: "SHVL 804", Valid: true},
		Condition:         sql.NullString{String: "Near Mint", Valid: true},
		Notes:             sql.NullString{String: "Original UK pressing", Valid: true},
		PlayCount:         sql.NullInt64{Int64: 0, Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateRecord() error = %v", err)
	}

	if record.Title != "Dark Side of the Moon" {
		t.Errorf("Record title = %v, want %v", record.Title, "Dark Side of the Moon")
	}

	if !record.ArtistID.Valid || record.ArtistID.Int64 != artist.ID {
		t.Errorf("Record artist_id = %v, want %v", record.ArtistID, artist.ID)
	}
}

// TestCreateRecord_MinimalFields tests record with only required fields
func TestCreateRecord_MinimalFields(t *testing.T) {
	db, queries := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create record with only title (minimum required)
	record, err := queries.CreateRecord(ctx, store.CreateRecordParams{
		Title:     "Unknown Record",
		PlayCount: sql.NullInt64{Int64: 0, Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateRecord() error = %v", err)
	}

	if record.Title != "Unknown Record" {
		t.Errorf("Record title = %v, want %v", record.Title, "Unknown Record")
	}

	// Verify optional fields are NULL
	if record.ArtistID.Valid {
		t.Error("ArtistID should be NULL for minimal record")
	}
	if record.AlbumTitle.Valid {
		t.Error("AlbumTitle should be NULL for minimal record")
	}
	if record.ReleaseYear.Valid {
		t.Error("ReleaseYear should be NULL for minimal record")
	}
}

// TestCreateRecord_EmptyTitle tests empty title validation
func TestCreateRecord_EmptyTitle(t *testing.T) {
	db, queries := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	_, err := queries.CreateRecord(ctx, store.CreateRecordParams{
		Title:     "",
		PlayCount: sql.NullInt64{Int64: 0, Valid: true},
	})

	if err == nil {
		t.Error("CreateRecord() should fail with empty title")
	}
}

// TestGetRecordWithDetails tests the join query
func TestGetRecordWithDetails(t *testing.T) {
	db, queries := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create artist
	artist, err := queries.CreateArtist(ctx, "The Beatles")
	if err != nil {
		t.Fatalf("Failed to create artist: %v", err)
	}

	// Create locations
	currentLoc, err := queries.CreateLocation(ctx, store.CreateLocationParams{
		Name:        "Currently Playing",
		Description: sql.NullString{String: "Test", Valid: true},
		IsDefault:   sql.NullBool{Bool: false, Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create current location: %v", err)
	}

	homeLoc, err := queries.CreateLocation(ctx, store.CreateLocationParams{
		Name:        "Home Shelf",
		Description: sql.NullString{String: "Test", Valid: true},
		IsDefault:   sql.NullBool{Bool: true, Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create home location: %v", err)
	}

	// Create record
	record, err := queries.CreateRecord(ctx, store.CreateRecordParams{
		Title:             "Abbey Road",
		ArtistID:          sql.NullInt64{Int64: artist.ID, Valid: true},
		CurrentLocationID: sql.NullInt64{Int64: currentLoc.ID, Valid: true},
		HomeLocationID:    sql.NullInt64{Int64: homeLoc.ID, Valid: true},
		PlayCount:         sql.NullInt64{Int64: 0, Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create record: %v", err)
	}

	// Get record with details
	details, err := queries.GetRecordWithDetails(ctx, record.ID)
	if err != nil {
		t.Fatalf("GetRecordWithDetails() error = %v", err)
	}

	// Verify artist name is joined
	if !details.ArtistName.Valid || details.ArtistName.String != "The Beatles" {
		t.Errorf("ArtistName = %v, want %v", details.ArtistName, "The Beatles")
	}

	// Verify current location name is joined
	if !details.CurrentLocationName.Valid || details.CurrentLocationName.String != "Currently Playing" {
		t.Errorf("CurrentLocationName = %v, want %v", details.CurrentLocationName, "Currently Playing")
	}

	// Verify home location name is joined
	if !details.HomeLocationName.Valid || details.HomeLocationName.String != "Home Shelf" {
		t.Errorf("HomeLocationName = %v, want %v", details.HomeLocationName, "Home Shelf")
	}
}

// TestUpdateRecord tests record updates
func TestUpdateRecord(t *testing.T) {
	db, queries := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create initial record
	record, err := queries.CreateRecord(ctx, store.CreateRecordParams{
		Title:     "Original Title",
		PlayCount: sql.NullInt64{Int64: 0, Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create record: %v", err)
	}

	// Create artist for update
	artist, err := queries.CreateArtist(ctx, "Updated Artist")
	if err != nil {
		t.Fatalf("Failed to create artist: %v", err)
	}

	// Update record
	updated, err := queries.UpdateRecord(ctx, store.UpdateRecordParams{
		ID:          record.ID,
		Title:       "Updated Title",
		ArtistID:    sql.NullInt64{Int64: artist.ID, Valid: true},
		AlbumTitle:  sql.NullString{String: "New Album", Valid: true},
		ReleaseYear: sql.NullInt64{Int64: 2020, Valid: true},
		Condition:   sql.NullString{String: "Good", Valid: true},
		Notes:       sql.NullString{String: "Updated notes", Valid: true},
	})
	if err != nil {
		t.Fatalf("UpdateRecord() error = %v", err)
	}

	if updated.Title != "Updated Title" {
		t.Errorf("Updated title = %v, want %v", updated.Title, "Updated Title")
	}

	if !updated.ArtistID.Valid || updated.ArtistID.Int64 != artist.ID {
		t.Errorf("Updated artist_id = %v, want %v", updated.ArtistID, artist.ID)
	}
}

// TestUpdateRecordLocation tests location update
func TestUpdateRecordLocation(t *testing.T) {
	db, queries := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create locations
	loc1, err := queries.CreateLocation(ctx, store.CreateLocationParams{
		Name:        "Location 1",
		Description: sql.NullString{String: "Test", Valid: true},
		IsDefault:   sql.NullBool{Bool: true, Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create location 1: %v", err)
	}

	loc2, err := queries.CreateLocation(ctx, store.CreateLocationParams{
		Name:        "Location 2",
		Description: sql.NullString{String: "Test", Valid: true},
		IsDefault:   sql.NullBool{Bool: false, Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create location 2: %v", err)
	}

	// Create record at location 1
	record, err := queries.CreateRecord(ctx, store.CreateRecordParams{
		Title:             "Test Record",
		CurrentLocationID: sql.NullInt64{Int64: loc1.ID, Valid: true},
		PlayCount:         sql.NullInt64{Int64: 0, Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create record: %v", err)
	}

	// Move to location 2
	updated, err := queries.UpdateRecordLocation(ctx, store.UpdateRecordLocationParams{
		CurrentLocationID: sql.NullInt64{Int64: loc2.ID, Valid: true},
		ID:                record.ID,
	})
	if err != nil {
		t.Fatalf("UpdateRecordLocation() error = %v", err)
	}

	if !updated.CurrentLocationID.Valid || updated.CurrentLocationID.Int64 != loc2.ID {
		t.Errorf("CurrentLocationID = %v, want %v", updated.CurrentLocationID, loc2.ID)
	}
}

// TestUpdateRecordCondition tests condition update
func TestUpdateRecordCondition(t *testing.T) {
	db, queries := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create record
	record, err := queries.CreateRecord(ctx, store.CreateRecordParams{
		Title:     "Test Record",
		Condition: sql.NullString{String: "Mint", Valid: true},
		PlayCount: sql.NullInt64{Int64: 0, Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create record: %v", err)
	}

	// Update condition
	updated, err := queries.UpdateRecordCondition(ctx, store.UpdateRecordConditionParams{
		Condition: sql.NullString{String: "Good", Valid: true},
		ID:        record.ID,
	})
	if err != nil {
		t.Fatalf("UpdateRecordCondition() error = %v", err)
	}

	if !updated.Condition.Valid || updated.Condition.String != "Good" {
		t.Errorf("Condition = %v, want %v", updated.Condition, "Good")
	}
}

// TestRecordPlayback tests play count tracking
func TestRecordPlayback(t *testing.T) {
	db, queries := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create record
	record, err := queries.CreateRecord(ctx, store.CreateRecordParams{
		Title:     "Test Record",
		PlayCount: sql.NullInt64{Int64: 0, Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create record: %v", err)
	}

	// Verify initial play count
	if record.PlayCount.Int64 != 0 {
		t.Errorf("Initial play count = %v, want 0", record.PlayCount.Int64)
	}

	if record.LastPlayedAt.Valid {
		t.Error("LastPlayedAt should be NULL initially")
	}

	// Record playback
	played, err := queries.RecordPlayback(ctx, record.ID)
	if err != nil {
		t.Fatalf("RecordPlayback() error = %v", err)
	}

	// Verify play count incremented
	if played.PlayCount.Int64 != 1 {
		t.Errorf("Play count after first play = %v, want 1", played.PlayCount.Int64)
	}

	// Verify last_played_at is set
	if !played.LastPlayedAt.Valid {
		t.Error("LastPlayedAt should be set after playback")
	}

	// Wait to ensure timestamp difference
	time.Sleep(1100 * time.Millisecond)

	// Record playback again
	played2, err := queries.RecordPlayback(ctx, record.ID)
	if err != nil {
		t.Fatalf("RecordPlayback() second time error = %v", err)
	}

	// Verify play count incremented again
	if played2.PlayCount.Int64 != 2 {
		t.Errorf("Play count after second play = %v, want 2", played2.PlayCount.Int64)
	}

	// Verify last_played_at is updated
	if !played2.LastPlayedAt.Valid {
		t.Error("LastPlayedAt should still be set")
	}

	if !played2.LastPlayedAt.Time.After(played.LastPlayedAt.Time) {
		t.Error("LastPlayedAt should be updated to more recent time")
	}
}

// TestDeleteRecord tests record deletion
func TestDeleteRecord(t *testing.T) {
	db, queries := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create record
	record, err := queries.CreateRecord(ctx, store.CreateRecordParams{
		Title:     "Temporary Record",
		PlayCount: sql.NullInt64{Int64: 0, Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create record: %v", err)
	}

	// Delete record
	err = queries.DeleteRecord(ctx, record.ID)
	if err != nil {
		t.Fatalf("DeleteRecord() error = %v", err)
	}

	// Verify deletion
	_, err = queries.GetRecord(ctx, record.ID)
	if err == nil {
		t.Error("GetRecord() should fail for deleted record")
	}
}

// TestGetRecordsByArtist tests filtering by artist
func TestGetRecordsByArtist(t *testing.T) {
	db, queries := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create artists
	artist1, err := queries.CreateArtist(ctx, "Artist 1")
	if err != nil {
		t.Fatalf("Failed to create artist 1: %v", err)
	}

	artist2, err := queries.CreateArtist(ctx, "Artist 2")
	if err != nil {
		t.Fatalf("Failed to create artist 2: %v", err)
	}

	// Create records for artist 1
	for i := 0; i < 3; i++ {
		_, err := queries.CreateRecord(ctx, store.CreateRecordParams{
			Title:     "Record " + string(rune('A'+i)),
			ArtistID:  sql.NullInt64{Int64: artist1.ID, Valid: true},
			PlayCount: sql.NullInt64{Int64: 0, Valid: true},
		})
		if err != nil {
			t.Fatalf("Failed to create record: %v", err)
		}
	}

	// Create records for artist 2
	for i := 0; i < 2; i++ {
		_, err := queries.CreateRecord(ctx, store.CreateRecordParams{
			Title:     "Record " + string(rune('X'+i)),
			ArtistID:  sql.NullInt64{Int64: artist2.ID, Valid: true},
			PlayCount: sql.NullInt64{Int64: 0, Valid: true},
		})
		if err != nil {
			t.Fatalf("Failed to create record: %v", err)
		}
	}

	// Get records for artist 1
	records1, err := queries.GetRecordsByArtist(ctx, sql.NullInt64{Int64: artist1.ID, Valid: true})
	if err != nil {
		t.Fatalf("GetRecordsByArtist() error = %v", err)
	}

	if len(records1) != 3 {
		t.Errorf("Artist 1 record count = %d, want 3", len(records1))
	}

	// Get records for artist 2
	records2, err := queries.GetRecordsByArtist(ctx, sql.NullInt64{Int64: artist2.ID, Valid: true})
	if err != nil {
		t.Fatalf("GetRecordsByArtist() error = %v", err)
	}

	if len(records2) != 2 {
		t.Errorf("Artist 2 record count = %d, want 2", len(records2))
	}
}

// TestGetRecordsByLocation tests filtering by location
func TestGetRecordsByLocation(t *testing.T) {
	db, queries := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create locations
	loc1, err := queries.CreateLocation(ctx, store.CreateLocationParams{
		Name:        "Shelf A",
		Description: sql.NullString{String: "Test", Valid: true},
		IsDefault:   sql.NullBool{Bool: true, Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create location 1: %v", err)
	}

	loc2, err := queries.CreateLocation(ctx, store.CreateLocationParams{
		Name:        "Shelf B",
		Description: sql.NullString{String: "Test", Valid: true},
		IsDefault:   sql.NullBool{Bool: false, Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create location 2: %v", err)
	}

	// Create records at location 1
	for i := 0; i < 4; i++ {
		_, err := queries.CreateRecord(ctx, store.CreateRecordParams{
			Title:             "Record " + string(rune('A'+i)),
			CurrentLocationID: sql.NullInt64{Int64: loc1.ID, Valid: true},
			PlayCount:         sql.NullInt64{Int64: 0, Valid: true},
		})
		if err != nil {
			t.Fatalf("Failed to create record: %v", err)
		}
	}

	// Create records at location 2
	for i := 0; i < 2; i++ {
		_, err := queries.CreateRecord(ctx, store.CreateRecordParams{
			Title:             "Record " + string(rune('X'+i)),
			CurrentLocationID: sql.NullInt64{Int64: loc2.ID, Valid: true},
			PlayCount:         sql.NullInt64{Int64: 0, Valid: true},
		})
		if err != nil {
			t.Fatalf("Failed to create record: %v", err)
		}
	}

	// Get records at location 1
	records1, err := queries.GetRecordsByLocation(ctx, sql.NullInt64{Int64: loc1.ID, Valid: true})
	if err != nil {
		t.Fatalf("GetRecordsByLocation() error = %v", err)
	}

	if len(records1) != 4 {
		t.Errorf("Location 1 record count = %d, want 4", len(records1))
	}

	// Get records at location 2
	records2, err := queries.GetRecordsByLocation(ctx, sql.NullInt64{Int64: loc2.ID, Valid: true})
	if err != nil {
		t.Fatalf("GetRecordsByLocation() error = %v", err)
	}

	if len(records2) != 2 {
		t.Errorf("Location 2 record count = %d, want 2", len(records2))
	}
}

// TestGetMostPlayedRecords tests sorting by play count
func TestGetMostPlayedRecords(t *testing.T) {
	db, queries := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create records with different play counts
	record1, _ := queries.CreateRecord(ctx, store.CreateRecordParams{
		Title:     "Most Played",
		PlayCount: sql.NullInt64{Int64: 0, Valid: true},
	})

	record2, _ := queries.CreateRecord(ctx, store.CreateRecordParams{
		Title:     "Second Most",
		PlayCount: sql.NullInt64{Int64: 0, Valid: true},
	})

	record3, _ := queries.CreateRecord(ctx, store.CreateRecordParams{
		Title:     "Third Most",
		PlayCount: sql.NullInt64{Int64: 0, Valid: true},
	})

	// Simulate plays
	for i := 0; i < 5; i++ {
		queries.RecordPlayback(ctx, record1.ID)
	}
	for i := 0; i < 3; i++ {
		queries.RecordPlayback(ctx, record2.ID)
	}
	queries.RecordPlayback(ctx, record3.ID)

	// Get most played
	mostPlayed, err := queries.GetMostPlayedRecords(ctx, 10)
	if err != nil {
		t.Fatalf("GetMostPlayedRecords() error = %v", err)
	}

	if len(mostPlayed) != 3 {
		t.Errorf("Most played count = %d, want 3", len(mostPlayed))
	}

	// Verify order (descending by play count)
	if mostPlayed[0].PlayCount.Int64 != 5 {
		t.Errorf("First record play count = %d, want 5", mostPlayed[0].PlayCount.Int64)
	}
	if mostPlayed[1].PlayCount.Int64 != 3 {
		t.Errorf("Second record play count = %d, want 3", mostPlayed[1].PlayCount.Int64)
	}
	if mostPlayed[2].PlayCount.Int64 != 1 {
		t.Errorf("Third record play count = %d, want 1", mostPlayed[2].PlayCount.Int64)
	}
}

// TestSearchRecordsByTitle tests title search
func TestSearchRecordsByTitle(t *testing.T) {
	db, queries := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create test records
	titles := []string{
		"Dark Side of the Moon",
		"The Dark Knight Returns",
		"Moonlight Sonata",
		"Abbey Road",
	}

	for _, title := range titles {
		_, err := queries.CreateRecord(ctx, store.CreateRecordParams{
			Title:     title,
			PlayCount: sql.NullInt64{Int64: 0, Valid: true},
		})
		if err != nil {
			t.Fatalf("Failed to create record: %v", err)
		}
	}

	tests := []struct {
		name          string
		searchTerm    string
		expectedCount int
		shouldContain []string
	}{
		{"search 'Dark'", "Dark", 2, []string{"Dark Side of the Moon", "The Dark Knight Returns"}},
		{"search 'Moon'", "Moon", 2, []string{"Dark Side of the Moon", "Moonlight Sonata"}},
		{"search 'Abbey'", "Abbey", 1, []string{"Abbey Road"}},
		{"search nonexistent", "Nonexistent", 0, []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := queries.SearchRecordsByTitle(ctx, sql.NullString{String: tt.searchTerm, Valid: true})
			if err != nil {
				t.Fatalf("SearchRecordsByTitle() error = %v", err)
			}

			if len(results) != tt.expectedCount {
				t.Errorf("Result count = %d, want %d", len(results), tt.expectedCount)
			}

			resultTitles := make(map[string]bool)
			for _, rec := range results {
				resultTitles[rec.Title] = true
			}

			for _, expectedTitle := range tt.shouldContain {
				if !resultTitles[expectedTitle] {
					t.Errorf("Expected title %q not found in results", expectedTitle)
				}
			}
		})
	}
}