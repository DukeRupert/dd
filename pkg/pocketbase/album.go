package pocketbase

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"
)

// AlbumListResponse represents a paginated list of albums
type AlbumListResponse struct {
	Page       int     `json:"page"`
	PerPage    int     `json:"perPage"`
	TotalPages int     `json:"totalPages"`
	TotalItems int     `json:"totalItems"`
	Items      []Album `json:"items"`
}

// ListAlbums fetches a list of albums with the provided query parameters
func (c *Client) ListAlbums(params QueryParams) (*ListResult[Album], error) {
    collection := "albums"
    endpoint := "/api/collections/" + collection + "/records"
    
    // Build URL with query parameters
    u, err := url.Parse(c.BaseURL + endpoint)
    if err != nil {
        c.Logger.Error().Err(err).Str("collection", collection).Msg("Failed to parse URL")
        return nil, err
    }
    
    u.RawQuery = params.ToURLValues().Encode()
    
    // Create request
    req, err := http.NewRequest("GET", u.String(), nil)
    if err != nil {
        c.Logger.Error().Err(err).Str("collection", collection).Msg("Failed to create request")
        return nil, err
    }
    
    // Set authentication header if authenticated
    if c.IsAuthenticated() {
        req.Header.Set("Authorization", c.authToken)
    }
    
    // Set content type
    req.Header.Set("Content-Type", "application/json")
    
    // Execute request
    c.Logger.Debug().Str("collection", collection).Msg("Fetching albums")
    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        c.Logger.Error().Err(err).Str("collection", collection).Msg("Request failed")
        return nil, err
    }
    defer resp.Body.Close()
    
    // Read response body
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        c.Logger.Error().Err(err).Msg("Failed to read response body")
        return nil, err
    }
    
    // Check for error status codes
    if resp.StatusCode >= 400 {
        c.Logger.Error().Int("status_code", resp.StatusCode).Str("response", string(body)).Msg("Request failed")
        return nil, errors.New("request failed with status " + resp.Status + ": " + string(body))
    }
    
    // Log a small sample of the response for debugging
    if len(body) > 0 {
        sampleSize := 200
        if len(body) < sampleSize {
            sampleSize = len(body)
        }
        c.Logger.Debug().Str("sample_response", string(body[:sampleSize])+"...").Msg("Response sample")
    }
    
    // Create a custom decoder with DisallowUnknownFields disabled
    decoder := json.NewDecoder(bytes.NewReader(body))
    
    // Parse response into typed result
    var result ListResult[Album]
    if err := decoder.Decode(&result); err != nil {
        // If standard decode fails, try an alternative approach
        c.Logger.Warn().Err(err).Msg("Standard JSON decode failed, trying alternative approach")
        
        // Parse to a map first to handle the time fields manually
        var rawResponse map[string]interface{}
        if err := json.Unmarshal(body, &rawResponse); err != nil {
            c.Logger.Error().Err(err).Msg("Failed to unmarshal raw response")
            return nil, err
        }
        
        // Parse page information
        result.Page = int(rawResponse["page"].(float64))
        result.PerPage = int(rawResponse["perPage"].(float64))
        result.TotalItems = int(rawResponse["totalItems"].(float64))
        result.TotalPages = int(rawResponse["totalPages"].(float64))
        
        // Parse items
        items, ok := rawResponse["items"].([]interface{})
        if !ok {
            c.Logger.Error().Msg("Items field not found or not an array")
            return nil, errors.New("unexpected response format")
        }
        
        // Initialize items slice
        result.Items = make([]Album, 0, len(items))
        
        // Process each item
        for _, item := range items {
            itemMap, ok := item.(map[string]interface{})
            if !ok {
                continue
            }
            
            album := Album{}
            
            // Set base model fields
            if id, ok := itemMap["id"].(string); ok {
                album.ID = id
            }
            
            if collectionId, ok := itemMap["collectionId"].(string); ok {
                album.CollectionID = collectionId
            }
            
            if collectionName, ok := itemMap["collectionName"].(string); ok {
                album.CollectionName = collectionName
            }
            
            // Handle time fields
            if created, ok := itemMap["created"].(string); ok {
                t, err := time.Parse("2006-01-02 15:04:05.999Z", created)
                if err == nil {
                    album.Created = PocketBaseTime(t)
                } else {
                    c.Logger.Warn().Err(err).Str("time_str", created).Msg("Failed to parse created time")
                }
            }
            
            if updated, ok := itemMap["updated"].(string); ok {
                t, err := time.Parse("2006-01-02 15:04:05.999Z", updated)
                if err == nil {
                    album.Updated = PocketBaseTime(t)
                } else {
                    c.Logger.Warn().Err(err).Str("time_str", updated).Msg("Failed to parse updated time")
                }
            }
            
            // Set album-specific fields
            if title, ok := itemMap["title"].(string); ok {
                album.Title = title
            }
            
            if artistID, ok := itemMap["artist_id"].(string); ok {
                album.ArtistID = artistID
            }
            
            if locationID, ok := itemMap["location_id"].(string); ok {
                album.LocationID = locationID
            }
            
            if releaseYear, ok := itemMap["release_year"].(float64); ok {
                album.ReleaseYear = int(releaseYear)
            }
            
            if genre, ok := itemMap["genre"].(string); ok {
                album.Genre = genre
            }
            
            if condition, ok := itemMap["condition"].(string); ok {
                album.Condition = condition
            }
            
            if purchaseDate, ok := itemMap["purchase_date"].(string); ok && purchaseDate != "" {
                t, err := time.Parse("2006-01-02 15:04:05.999Z", purchaseDate)
                if err == nil {
                    album.PurchaseDate = PocketBaseTime(t)
                } else {
                    c.Logger.Warn().Err(err).Str("time_str", purchaseDate).Msg("Failed to parse purchase date")
                }
            }
            
            if purchasePrice, ok := itemMap["purchase_price"].(float64); ok {
                album.PurchasePrice = purchasePrice
            }
            
            if notes, ok := itemMap["notes"].(string); ok {
                album.Notes = notes
            }
            
            // Handle expanded fields if present
            if expand, ok := itemMap["expand"].(map[string]interface{}); ok {
                // Artist expansion
                if artistMap, ok := expand["artist_id"].(map[string]interface{}); ok {
                    artist := &Artist{}
                    
                    if id, ok := artistMap["id"].(string); ok {
                        artist.ID = id
                    }
                    
                    if name, ok := artistMap["name"].(string); ok {
                        artist.Name = name
                    }
                    
                    // Set other artist fields as needed
                    
                    album.Expand.Artist = artist
                }
                
                // Location expansion
                if locationMap, ok := expand["location_id"].(map[string]interface{}); ok {
                    location := &Location{}
                    
                    if id, ok := locationMap["id"].(string); ok {
                        location.ID = id
                    }
                    
                    if name, ok := locationMap["name"].(string); ok {
                        location.Name = name
                    }
                    
                    // Set other location fields as needed
                    
                    album.Expand.Location = location
                }
            }
            
            // Add to results
            result.Items = append(result.Items, album)
        }
    }
    
    c.Logger.Debug().
        Str("collection", collection).
        Int("items_count", len(result.Items)).
        Int("total_items", result.TotalItems).
        Msg("Successfully fetched albums")
    
    return &result, nil
}

