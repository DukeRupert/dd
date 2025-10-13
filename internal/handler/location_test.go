package handler

import (
	"context"
	"database/sql"
	"testing"

	"github.com/dukerupert/dd/internal/store"
)

// TestCreateLocation_Success tests successful location creation
func TestCreateLocation_Success(t *testing.T) {
	db, queries := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	tests := []struct {
		name        string
		locationName string
		description string
		isDefault   bool
	}{
		{"basic location", "Main Shelf", "Primary storage", false},
		{"default location", "Collection", "Main collection", true},
		{"minimal location", "Shelf A", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			location, err := queries.CreateLocation(ctx, store.CreateLocationParams{
				Name:        tt.locationName,
				Description: sql.NullString{String: tt.description, Valid: tt.description != ""},
				IsDefault:   sql.NullBool{Bool: tt.isDefault, Valid: true},
			})
			if err != nil {
				t.Fatalf("CreateLocation() error = %v", err)
			}

			if location.Name != tt.locationName {
				t.Errorf("Location name = %v, want %v", location.Name, tt.locationName)
			}

			if location.IsDefault.Bool != tt.isDefault {
				t.Errorf("Location isDefault = %v, want %v", location.IsDefault.Bool, tt.isDefault)
			}
		})
	}
}

// TestCreateLocation_EmptyName tests empty name validation
func TestCreateLocation_EmptyName(t *testing.T) {
	db, queries := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	_, err := queries.CreateLocation(ctx, store.CreateLocationParams{
		Name:        "",
		Description: sql.NullString{String: "Test", Valid: true},
		IsDefault:   sql.NullBool{Bool: false, Valid: true},
	})

	if err == nil {
		t.Error("CreateLocation() should fail with empty name")
	}
}

