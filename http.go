package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"text/template"

	"github.com/dukerupert/dd/internal/store"
	"github.com/go-playground/validator/v10"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// func handleGetArtistsPage(logger *slog.Logger, queries *store.Queries, t *TemplateRenderer) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

// 		// do the records exist?
// 		artists, err := queries.ListArtists(r.Context())
// 		if err != nil {
// 			logger.Error("Failed to retrieve artists from database", slog.String("error", err.Error()))
// 			http.Error(w, "Artists not found", http.StatusInternalServerError)
// 			return
// 		}

// 		type PageData struct {
// 			Title   string
// 			Artists []store.Artist
// 		}
// 		data := PageData{
// 			Title:   "Artists",
// 			Artists: artists,
// 		}

// 		// great success!
// 		logger.Info("List artists", slog.Int("artists", len(artists)))
// 		if r.Header.Get("Content-Type") == "application/json" {
// 			w.Header().Set("Content-Type", "application/json")
// 			json.NewEncoder(w).Encode(artists)
// 			return
// 		}
// 		err = t.Render(w, "artists", data)
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 	})
// }

func getArtistHandler(logger *slog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// is there an id?
		id := r.PathValue("id")
		if id == "" {
			logger.Info("Bad request. Missing path id")
			http.Error(w, "Missing parameter: id", http.StatusBadRequest)
			return
		}

		// is it an integer?
		artistID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			logger.Error("Invalid parameter id", slog.String("error", err.Error()))
			http.Error(w, "Invalid parameter: id", http.StatusBadRequest)
			return
		}

		// does the record exist?
		artist, err := queries.GetArtist(r.Context(), int64(artistID))
		if err != nil {
			logger.Error("Failed to retrieve artist record", slog.String("error", err.Error()), slog.Int64("artistID", artistID))
			http.Error(w, "Missing record", http.StatusNotFound)
			return
		}

		// great success!
		logger.Info("Artist record retrieved", slog.Int64("artistID", artistID), slog.String("name", artist.Name))
		json.NewEncoder(w).Encode(artist)

	})
}

func putArtistHandler(logger *slog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// is there an id?
		id := r.PathValue("id")
		if id == "" {
			logger.Info("Bad request. Missing path id")
			http.Error(w, "Missing parameter: id", http.StatusBadRequest)
			return
		}

		// is it an integer?
		artistID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			logger.Error("Invalid parameter id", slog.String("error", err.Error()))
			http.Error(w, "Invalid parameter: id", http.StatusBadRequest)
			return
		}

		// does the record exist?
		artist, err := queries.GetArtist(r.Context(), int64(artistID))
		if err != nil {
			logger.Error("Failed to retrieve artist record", slog.String("error", err.Error()), slog.Int64("artistID", artistID))
			http.Error(w, "Missing record", http.StatusNotFound)
			return
		}

		type UpdateArtistRequest struct {
			Name string `json:"name" form:"name" validate:"required,min=1,max=100"`
		}
		var req UpdateArtistRequest

		if err := bind(r, &req); err != nil {
			if _, ok := err.(validator.ValidationErrors); ok {
				writeErrorJSON(w, "failed to encode response", http.StatusInternalServerError)
				return
			}
			writeErrorJSON(w, err.Error(), http.StatusBadRequest)
			return
		}

		logger.Debug("Request data after binding", slog.Int64("artistID", int64(artistID)), slog.String("artistName", req.Name))

		artist, err = queries.UpdateArtist(r.Context(), store.UpdateArtistParams{ID: int64(artistID), Name: req.Name})
		if err != nil {
			logger.Error("Failed to write artist record", slog.String("error", err.Error()), slog.String("Name", req.Name))
			writeErrorJSON(w, "Failed to write record", http.StatusInternalServerError)
			return
		}

		logger.Info("Artist record updated", slog.Int64("artistID", artist.ID), slog.String("Name", artist.Name))

		if r.Header.Get("HX-Request") == "true" {
			log.Debug().Msg("HX-Request header is present")
			tmpl := template.Must(template.ParseFiles("templates/partial/artists-row.html"))
			tmpl.Execute(w, artist)
			return
		}
		writeJSON(w, artist, http.StatusOK)
	})
}

