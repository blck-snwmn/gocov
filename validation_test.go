package main

import (
	"errors"
	"testing"
)

func TestValidateCoverageConfig(t *testing.T) {
	tests := []struct {
		name    string
		min     float64
		max     float64
		wantErr bool
	}{
		{
			name:    "valid range",
			min:     0,
			max:     100,
			wantErr: false,
		},
		{
			name:    "valid custom range",
			min:     50,
			max:     80,
			wantErr: false,
		},
		{
			name:    "min below 0",
			min:     -10,
			max:     100,
			wantErr: true,
		},
		{
			name:    "min above 100",
			min:     110,
			max:     100,
			wantErr: true,
		},
		{
			name:    "max below 0",
			min:     0,
			max:     -10,
			wantErr: true,
		},
		{
			name:    "max above 100",
			min:     0,
			max:     110,
			wantErr: true,
		},
		{
			name:    "min greater than max",
			min:     80,
			max:     50,
			wantErr: true,
		},
		{
			name:    "equal min and max",
			min:     50,
			max:     50,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCoverageConfig(tt.min, tt.max)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCoverageConfig(%v, %v) error = %v, wantErr %v", tt.min, tt.max, err, tt.wantErr)
			}

			if err != nil {
				var validationErr *ValidationError
				if !errors.As(err, &validationErr) {
					t.Errorf("Expected ValidationError, got %T", err)
				}
			}
		})
	}
}

func TestValidateFormat(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		wantErr bool
	}{
		{
			name:    "valid table format",
			format:  "table",
			wantErr: false,
		},
		{
			name:    "valid json format",
			format:  "json",
			wantErr: false,
		},
		{
			name:    "invalid xml format",
			format:  "xml",
			wantErr: true,
		},
		{
			name:    "invalid empty format",
			format:  "",
			wantErr: true,
		},
		{
			name:    "invalid csv format",
			format:  "csv",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFormat(tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFormat(%v) error = %v, wantErr %v", tt.format, err, tt.wantErr)
			}

			if err != nil {
				var validationErr *ValidationError
				if !errors.As(err, &validationErr) {
					t.Errorf("Expected ValidationError, got %T", err)
				}
			}
		})
	}
}
