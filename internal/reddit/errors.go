package reddit

import (
	"fmt"
	"time"
)

// MissingInputError is returned when a required parameter is not provided.
type MissingInputError struct {
	input string
}

func (e *MissingInputError) Error() string {
	return "missing required input: " + e.input
}

func NewMissingInputError(input string) *MissingInputError {
	return &MissingInputError{input: input}
}

// NotInitializedError is returned when the client has not been initialized by the constructor.
type NotInitializedError struct{}

func (e *NotInitializedError) Error() string {
	return "client uninitialized, use constructor"
}

func NewNotInitializedError() *NotInitializedError {
	return &NotInitializedError{}
}

// NotAuthenticatedError is returned when a request is made to a protected endpoint but no bearer
// token has been generated.
type NotAuthenticatedError struct{}

func (e *NotAuthenticatedError) Error() string {
	return "not authenticated"
}

func NewNotAuthenticatedError() *NotAuthenticatedError {
	return &NotAuthenticatedError{}
}

// RateLimitExceededError is returned when an HTTP request is denied with http.StatusTooManyRequests(429).
type RateLimitExceededError struct {
	ResetsIn time.Duration
}

func (e *RateLimitExceededError) Error() string {
	return fmt.Sprintf("rate limit exceeded, resets in %s", e.ResetsIn)
}

func NewRateLimitExceededError(resetsIn time.Duration) *RateLimitExceededError {
	return &RateLimitExceededError{ResetsIn: resetsIn}
}

// UnexpectedStatusError is returned when an HTTP response status code did not indicate success or throttling.
type UnexpectedStatusError struct {
	Method, URL string
	Status      int
}

func (u *UnexpectedStatusError) Error() string {
	return fmt.Sprintf("unexpected status code %d (%s %s)", u.Status, u.Method, u.URL)
}

func NewUnexpectedStatusError(method string, url string, status int) *UnexpectedStatusError {
	return &UnexpectedStatusError{Method: method, URL: url, Status: status}
}
