// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: locations.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const countLocations = `-- name: CountLocations :one
SELECT COUNT(*) FROM locations
`

func (q *Queries) CountLocations(ctx context.Context) (int64, error) {
	row := q.db.QueryRow(ctx, countLocations)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const createLocation = `-- name: CreateLocation :one
INSERT INTO locations (name, description, is_default)
VALUES ($1, $2, $3)
RETURNING id, name, description, is_default, created_at, updated_at
`

type CreateLocationParams struct {
	Name        string      `json:"name"`
	Description pgtype.Text `json:"description"`
	IsDefault   pgtype.Bool `json:"is_default"`
}

func (q *Queries) CreateLocation(ctx context.Context, arg CreateLocationParams) (Location, error) {
	row := q.db.QueryRow(ctx, createLocation, arg.Name, arg.Description, arg.IsDefault)
	var i Location
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.IsDefault,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteLocation = `-- name: DeleteLocation :exec
DELETE FROM locations
WHERE id = $1
`

func (q *Queries) DeleteLocation(ctx context.Context, id int32) error {
	_, err := q.db.Exec(ctx, deleteLocation, id)
	return err
}

const getDefaultLocation = `-- name: GetDefaultLocation :one
SELECT id, name, description, is_default, created_at, updated_at
FROM locations
WHERE is_default = true
LIMIT 1
`

func (q *Queries) GetDefaultLocation(ctx context.Context) (Location, error) {
	row := q.db.QueryRow(ctx, getDefaultLocation)
	var i Location
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.IsDefault,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getLocation = `-- name: GetLocation :one
SELECT id, name, description, is_default, created_at, updated_at
FROM locations
WHERE id = $1
`

func (q *Queries) GetLocation(ctx context.Context, id int32) (Location, error) {
	row := q.db.QueryRow(ctx, getLocation, id)
	var i Location
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.IsDefault,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getLocationByName = `-- name: GetLocationByName :one
SELECT id, name, description, is_default, created_at, updated_at
FROM locations
WHERE name = $1
`

func (q *Queries) GetLocationByName(ctx context.Context, name string) (Location, error) {
	row := q.db.QueryRow(ctx, getLocationByName, name)
	var i Location
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.IsDefault,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const listLocations = `-- name: ListLocations :many
SELECT id, name, description, is_default, created_at, updated_at
FROM locations
ORDER BY name ASC
`

func (q *Queries) ListLocations(ctx context.Context) ([]Location, error) {
	rows, err := q.db.Query(ctx, listLocations)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Location{}
	for rows.Next() {
		var i Location
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.IsDefault,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listLocationsWithPagination = `-- name: ListLocationsWithPagination :many
SELECT id, name, description, is_default, created_at, updated_at
FROM locations
ORDER BY name ASC
LIMIT $1 OFFSET $2
`

type ListLocationsWithPaginationParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) ListLocationsWithPagination(ctx context.Context, arg ListLocationsWithPaginationParams) ([]Location, error) {
	rows, err := q.db.Query(ctx, listLocationsWithPagination, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Location{}
	for rows.Next() {
		var i Location
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.IsDefault,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const searchLocationsByName = `-- name: SearchLocationsByName :many
SELECT id, name, description, is_default, created_at, updated_at
FROM locations
WHERE name ILIKE '%' || $1 || '%'
ORDER BY name ASC
`

func (q *Queries) SearchLocationsByName(ctx context.Context, dollar_1 pgtype.Text) ([]Location, error) {
	rows, err := q.db.Query(ctx, searchLocationsByName, dollar_1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Location{}
	for rows.Next() {
		var i Location
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.IsDefault,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const setDefaultLocation = `-- name: SetDefaultLocation :exec
UPDATE locations
SET is_default = CASE
    WHEN id = $1 THEN true
    ELSE false
END
`

func (q *Queries) SetDefaultLocation(ctx context.Context, id int32) error {
	_, err := q.db.Exec(ctx, setDefaultLocation, id)
	return err
}

const updateLocation = `-- name: UpdateLocation :one
UPDATE locations
SET name = $2, description = $3, is_default = $4
WHERE id = $1
RETURNING id, name, description, is_default, created_at, updated_at
`

type UpdateLocationParams struct {
	ID          int32       `json:"id"`
	Name        string      `json:"name"`
	Description pgtype.Text `json:"description"`
	IsDefault   pgtype.Bool `json:"is_default"`
}

func (q *Queries) UpdateLocation(ctx context.Context, arg UpdateLocationParams) (Location, error) {
	row := q.db.QueryRow(ctx, updateLocation,
		arg.ID,
		arg.Name,
		arg.Description,
		arg.IsDefault,
	)
	var i Location
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.IsDefault,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateLocationName = `-- name: UpdateLocationName :one
UPDATE locations
SET name = $2
WHERE id = $1
RETURNING id, name, description, is_default, created_at, updated_at
`

type UpdateLocationNameParams struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

func (q *Queries) UpdateLocationName(ctx context.Context, arg UpdateLocationNameParams) (Location, error) {
	row := q.db.QueryRow(ctx, updateLocationName, arg.ID, arg.Name)
	var i Location
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.IsDefault,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
