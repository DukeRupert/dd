package pocketbase

import (
	"encoding/json"
)

// ListAlbums fetches a list of albums with the provided query parameters
func (c *Client) ListAlbums(params QueryParams) (*ListResult[Album], error) {
    collection := "albums"
    endpoint := "/api/collections/" + collection + "/records"
    
    body, err := c.sendRequest("GET", collection, endpoint, params)
    if err != nil {
        return nil, err
    }
    
    // Parse response into typed result
    var result ListResult[Album]
    if err := json.Unmarshal(body, &result); err != nil {
        c.Logger.Error().Err(err).Msg("Failed to unmarshal response")
        return nil, err
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
