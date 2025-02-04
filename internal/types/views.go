package types

import (
	"github.com/dukerupert/dd/db"
)

type Page struct {
    Title       string
    User        *db.User
    FlashMsg    *FlashMessage
}

type FlashMessage struct {
    Type    string // "success", "error", "info"
    Message string
}

type RecordsPage struct {
    Page
    Records     []db.Record
    CurrentPage int
    TotalPages  int
    SortBy      string
    SortOrder   string
    Genre       string
    Search      string
}

type RecordFormPage struct {
    Page
    Record    *db.Record // nil for new record, populated for edit
    Images    []db.RecordImage
    FormError map[string]string
}