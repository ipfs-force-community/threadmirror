package errors

import (
	"fmt"
	"net/http"
)

// Unified error definition & helpers for a Gin-based JSON API
//
// Error Code Ranges:
// - 0: Success
// - 1000-9999: System/Framework-level errors (HTTP mapping, validation, auth, etc.)
// - 10000+: Business/Domain-specific errors

// ErrorCode represents an application-specific error code with a message
type ErrorCode struct {
	Code    uint16
	Message string
}

func NewErrorCode(code uint16, message string) ErrorCode {
	return ErrorCode{
		Code:    code,
		Message: message,
	}
}

var codeSet = make(map[uint16]struct{})

// checkCode is a helper function to check if a code is already defined
func CheckCode(code uint16) uint16 {
	if _, ok := codeSet[code]; ok {
		panic(fmt.Sprintf("code %d already exists", code))
	}
	codeSet[code] = struct{}{}
	return code
}

// Predefined error codes
var (
	ErrCodeOk = NewErrorCode(CheckCode(0), "OK")
	// Common errors (4001-5999)
	ErrCodeBadRequest         = NewErrorCode(CheckCode(4001), "bad request")
	ErrCodeInvalidRequestBody = NewErrorCode(CheckCode(4002), "invalid request body")
	ErrCodeNotFound           = NewErrorCode(CheckCode(4003), "resource not found")
	ErrCodeForbidden          = NewErrorCode(CheckCode(4004), "access forbidden")
	ErrCodeConflict           = NewErrorCode(CheckCode(4005), "resource conflict")
	ErrCodeInternalError      = NewErrorCode(CheckCode(5000), "Internal error")
	// Business errors (10000+)
	ErrCodeBusinessError = NewErrorCode(CheckCode(10000), "Business error")
)

// BadRequest creates a new API error for bad request.
func BadRequest(cause error) *APIError {
	return NewAPIError(ErrCodeBadRequest, http.StatusBadRequest, cause)
}

// InvalidRequestBody creates a new API error for invalid request body.
func InvalidRequestBody(cause error) *APIError {
	return BadRequest(cause).WithCode(ErrCodeInvalidRequestBody)
}

// NotFound creates a new API error for resource not found.
func NotFound(cause error) *APIError {
	return NewAPIError(ErrCodeNotFound, http.StatusNotFound, cause)
}

// Forbidden creates a new API error for access forbidden.
func Forbidden(cause error) *APIError {
	return NewAPIError(ErrCodeForbidden, http.StatusForbidden, cause)
}

// Conflict creates a new API error for resource conflict.
func Conflict(cause error) *APIError {
	return NewAPIError(ErrCodeConflict, http.StatusConflict, cause)
}

// InternalServerError creates a new API error for internal server error.
func InternalServerError(cause error) *APIError {
	return NewAPIError(ErrCodeInternalError, http.StatusInternalServerError, cause)
}

// NewAPIError constructs a new API error.
func NewAPIError(code ErrorCode, status int, cause error) *APIError {
	return &APIError{
		ErrorCode: code,
		Status:    status,
		Cause:     cause,
	}
}

// APIError represents a standardized API error response
type APIError struct {
	ErrorCode ErrorCode
	// Status is the HTTP status code of the response that caused the error.
	Status int
	// Cause is the underlying error that caused the api error.
	Cause error
}

func (a *APIError) WithCode(errorCode ErrorCode) *APIError {
	return &APIError{
		ErrorCode: errorCode,
		Status:    a.Status,
		Cause:     a.Cause,
	}
}

func (a *APIError) WithCause(cause error) *APIError {
	return &APIError{
		ErrorCode: a.ErrorCode,
		Status:    a.Status,
		Cause:     cause,
	}
}

// Unwrap returns the underlying error. This also makes the error compatible
// with errors.As and errors.Is.
func (a *APIError) Unwrap() error {
	if a == nil {
		return nil
	}
	return a.Cause
}

// Error returns the API error's message.
func (a *APIError) Error() string {
	if a == nil {
		return ""
	}

	// If it's a success code, just return the cause error
	if a.ErrorCode == ErrCodeOk {
		if a.Cause != nil {
			return a.Cause.Error()
		}
		return a.ErrorCode.Message
	}

	// Build error message with code information
	var msg string
	if a.Cause != nil {
		msg = fmt.Sprintf("%d: %s", a.ErrorCode.Code, a.Cause.Error())
	} else {
		msg = fmt.Sprintf("%d: %s", a.ErrorCode.Code, a.ErrorCode.Message)
	}

	return msg
}
