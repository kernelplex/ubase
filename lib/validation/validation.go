package validation

import (
	"fmt"
	"strings"
)

type ValidationIssues struct {
	Issues []ValidationIssue `json:"issues"`
}

type ValidationIssue struct {
	Field string   `json:"field"`
	Error []string `json:"error"`
}

const (
	ErrEmailRequired          = "email is required"
	ErrEmailInvalid           = "email must be valid"
	ErrPasswordRequired       = "password is required"
	ErrPasswordEmpty          = "password cannot be empty if provided"
	ErrPasswordMinLength      = "must be at least 8 characters"
	ErrPasswordUppercase      = "must contain an uppercase letter"
	ErrPasswordLowercase      = "must contain a lowercase letter"
	ErrPasswordNumber         = "must contain a number"
	ErrPasswordSpecialChar    = "must contain a special character"
	ErrFieldRequiredTemplate  = "%s is required"
	ErrFieldEmptyTemplate     = "%s cannot be empty if provided"
	ErrFieldMinLengthTemplate = "%s must be at least %d characters"
)

func formatFieldName(fieldName string) string {
	switch fieldName {
	case "FirstName":
		return "first name"
	case "LastName":
		return "last name"
	case "DisplayName":
		return "display name"
	default:
		return strings.ToLower(fieldName)
	}
}

func ValidateEmail(email string) []ValidationIssue {
	var issues []ValidationIssue
	if email == "" {
		issues = append(issues, ValidationIssue{
			Field: "Email",
			Error: []string{ErrEmailRequired},
		})
	} else if !strings.Contains(email, "@") {
		issues = append(issues, ValidationIssue{
			Field: "Email",
			Error: []string{ErrEmailInvalid},
		})
	}
	return issues
}

func ValidateField(fieldName string, value string, required bool, minLength int) []ValidationIssue {
	var issues []ValidationIssue
	if value == "" {
		if required {
			issues = append(issues, ValidationIssue{
				Field: fieldName,
				Error: []string{fmt.Sprintf(ErrFieldRequiredTemplate, formatFieldName(fieldName))},
			})
		}
	} else if minLength > 0 && len(value) < minLength {
		issues = append(issues, ValidationIssue{
			Field: fieldName,
			Error: []string{fmt.Sprintf(ErrFieldMinLengthTemplate, formatFieldName(fieldName), minLength)},
		})
	}
	return issues
}

func ValidatePasswordComplexity(password string) []string {
	var errors []string
	if len(password) < 8 {
		errors = append(errors, ErrPasswordMinLength)
	}
	if !strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		errors = append(errors, ErrPasswordUppercase)
	}
	if !strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz") {
		errors = append(errors, ErrPasswordLowercase)
	}
	if !strings.ContainsAny(password, "0123456789") {
		errors = append(errors, ErrPasswordNumber)
	}
	if !strings.ContainsAny(password, "!@#$%^&*()-_=+[]{}|;:'\",.<>/?") {
		errors = append(errors, ErrPasswordSpecialChar)
	}
	return errors
}

func ValidateOptionalField(fieldName string, value *string, minLength int) []ValidationIssue {
	var issues []ValidationIssue
	if value != nil {
		if *value == "" {
			issues = append(issues, ValidationIssue{
				Field: fieldName,
				Error: []string{fmt.Sprintf(ErrFieldEmptyTemplate, formatFieldName(fieldName))},
			})
		} else if minLength > 0 && len(*value) < minLength {
			issues = append(issues, ValidationIssue{
				Field: fieldName,
				Error: []string{fmt.Sprintf("%s must be at least %d characters if provided", strings.ToLower(fieldName), minLength)},
			})
		}
	}
	return issues
}
