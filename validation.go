package main

import "fmt"

// ValidateCoverageConfig validates coverage configuration values
func ValidateCoverageConfig(min, max float64) error {
	if min < 0 || min > 100 {
		return NewValidationError("coverage.min", min, "must be between 0 and 100")
	}
	if max < 0 || max > 100 {
		return NewValidationError("coverage.max", max, "must be between 0 and 100")
	}
	if min > max {
		return NewValidationError("coverage", fmt.Sprintf("min=%v, max=%v", min, max), "min cannot be greater than max")
	}
	return nil
}

// ValidateFormat validates the output format
func ValidateFormat(format string) error {
	if format != "table" && format != "json" {
		return NewValidationError("format", format, "must be 'table' or 'json'")
	}
	return nil
}

// ValidateThreshold validates the coverage threshold
func ValidateThreshold(threshold float64) error {
	if threshold < 0 || threshold > 100 {
		return NewValidationError("threshold", threshold, "must be between 0 and 100")
	}
	return nil
}
