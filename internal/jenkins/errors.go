package jenkins

import (
	"errors"
	"fmt"
)

// ErrorCode represents standardized error codes for Jenkins operations
type ErrorCode string

const (
	// ErrorCodeAuthFailed indicates authentication failure
	ErrorCodeAuthFailed ErrorCode = "AUTH_FAILED"
	
	// ErrorCodeNotFound indicates a resource was not found
	ErrorCodeNotFound ErrorCode = "NOT_FOUND"
	
	// ErrorCodeInvalidInput indicates invalid input parameters
	ErrorCodeInvalidInput ErrorCode = "INVALID_INPUT"
	
	// ErrorCodeNetworkError indicates a network connectivity issue
	ErrorCodeNetworkError ErrorCode = "NETWORK_ERROR"
	
	// ErrorCodeTimeout indicates an operation timeout
	ErrorCodeTimeout ErrorCode = "TIMEOUT"
	
	// ErrorCodePermissionDenied indicates insufficient permissions
	ErrorCodePermissionDenied ErrorCode = "PERMISSION_DENIED"
	
	// ErrorCodeJenkinsError indicates a Jenkins API error
	ErrorCodeJenkinsError ErrorCode = "JENKINS_ERROR"
	
	// ErrorCodeInternalError indicates an unexpected server error
	ErrorCodeInternalError ErrorCode = "INTERNAL_ERROR"
)

// ErrorResponse represents a structured error response
type ErrorResponse struct {
	Code    ErrorCode              `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Error implements the error interface for ErrorResponse
func (e *ErrorResponse) Error() string {
	if len(e.Details) > 0 {
		return fmt.Sprintf("%s: %s (details: %v)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewError creates a new ErrorResponse with the given code and message
func NewError(code ErrorCode, message string) *ErrorResponse {
	return &ErrorResponse{
		Code:    code,
		Message: message,
		Details: nil,
	}
}

// NewErrorWithDetails creates a new ErrorResponse with code, message, and details
func NewErrorWithDetails(code ErrorCode, message string, details map[string]interface{}) *ErrorResponse {
	return &ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// WrapError wraps an existing error with an ErrorResponse
func WrapError(code ErrorCode, message string, err error) error {
	if err == nil {
		return NewError(code, message)
	}
	
	details := map[string]interface{}{
		"underlying_error": err.Error(),
	}
	
	return &ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// Helper functions for common error scenarios

// NewAuthError creates an authentication failure error
func NewAuthError(message string) *ErrorResponse {
	return NewError(ErrorCodeAuthFailed, message)
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *ErrorResponse {
	return NewError(ErrorCodeNotFound, fmt.Sprintf("%s not found", resource))
}

// NewInvalidInputError creates an invalid input error
func NewInvalidInputError(message string) *ErrorResponse {
	return NewError(ErrorCodeInvalidInput, message)
}

// NewNetworkError creates a network error
func NewNetworkError(message string) *ErrorResponse {
	return NewError(ErrorCodeNetworkError, message)
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(operation string) *ErrorResponse {
	return NewError(ErrorCodeTimeout, fmt.Sprintf("operation timed out: %s", operation))
}

// NewPermissionDeniedError creates a permission denied error
func NewPermissionDeniedError(resource string) *ErrorResponse {
	return NewError(ErrorCodePermissionDenied, fmt.Sprintf("permission denied: %s", resource))
}

// NewJenkinsError creates a Jenkins API error
func NewJenkinsError(message string) *ErrorResponse {
	return NewError(ErrorCodeJenkinsError, message)
}

// NewInternalError creates an internal error
func NewInternalError(message string) *ErrorResponse {
	return NewError(ErrorCodeInternalError, message)
}

// IsErrorCode checks if an error has a specific error code
func IsErrorCode(err error, code ErrorCode) bool {
	var errResp *ErrorResponse
	if errors.As(err, &errResp) {
		return errResp.Code == code
	}
	return false
}

// GetErrorCode extracts the error code from an error, if it's an ErrorResponse
func GetErrorCode(err error) (ErrorCode, bool) {
	var errResp *ErrorResponse
	if errors.As(err, &errResp) {
		return errResp.Code, true
	}
	return "", false
}