// GetAlbum fetches a single album by ID
func (c *Client) GetAlbum(id string) (*Album, error) {
	collection := "albums"
	endpoint := "/api/collections/" + collection + "/records/" + id

	c.Logger.Debug().Str("collection", collection).Str("id", id).Msg("Fetching album")

	responseBody, err := c.Get(endpoint)
	if err != nil {
		c.Logger.Error().Err(err).Str("collection", collection).Str("id", id).Msg("Failed to fetch album")
		return nil, err
	}

	var album Album
	if err := json.Unmarshal(responseBody, &album); err != nil {
		c.Logger.Error().Err(err).Msg("Failed to unmarshal album")
		return nil, err
	}

	c.Logger.Debug().Str("collection", collection).Str("id", id).Msg("Successfully fetched album")
	return &album, nil
}

// CreateAlbum creates a new album record
func (c *Client) CreateAlbum(data map[string]interface{}) (*Album, error) {
	collection := "albums"
	endpoint := "/api/collections/" + collection + "/records"

	c.Logger.Debug().Str("collection", collection).Msg("Creating album")

	responseBody, err := c.Post(endpoint, data)
	if err != nil {
		c.Logger.Error().Err(err).Str("collection", collection).Msg("Failed to create album")
		return nil, err
	}

	var album Album
	if err := json.Unmarshal(responseBody, &album); err != nil {
		c.Logger.Error().Err(err).Msg("Failed to unmarshal album")
		return nil, err
	}

	c.Logger.Debug().Str("collection", collection).Msg("Successfully created album")
	return &album, nil
}

// UpdateAlbum updates an existing album
func (c *Client) UpdateAlbum(id string, data map[string]interface{}) (*Album, error) {
	collection := "albums"
	endpoint := "/api/collections/" + collection + "/records/" + id

	c.Logger.Debug().Str("collection", collection).Str("id", id).Msg("Updating album")

	// Marshal request body
	requestBody, err := json.Marshal(data)
	if err != nil {
		c.Logger.Error().Err(err).Msg("Failed to marshal request body")
		return nil, err
	}

	// Make PATCH request
	responseBody, err := c.Request("PATCH", endpoint, requestBody)
	if err != nil {
		c.Logger.Error().Err(err).Str("collection", collection).Str("id", id).Msg("Failed to update album")
		return nil, err
	}

	var album Album
	if err := json.Unmarshal(responseBody, &album); err != nil {
		c.Logger.Error().Err(err).Msg("Failed to unmarshal album")
		return nil, err
	}

	c.Logger.Debug().Str("collection", collection).Str("id", id).Msg("Successfully updated album")
	return &album, nil
}

// DeleteAlbum deletes an album
func (c *Client) DeleteAlbum(id string) error {
	collection := "albums"
	endpoint := "/api/collections/" + collection + "/records/" + id

	c.Logger.Debug().Str("collection", collection).Str("id", id).Msg("Deleting album")

	_, err := c.Request("DELETE", endpoint, nil)
	if err != nil {
		c.Logger.Error().Err(err).Str("collection", collection).Str("id", id).Msg("Failed to delete album")
		return err
	}

	c.Logger.Debug().Str("collection", collection).Str("id", id).Msg("Successfully deleted album")
	return nil
}