func deleteArtistHandler(logger zerolog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// is there an id?
		id := r.PathValue("id")
		if id == "" {
			logger.Info().Msg("Bad request. Missing path id")
			http.Error(w, "Missing parameter: id", http.StatusBadRequest)
			return
		}

		// is it an integer?
		artistID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			logger.Error().Err(err).Msg("Invalid parameter id")
			http.Error(w, "Invalid parameter: id", http.StatusBadRequest)
			return
		}

		// delete the record
		err = queries.DeleteArtist(r.Context(), int64(artistID))
		if err != nil {
			logger.Error().Err(err).Int64("artistID", artistID).Msg("Failed to retrieve artist record")
			http.Error(w, "Missing record", http.StatusNotFound)
			return
		}

		logger.Info().Int64("artistID", int64(artistID)).Msg("Artist record created")
		// Set the status code
		w.WriteHeader(http.StatusOK)

		// Write the response body
		fmt.Fprintf(w, "Record deleted")
	})
}

func getLocationsHandler(logger zerolog.Logger, queries *store.Queries, t *TemplateRenderer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// do the records exist?
		locations, err := queries.ListLocations(r.Context())
		if err != nil {
			logger.Error().Err(err).Msg("Failed to retrieve artists from database")
			http.Error(w, "Artists not found", http.StatusInternalServerError)
			return
		}

		type PageData struct {
			Locations []store.Location
		}
		data := PageData{
			Locations: locations,
		}

		// great success!
		logger.Info().Int("artists", len(locations)).Msg("List locations")
		if r.Header.Get("Content-Type") == "application/json" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(locations)
		}
		err = t.Render(w, "locations.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func getRecordsHandler(logger zerolog.Logger, queries *store.Queries, t *TemplateRenderer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// do the records exist?
		records, err := queries.ListRecords(r.Context())
		if err != nil {
			logger.Error().Err(err).Msg("Failed to retrieve artists from database")
			http.Error(w, "Artists not found", http.StatusInternalServerError)
			return
		}

		type PageData struct {
			Title   string
			Records []store.Record
		}
		data := PageData{
			Title:   "Records",
			Records: records,
		}

		// great success!
		logger.Info().Int("records", len(records)).Msg("List records")
		if r.Header.Get("Content-Type") == "application/json" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(records)
		}
		err = t.Render(w, "albums", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func getCreateArtistForm(t *TemplateRenderer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Render(w, "create-artist-form.html", nil)
	})
}

func getUpdateArtistForm(logger zerolog.Logger, queries *store.Queries, t *TemplateRenderer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// is there an id?
		id := r.PathValue("id")
		if id == "" {
			logger.Info().Msg("Bad request. Missing path id")
			http.Error(w, "Missing parameter: id", http.StatusBadRequest)
			return
		}

		// is it an integer?
		artistID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			logger.Error().Err(err).Msg("Invalid parameter id")
			http.Error(w, "Invalid parameter: id", http.StatusBadRequest)
			return
		}

		// does the record exist?
		artist, err := queries.GetArtist(r.Context(), int64(artistID))
		if err != nil {
			logger.Error().Err(err).Int64("artistID", artistID).Msg("Failed to retrieve artist record")
			http.Error(w, "Missing record", http.StatusNotFound)
			return
		}

		t.Render(w, "update-artist-form.html", artist)
	})
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	// Get file extension
	ext := filepath.Ext(r.URL.Path)

	// Set appropriate MIME type
	switch ext {
	case ".css":
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".gif":
		w.Header().Set("Content-Type", "image/gif")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	}

	// Serve the file
	fs := http.FileServer(http.Dir("static/"))
	http.StripPrefix("/static/", fs).ServeHTTP(w, r)
}

// use a single instance of Validate, it caches struct info
var validate *validator.Validate

// Validation error response
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

type ValidationErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Details []ValidationError `json:"details"`
}

