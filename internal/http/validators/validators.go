package validators

import (
	"fmt"
	"strings"
	"unicode"
)

func ValidateNonEmpty(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s cannot be empty", field)
	}
	return nil
}

func ValidateMaxLength(field, value string, maxLen int) error {
	if len(value) > maxLen {
		return fmt.Errorf("%s cannot exceed %d characters", field, maxLen)
	}
	return nil
}

func ValidateMinLength(field, value string, minLen int) error {
	if len(strings.TrimSpace(value)) < minLen {
		return fmt.Errorf("%s must be at least %d characters", field, minLen)
	}
	return nil
}

func ValidateAlphanumeric(field, value string) error {
	for _, r := range value {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-' {
			return fmt.Errorf("%s can only contain letters, numbers, underscores, and hyphens", field)
		}
	}
	return nil
}

func ValidateSliceNotEmpty(field string, slice any) error {
	switch v := slice.(type) {
	case []any:
		if len(v) == 0 {
			return fmt.Errorf("%s cannot be empty", field)
		}
	case []string:
		if len(v) == 0 {
			return fmt.Errorf("%s cannot be empty", field)
		}
	}
	return nil
}
