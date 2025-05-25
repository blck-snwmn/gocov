package main

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestConfigError(t *testing.T) {
	baseErr := errors.New("base error")
	err := NewConfigError("testField", "testValue", baseErr)

	configErr, ok := err.(*ConfigError)
	if !ok {
		t.Fatal("NewConfigError did not return *ConfigError")
	}

	// Test Error() method
	errStr := configErr.Error()
	if !strings.Contains(errStr, "testField") {
		t.Errorf("Error string should contain field name, got: %s", errStr)
	}
	if !strings.Contains(errStr, "testValue") {
		t.Errorf("Error string should contain value, got: %s", errStr)
	}
	if !strings.Contains(errStr, "base error") {
		t.Errorf("Error string should contain base error, got: %s", errStr)
	}

	// Test Unwrap() method
	unwrapped := configErr.Unwrap()
	if unwrapped != baseErr {
		t.Errorf("Unwrap() should return base error, got: %v", unwrapped)
	}

	// Test with errors.Is
	if !errors.Is(err, baseErr) {
		t.Error("errors.Is should match base error")
	}
}

func TestValidationError(t *testing.T) {
	err := NewValidationError("coverage.min", -10, "must be between 0 and 100")

	validationErr, ok := err.(*ValidationError)
	if !ok {
		t.Fatal("NewValidationError did not return *ValidationError")
	}

	// Test Error() method
	errStr := validationErr.Error()
	if !strings.Contains(errStr, "coverage.min") {
		t.Errorf("Error string should contain field name, got: %s", errStr)
	}
	if !strings.Contains(errStr, "-10") {
		t.Errorf("Error string should contain value, got: %s", errStr)
	}
	if !strings.Contains(errStr, "must be between 0 and 100") {
		t.Errorf("Error string should contain message, got: %s", errStr)
	}
}

func TestParseError(t *testing.T) {
	baseErr := errors.New("invalid format")
	err := NewParseError("coverage.out", baseErr)

	parseErr, ok := err.(*ParseError)
	if !ok {
		t.Fatal("NewParseError did not return *ParseError")
	}

	// Test Error() method
	errStr := parseErr.Error()
	if !strings.Contains(errStr, "coverage.out") {
		t.Errorf("Error string should contain file name, got: %s", errStr)
	}
	if !strings.Contains(errStr, "invalid format") {
		t.Errorf("Error string should contain base error, got: %s", errStr)
	}

	// Test Unwrap() method
	unwrapped := parseErr.Unwrap()
	if unwrapped != baseErr {
		t.Errorf("Unwrap() should return base error, got: %v", unwrapped)
	}

	// Test with errors.Is
	if !errors.Is(err, baseErr) {
		t.Error("errors.Is should match base error")
	}
}

func TestErrorConstants(t *testing.T) {
	// Test that error constants are properly defined
	tests := []struct {
		err      error
		contains string
	}{
		{ErrNoInput, "coverprofile is required"},
		{ErrInvalidFormat, "invalid output format"},
		{ErrConfigNotFound, "configuration file not found"},
		{ErrInvalidConfig, "invalid configuration"},
		{ErrInvalidMinCoverage, "min must be between 0 and 100"},
		{ErrInvalidMaxCoverage, "max must be between 0 and 100"},
		{ErrMinGreaterThanMax, "min cannot be greater than max"},
		{ErrParseCoverage, "failed to parse coverage profile"},
	}

	for _, tt := range tests {
		t.Run(tt.err.Error(), func(t *testing.T) {
			if !strings.Contains(tt.err.Error(), tt.contains) {
				t.Errorf("Error %v should contain '%s', got: %s", tt.err, tt.contains, tt.err.Error())
			}
		})
	}
}

func TestErrorTypes(t *testing.T) {
	// Test that custom error types can be used with errors.As
	t.Run("ConfigError with errors.As", func(t *testing.T) {
		err := fmt.Errorf("wrapped: %w", NewConfigError("field", "value", ErrInvalidConfig))

		var configErr *ConfigError
		if !errors.As(err, &configErr) {
			t.Error("errors.As should work with wrapped ConfigError")
		}

		if configErr.Field != "field" {
			t.Errorf("Expected field 'field', got %s", configErr.Field)
		}
	})

	t.Run("ValidationError with errors.As", func(t *testing.T) {
		err := fmt.Errorf("wrapped: %w", NewValidationError("field", 123, "test message"))

		var validationErr *ValidationError
		if !errors.As(err, &validationErr) {
			t.Error("errors.As should work with wrapped ValidationError")
		}

		if validationErr.Field != "field" {
			t.Errorf("Expected field 'field', got %s", validationErr.Field)
		}
		if validationErr.Value != 123 {
			t.Errorf("Expected value 123, got %v", validationErr.Value)
		}
	})

	t.Run("ParseError with errors.As", func(t *testing.T) {
		err := fmt.Errorf("wrapped: %w", NewParseError("test.out", ErrParseCoverage))

		var parseErr *ParseError
		if !errors.As(err, &parseErr) {
			t.Error("errors.As should work with wrapped ParseError")
		}

		if parseErr.File != "test.out" {
			t.Errorf("Expected file 'test.out', got %s", parseErr.File)
		}
	})
}
