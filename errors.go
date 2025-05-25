package main

import (
	"errors"
	"fmt"
)

// Error types
var (
	// Configuration errors
	ErrNoInput        = errors.New("coverprofile is required")
	ErrInvalidFormat  = errors.New("invalid output format")
	ErrConfigNotFound = errors.New("configuration file not found")
	ErrInvalidConfig  = errors.New("invalid configuration")

	// Validation errors
	ErrInvalidMinCoverage = errors.New("min must be between 0 and 100")
	ErrInvalidMaxCoverage = errors.New("max must be between 0 and 100")
	ErrMinGreaterThanMax  = errors.New("min cannot be greater than max")

	// Parse errors
	ErrParseCoverage = errors.New("failed to parse coverage profile")
)

// ConfigError represents a configuration-related error
type ConfigError struct {
	Field string
	Value interface{}
	Err   error
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("config error in field '%s' with value '%v': %v", e.Field, e.Value, e.Err)
}

func (e *ConfigError) Unwrap() error {
	return e.Err
}

// ValidationError represents a validation-related error
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s (field: %s, value: %v)", e.Message, e.Field, e.Value)
}

// ParseError represents a parsing-related error
type ParseError struct {
	File string
	Err  error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error in file '%s': %v", e.File, e.Err)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

// NewConfigError creates a new ConfigError
func NewConfigError(field string, value interface{}, err error) error {
	return &ConfigError{
		Field: field,
		Value: value,
		Err:   err,
	}
}

// NewValidationError creates a new ValidationError
func NewValidationError(field string, value interface{}, message string) error {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}

// NewParseError creates a new ParseError
func NewParseError(file string, err error) error {
	return &ParseError{
		File: file,
		Err:  err,
	}
}

