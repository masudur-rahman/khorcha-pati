package dberr

import (
	"errors"
	"fmt"
	"strings"
)

// RequirementMissing represents an error when a required field or parameter is missing.
type RequirementMissing struct {
	Message string
}

func (e RequirementMissing) Error() string {
	return e.Message
}

func NewRequirementMissing(msg string) RequirementMissing {
	return RequirementMissing{Message: msg}
}

// DataNotFound is returned when a query returns no results.
// Deprecated: Use ErrNotFound instead.
var DataNotFound = errors.New("data not found")

// Typed errors for common database operations.
var (
	// ErrNotFound is returned when a record is not found.
	ErrNotFound = errors.New("styx: record not found")

	// ErrDuplicateEntry is returned when a unique constraint is violated.
	ErrDuplicateEntry = errors.New("styx: duplicate entry")

	// ErrInvalidQuery is returned when a query is malformed or invalid.
	ErrInvalidQuery = errors.New("styx: invalid query")

	// ErrMissingWhereClause is returned when an update or delete is attempted without a WHERE clause.
	ErrMissingWhereClause = errors.New("styx: WHERE clause is required for update/delete operations")

	// ErrTransactionNotStarted is returned when commit or rollback is called without an active transaction.
	ErrTransactionNotStarted = errors.New("styx: no transaction in progress")

	// ErrTransactionAlreadyStarted is returned when a new transaction is attempted while one is already active.
	ErrTransactionAlreadyStarted = errors.New("styx: session already in progress")

	// ErrInvalidEntityName is returned when an entity/table name is empty.
	ErrInvalidEntityName = errors.New("styx: entity name cannot be empty")

	// ErrInvalidID is returned when a required document ID is missing.
	ErrInvalidID = errors.New("styx: ID cannot be empty")

	// ErrConnectionFailed is returned when a database connection cannot be established.
	ErrConnectionFailed = errors.New("styx: connection failed")

	// ErrValidationFailed is returned when validation rules are not satisfied.
	ErrValidationFailed = errors.New("styx: validation failed")
)

// ValidationError represents a collection of validation errors.
type ValidationError struct {
	FieldErrors map[string][]string
}

func (e *ValidationError) Error() string {
	var msgs []string
	for field, errors := range e.FieldErrors {
		for _, err := range errors {
			msgs = append(msgs, fmt.Sprintf("%s: %s", field, err))
		}
	}
	return fmt.Sprintf("styx: validation failed: %s", strings.Join(msgs, "; "))
}

// NewValidationError creates a new ValidationError.
func NewValidationError(fieldErrors map[string][]string) *ValidationError {
	return &ValidationError{FieldErrors: fieldErrors}
}

// IsNotFound checks if an error indicates a record not found.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrNotFound) || errors.Is(err, DataNotFound)
}

// IsDuplicate checks if an error indicates a duplicate entry.
func IsDuplicate(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrDuplicateEntry)
}

// IsValidationError checks if an error is a validation error.
func IsValidationError(err error) bool {
	if err == nil {
		return false
	}
	var ve *ValidationError
	return errors.As(err, &ve)
}