// Convert validator errors to friendly messages
func getValidationErrors(err error) []ValidationError {
	var validationErrors []ValidationError

	if validatorErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validatorErrors {
			validationError := ValidationError{
				Field: fieldError.Field(),
				Tag:   fieldError.Tag(),
				Value: fieldError.Param(),
			}

			// Custom error messages based on tag
			switch fieldError.Tag() {
			case "required":
				validationError.Message = fmt.Sprintf("%s is required", fieldError.Field())
			case "min":
				validationError.Message = fmt.Sprintf("%s must be at least %s characters", fieldError.Field(), fieldError.Param())
			case "max":
				validationError.Message = fmt.Sprintf("%s must be at most %s characters", fieldError.Field(), fieldError.Param())
			case "email":
				validationError.Message = fmt.Sprintf("%s must be a valid email", fieldError.Field())
			default:
				validationError.Message = fmt.Sprintf("%s failed validation for '%s'", fieldError.Field(), fieldError.Tag())
			}

			validationErrors = append(validationErrors, validationError)
		}
	}

	return validationErrors
}

func bind(r *http.Request, v interface{}) error {
	if r.Body == nil && r.Method != "GET" {
		return fmt.Errorf("request body is required")
	}

	contentType := r.Header.Get("Content-Type")

	switch {
	case strings.Contains(contentType, "application/json"):
		return bindJSON(r, v)
	case strings.Contains(contentType, "application/x-www-form-urlencoded"):
		return bindForm(r, v)
	case strings.Contains(contentType, "multipart/form-data"):
		return bindForm(r, v)
	default:
		return fmt.Errorf("unsupported content-type: %s", contentType)
	}
}

func bindJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	return validate.Struct(v)
}

func bindForm(r *http.Request, v interface{}) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("failed to parse form: %w", err)
	}

	if err := mapFormToStruct(r, v); err != nil {
		return err
	}

	return validate.Struct(v)
}

func mapFormToStruct(r *http.Request, v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("v must be a pointer to struct")
	}

	rv = rv.Elem()
	rt := rv.Type()

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		fieldValue := rv.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		// Get form tag value
		formTag := field.Tag.Get("form")
		if formTag == "" {
			continue
		}

		// Get value from form
		formValue := r.FormValue(formTag)
		if formValue == "" {
			continue
		}

		// Set value based on field type
		switch fieldValue.Kind() {
		case reflect.String:
			fieldValue.SetString(formValue)
		case reflect.Int, reflect.Int32, reflect.Int64:
			if intVal, err := strconv.ParseInt(formValue, 10, 64); err == nil {
				fieldValue.SetInt(intVal)
			}
		case reflect.Float32, reflect.Float64:
			if floatVal, err := strconv.ParseFloat(formValue, 64); err == nil {
				fieldValue.SetFloat(floatVal)
			}
		case reflect.Bool:
			if boolVal, err := strconv.ParseBool(formValue); err == nil {
				fieldValue.SetBool(boolVal)
			}
		}
	}

	return validate.Struct(v)
}

func formatValidationErrorsHTML(errs validator.ValidationErrors) string {
	var html strings.Builder
	html.WriteString(`<div class="rounded-md bg-red-50 p-4">`)
	html.WriteString(`<div class="flex"><div class="ml-3">`)
	html.WriteString(`<h3 class="text-sm font-medium text-red-800">Validation errors:</h3>`)
	html.WriteString(`<div class="mt-2 text-sm text-red-700"><ul class="list-disc space-y-1 pl-5">`)
	
	for _, fieldError := range errs {
		var msg string
		switch fieldError.Tag() {
		case "required":
			msg = fmt.Sprintf("%s is required", fieldError.Field())
		case "min":
			msg = fmt.Sprintf("%s must be at least %s characters", fieldError.Field(), fieldError.Param())
		case "max":
			msg = fmt.Sprintf("%s must be at most %s characters", fieldError.Field(), fieldError.Param())
		case "email":
			msg = fmt.Sprintf("%s must be a valid email", fieldError.Field())
		default:
			msg = fmt.Sprintf("%s failed validation", fieldError.Field())
		}
		html.WriteString(fmt.Sprintf(`<li>%s</li>`, msg))
	}
	
	html.WriteString(`</ul></div></div></div></div>`)
	return html.String()
}

// Regular error response helper (for non-validation errors)
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func writeJSON(w http.ResponseWriter, data interface{}, statusCode int) error {
	// Encode to buffer first to catch errors before writing headers
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		// Now we can send an error response since we haven't written headers yet
		writeErrorJSON(w, "Failed to encode response", http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	buf.WriteTo(w)
	return nil
}

func writeErrorJSON(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
		Code:    statusCode,
	}

	json.NewEncoder(w).Encode(response)
}
