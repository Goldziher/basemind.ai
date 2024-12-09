package apierror

import (
	"context"
	"fmt"
	"github.com/basemind-ai/monorepo/shared/go/serialization"
	"net/http"
	"strings"
)

type extraContextKey int

// ExtraContextKey - the key used to store extra information in the context.
const ExtraContextKey extraContextKey = iota

// APIError - represents an API error.
// Its a type that can render itself into an HTTP response.
type APIError struct {
	Extra      any    `json:"extra,omitempty"`
	Message    string `json:"message"`
	StatusText string `json:"status"`
	StatusCode int    `json:"statusCode"`
}

// Render - renders the APIError into an HTTP response.
func (e *APIError) Render(w http.ResponseWriter) {
	serialization.RenderJSONResponse(w, e.StatusCode, e)
}

// RenderWithExtra - renders the APIError into an HTTP response,
// serializing any value set for the ExtraContextKey in the context as part of the JSON response.
// This is useful for returning additional information about an error from nested operations.
func (e *APIError) RenderWithExtra(ctx context.Context, w http.ResponseWriter) {
	e.Extra = ctx.Value(ExtraContextKey)
	serialization.RenderJSONResponse(w, e.StatusCode, e)
}

// Error - returns the error message.
// This receiver ensures that APIError fulfills the Error interface.
func (e *APIError) Error() string {
	return fmt.Sprintf("status: %d, message: %s", e.StatusCode, e.Message)
}

// New - creates a new APIError.
func New(statusCode int, messages ...string) *APIError {
	var message string
	if len(messages) > 0 {
		message = strings.Join(messages, ", ")
	} else {
		message = http.StatusText(statusCode)
	}

	return &APIError{
		StatusCode: statusCode,
		Message:    message,
		StatusText: http.StatusText(statusCode),
	}
}

// NotFound - creates a new APIError with status code 404.
func NotFound(messages ...string) *APIError {
	return New(http.StatusNotFound, messages...)
}

// BadRequest - creates a new APIError with status code 400.
func BadRequest(messages ...string) *APIError {
	return New(http.StatusBadRequest, messages...)
}

// Unauthorized - creates a new APIError with status code 401.
func Unauthorized(messages ...string) *APIError {
	return New(http.StatusUnauthorized, messages...)
}

// Forbidden - creates a new APIError with status code 403.
func Forbidden(messages ...string) *APIError {
	return New(http.StatusForbidden, messages...)
}

// UnprocessableContent - creates a new APIError with status code 422.
func UnprocessableContent(messages ...string) *APIError {
	return New(http.StatusUnprocessableEntity, messages...)
}

// InternalServerError - creates a new APIError with status code 500.
func InternalServerError(messages ...string) *APIError {
	return New(http.StatusInternalServerError, messages...)
}