// TestSetDefaultLocation tests default location logic
func TestSetDefaultLocation(t *testing.T) {
	db, queries := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create three locations
	loc1, err := queries.CreateLocation(ctx, store.CreateLocationParams{
		Name:        "Location 1",
		Description: sql.NullString{String: "First", Valid: true},
		IsDefault:   sql.NullBool{Bool: true, Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create location 1: %v", err)
	}

	loc2, err := queries.CreateLocation(ctx, store.CreateLocationParams{
		Name:        "Location 2",
		Description: sql.NullString{String: "Second", Valid: true},
		IsDefault:   sql.NullBool{Bool: false, Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create location 2: %v", err)
	}

	loc3, err := queries.CreateLocation(ctx, store.CreateLocationParams{
		Name:        "Location 3",
		Description: sql.NullString{String: "Third", Valid: true},
		IsDefault:   sql.NullBool{Bool: false, Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create location 3: %v", err)
	}

	// Set location 2 as default
	err = queries.SetDefaultLocation(ctx, loc2.ID)
	if err != nil {
		t.Fatalf("SetDefaultLocation() error = %v", err)
	}

	// Verify only location 2 is default
	locations, err := queries.ListLocations(ctx)
	if err != nil {
		t.Fatalf("ListLocations() error = %v", err)
	}

	for _, loc := range locations {
		shouldBeDefault := loc.ID == loc2.ID
		if loc.IsDefault.Bool != shouldBeDefault {
			t.Errorf("Location %s (ID: %d) isDefault = %v, want %v",
				loc.Name, loc.ID, loc.IsDefault.Bool, shouldBeDefault)
		}
	}

	// Verify GetDefaultLocation returns location 2
	defaultLoc, err := queries.GetDefaultLocation(ctx)
	if err != nil {
		t.Fatalf("GetDefaultLocation() error = %v", err)
	}

	if defaultLoc.ID != loc2.ID {
		t.Errorf("Default location ID = %v, want %v", defaultLoc.ID, loc2.ID)
	}

	// Change default to location 3
	err = queries.SetDefaultLocation(ctx, loc3.ID)
	if err != nil {
		t.Fatalf("SetDefaultLocation() error = %v", err)
	}

	// Verify only location 3 is default now
	defaultLoc, err = queries.GetDefaultLocation(ctx)
	if err != nil {
		t.Fatalf("GetDefaultLocation() error = %v", err)
	}

	if defaultLoc.ID != loc3.ID {
		t.Errorf("Default location ID = %v, want %v", defaultLoc.ID, loc3.ID)
	}

	// Verify loc1 and loc2 are not default
	loc1Check, _ := queries.GetLocation(ctx, loc1.ID)
	loc2Check, _ := queries.GetLocation(ctx, loc2.ID)

	if loc1Check.IsDefault.Bool {
		t.Error("Location 1 should not be default")
	}
	if loc2Check.IsDefault.Bool {
		t.Error("Location 2 should not be default")
	}
}

// TestGetDefaultLocation_None tests when no default exists
func TestGetDefaultLocation_None(t *testing.T) {
	db, queries := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Delete default locations
	for i := 1; i < 4; i++ {
		queries.DeleteLocation(ctx, int64(i))
	}

	// Create locations without any default
	_, err := queries.CreateLocation(ctx, store.CreateLocationParams{
		Name:        "Location 1",
		Description: sql.NullString{String: "First", Valid: true},
		IsDefault:   sql.NullBool{Bool: false, Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create location: %v", err)
	}

	// Try to get default location
	_, err = queries.GetDefaultLocation(ctx)
	if err == nil {
		t.Error("GetDefaultLocation() should fail when no default exists")
	}
}

// TestUpdateLocation tests location updates
func TestUpdateLocation(t *testing.T) {
	db, queries := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create location
	location, err := queries.CreateLocation(ctx, store.CreateLocationParams{
		Name:        "Original Name",
		Description: sql.NullString{String: "Original desc", Valid: true},
		IsDefault:   sql.NullBool{Bool: false, Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create location: %v", err)
	}

	// Update location
	updated, err := queries.UpdateLocation(ctx, store.UpdateLocationParams{
		ID:          location.ID,
		Name:        "Updated Name",
		Description: sql.NullString{String: "Updated desc", Valid: true},
		IsDefault:   sql.NullBool{Bool: true, Valid: true},
	})
	if err != nil {
		t.Fatalf("UpdateLocation() error = %v", err)
	}

	if updated.Name != "Updated Name" {
		t.Errorf("Updated name = %v, want %v", updated.Name, "Updated Name")
	}

	if updated.Description.String != "Updated desc" {
		t.Errorf("Updated description = %v, want %v", updated.Description.String, "Updated desc")
	}

	if !updated.IsDefault.Bool {
		t.Error("Updated isDefault should be true")
	}
}

// TestUpdateLocation_EmptyName tests updating to empty name
func TestUpdateLocation_EmptyName(t *testing.T) {
	db, queries := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create location
	location, err := queries.CreateLocation(ctx, store.CreateLocationParams{
		Name:        "Valid Name",
		Description: sql.NullString{String: "Test", Valid: true},
		IsDefault:   sql.NullBool{Bool: false, Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create location: %v", err)
	}

	// Try to update to empty name
	_, err = queries.UpdateLocation(ctx, store.UpdateLocationParams{
		ID:          location.ID,
		Name:        "",
		Description: sql.NullString{String: "Test", Valid: true},
		IsDefault:   sql.NullBool{Bool: false, Valid: true},
	})

	if err == nil {
		t.Error("UpdateLocation() should fail with empty name")
	}
}

// TestDeleteLocation tests location deletion
func TestDeleteLocation(t *testing.T) {
	db, queries := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create location
	location, err := queries.CreateLocation(ctx, store.CreateLocationParams{
		Name:        "Temporary Location",
		Description: sql.NullString{String: "To be deleted", Valid: true},
		IsDefault:   sql.NullBool{Bool: false, Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create location: %v", err)
	}

	// Delete location
	err = queries.DeleteLocation(ctx, location.ID)
	if err != nil {
		t.Fatalf("DeleteLocation() error = %v", err)
	}

	// Verify deletion
	_, err = queries.GetLocation(ctx, location.ID)
	if err == nil {
		t.Error("GetLocation() should fail for deleted location")
	}
}

// TestDeleteLocation_WithRecords tests cascade behavior
func TestDeleteLocation_WithRecords(t *testing.T) {
	db, queries := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create artist
	artist, err := queries.CreateArtist(ctx, "Test Artist")
	if err != nil {
		t.Fatalf("Failed to create artist: %v", err)
	}

	// Create location
	location, err := queries.CreateLocation(ctx, store.CreateLocationParams{
		Name:        "Location With Records",
		Description: sql.NullString{String: "Test", Valid: true},
		IsDefault:   sql.NullBool{Bool: false, Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create location: %v", err)
	}

	// Create record at this location
	record, err := queries.CreateRecord(ctx, store.CreateRecordParams{
		Title:             "Test Album",
		ArtistID:          sql.NullInt64{Int64: artist.ID, Valid: true},
		CurrentLocationID: sql.NullInt64{Int64: location.ID, Valid: true},
		HomeLocationID:    sql.NullInt64{Int64: location.ID, Valid: true},
		PlayCount:         sql.NullInt64{Int64: 0, Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create record: %v", err)
	}

	// Delete location
	err = queries.DeleteLocation(ctx, location.ID)
	if err != nil {
		t.Fatalf("DeleteLocation() error = %v", err)
	}

	// Check what happened to the record
	// Based on schema: ON DELETE SET NULL for both current_location_id and home_location_id
	retrievedRecord, err := queries.GetRecord(ctx, record.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve record after location deletion: %v", err)
	}

	if retrievedRecord.CurrentLocationID.Valid {
		t.Errorf("Record current_location_id should be NULL after location deletion, got %v",
			retrievedRecord.CurrentLocationID.Int64)
	}

	if retrievedRecord.HomeLocationID.Valid {
		t.Errorf("Record home_location_id should be NULL after location deletion, got %v",
			retrievedRecord.HomeLocationID.Int64)
	}

	t.Logf("Record still exists with location IDs set to NULL (ON DELETE SET NULL behavior)")
}

// TestListLocations_Ordering tests alphabetical ordering
func TestListLocations_Ordering(t *testing.T) {
	db, queries := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create locations in non-alphabetical order
	locationNames := []string{"Zebra Shelf", "Alpha Shelf", "Bravo Shelf"}
	for _, name := range locationNames {
		_, err := queries.CreateLocation(ctx, store.CreateLocationParams{
			Name:        name,
			Description: sql.NullString{String: "Test", Valid: true},
			IsDefault:   sql.NullBool{Bool: false, Valid: true},
		})
		if err != nil {
			t.Fatalf("Failed to create location %s: %v", name, err)
		}
	}

	// Get all locations
	locations, err := queries.ListLocations(ctx)
	if err != nil {
		t.Fatalf("ListLocations() error = %v", err)
	}

	// Verify they're in alphabetical order (plus the 3 default locations from migration)
	// Migration creates: "Main Collection", "Currently Playing", "Cleaning Station"
	expectedOrder := []string{"Alpha Shelf", "Bravo Shelf", "Cleaning Station", "Currently Playing", "Main Collection", "Zebra Shelf"}
	if len(locations) != len(expectedOrder) {
		t.Fatalf("Location count = %d, want %d", len(locations), len(expectedOrder))
	}

	for i, location := range locations {
		if location.Name != expectedOrder[i] {
			t.Errorf("Location[%d] name = %v, want %v", i, location.Name, expectedOrder[i])
		}
	}
}

// TestSearchLocationsByName tests search functionality
func TestSearchLocationsByName(t *testing.T) {
	db, queries := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Delete default locations
	for i := 1; i < 4; i++ {
		queries.DeleteLocation(ctx, int64(i))
	}

	// Create test locations
	locationNames := []string{"Main Shelf A", "Main Shelf B", "Storage Room", "Basement Storage"}
	for _, name := range locationNames {
		_, err := queries.CreateLocation(ctx, store.CreateLocationParams{
			Name:        name,
			Description: sql.NullString{String: "Test", Valid: true},
			IsDefault:   sql.NullBool{Bool: false, Valid: true},
		})
		if err != nil {
			t.Fatalf("Failed to create location %s: %v", name, err)
		}
	}

	tests := []struct {
		name          string
		searchTerm    string
		expectedCount int
		shouldContain []string
	}{
		{"search 'Main'", "Main", 2, []string{"Main Shelf A", "Main Shelf B"}},
		{"search 'Storage'", "Storage", 2, []string{"Storage Room", "Basement Storage"}},
		{"search 'Shelf'", "Shelf", 2, []string{"Main Shelf A", "Main Shelf B"}},
		{"search 'Basement'", "Basement", 1, []string{"Basement Storage"}},
		{"search nonexistent", "Nonexistent", 0, []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := queries.SearchLocationsByName(ctx, sql.NullString{String: tt.searchTerm, Valid: true})
			if err != nil {
				t.Fatalf("SearchLocationsByName() error = %v", err)
			}

			if len(results) != tt.expectedCount {
				t.Errorf("Result count = %d, want %d", len(results), tt.expectedCount)
			}

			// Verify expected names are in results
			resultNames := make(map[string]bool)
			for _, loc := range results {
				resultNames[loc.Name] = true
			}

			for _, expectedName := range tt.shouldContain {
				if !resultNames[expectedName] {
					t.Errorf("Expected location %q not found in results", expectedName)
				}
			}
		})
	}
}

// TestCountLocations tests location counting
func TestCountLocations(t *testing.T) {
	db, queries := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Count initial locations (3 from migration)
	initialCount, err := queries.CountLocations(ctx)
	if err != nil {
		t.Fatalf("CountLocations() error = %v", err)
	}

	if initialCount != 3 {
		t.Errorf("Initial count = %d, want 3 (from migration)", initialCount)
	}

	// Add more locations
	for i := 0; i < 5; i++ {
		_, err := queries.CreateLocation(ctx, store.CreateLocationParams{
			Name:        "Test Location " + string(rune('A'+i)),
			Description: sql.NullString{String: "Test", Valid: true},
			IsDefault:   sql.NullBool{Bool: false, Valid: true},
		})
		if err != nil {
			t.Fatalf("Failed to create location: %v", err)
		}
	}

	// Count again
	finalCount, err := queries.CountLocations(ctx)
	if err != nil {
		t.Fatalf("CountLocations() error = %v", err)
	}

	expectedCount := initialCount + 5
	if finalCount != expectedCount {
		t.Errorf("Final count = %d, want %d", finalCount, expectedCount)
	}
}