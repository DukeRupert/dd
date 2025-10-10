package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidationError represents a single field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// ValidationErrorResponse is the response for validation errors
type ValidationErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Details []ValidationError `json:"details"`
}

// ErrorResponse is the response for general errors
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// bind binds request data to a struct based on content type
func (h *Handler) bind(r *http.Request, v interface{}) error {
	if r.Body == nil && r.Method != "GET" {
		return fmt.Errorf("request body is required")
	}

	contentType := r.Header.Get("Content-Type")

	switch {
	case strings.Contains(contentType, "application/json"):
		return h.bindJSON(r, v)
	case strings.Contains(contentType, "application/x-www-form-urlencoded"):
		return h.bindForm(r, v)
	case strings.Contains(contentType, "multipart/form-data"):
		return h.bindForm(r, v)
	default:
		return fmt.Errorf("unsupported content-type: %s", contentType)
	}
}

// bindJSON binds JSON request body to struct
func (h *Handler) bindJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	return h.validate.Struct(v)
}

// bindForm binds form data to struct
func (h *Handler) bindForm(r *http.Request, v interface{}) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("failed to parse form: %w", err)
	}

	if err := h.mapFormToStruct(r, v); err != nil {
		return err
	}

	return h.validate.Struct(v)
}

// mapFormToStruct maps form values to struct fields
func (h *Handler) mapFormToStruct(r *http.Request, v interface{}) error {
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

	return nil
}

// writeJSON writes JSON response
func (h *Handler) writeJSON(w http.ResponseWriter, data interface{}, statusCode int) error {
	// Encode to buffer first to catch errors before writing headers
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		// Now we can send an error response since we haven't written headers yet
		h.writeErrorJSON(w, "Failed to encode response", http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	buf.WriteTo(w)
	return nil
}

// writeErrorJSON writes JSON error response
func (h *Handler) writeErrorJSON(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
		Code:    statusCode,
	}

	json.NewEncoder(w).Encode(response)
}

// getValidationErrors converts validator errors to friendly messages
func (h *Handler) getValidationErrors(err error) []ValidationError {
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

// formatValidationErrorsHTML formats validation errors as HTML
func (h *Handler) formatValidationErrorsHTML(errs validator.ValidationErrors) string {
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
			msg = fmt.Sprintf("%s must be at least %s characters", fieldError.Field(), fieldError.Param())
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
