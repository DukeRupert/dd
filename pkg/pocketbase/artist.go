package pocketbase

import (
	"encoding/json"
)

// ListArtists fetches a list of artists with the provided query parameters
func (c *Client) ListArtists(params QueryParams) (*ListResult[Artist], error) {
    collection := "artists"
    endpoint := "/api/collections/" + collection + "/records"
    
    body, err := c.sendRequest("GET", collection, endpoint, params)
    if err != nil {
        return nil, err
    }
    
    // Parse response into typed result
    var result ListResult[Artist]
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

// GetArtist fetches a single artist by ID
func (c *Client) GetArtist(id string) (*Artist, error) {
	collection := "artists"
	endpoint := "/api/collections/" + collection + "/records/" + id

	c.Logger.Debug().Str("collection", collection).Str("id", id).Msg("Fetching artist")

	responseBody, err := c.Get(endpoint)
	if err != nil {
		c.Logger.Error().Err(err).Str("collection", collection).Str("id", id).Msg("Failed to fetch artist")
		return nil, err
	}

	var artist Artist
	if err := json.Unmarshal(responseBody, &artist); err != nil {
		c.Logger.Error().Err(err).Msg("Failed to unmarshal artist")
		return nil, err
	}

	c.Logger.Debug().Str("collection", collection).Str("id", id).Msg("Successfully fetched artist")
	return &artist, nil
}