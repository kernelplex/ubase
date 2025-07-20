package ubvalidation

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

type ValidationTracker struct {
	issueMap map[string]*ValidationIssue
	IsValid  bool
}

func NewValidationTracker() *ValidationTracker {
	return &ValidationTracker{
		issueMap: make(map[string]*ValidationIssue),
		IsValid:  true,
	}
}

func (t *ValidationTracker) AddIssue(fieldName string, error string) {

	existingIssue, ok := t.issueMap[fieldName]
	if !ok {
		existingIssue = &ValidationIssue{
			Field: fieldName,
			Error: []string{error},
		}
		t.issueMap[fieldName] = existingIssue
		t.IsValid = false
	}
	existingIssue.Error = append(existingIssue.Error, error)
}

func (t *ValidationTracker) Valid() (bool, []ValidationIssue) {
	issues := make([]ValidationIssue, 0)
	isValid := true
	for _, issue := range t.issueMap {
		issues = append(issues, *issue)
		isValid = false
	}
	return isValid, issues
}

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

func (t *ValidationTracker) ValidateEmail(fieldName string, email string) {
	if email == "" {
		t.AddIssue(fieldName, ErrEmailRequired)
	} else if !strings.Contains(email, "@") {
		t.AddIssue(fieldName, ErrEmailInvalid)
	}
}

func (t *ValidationTracker) ValidateField(fieldName string, value string, required bool, minLength int) {
	if value == "" {
		if required {
			t.AddIssue(fieldName, fmt.Sprintf(ErrFieldRequiredTemplate, formatFieldName(fieldName)))
		}
	} else if minLength > 0 && len(value) < minLength {
		t.AddIssue(fieldName, fmt.Sprintf(ErrFieldMinLengthTemplate, formatFieldName(fieldName), minLength))
	}
}

func (t *ValidationTracker) ValidatePasswordComplexity(fieldName string, password string) {
	if len(password) < 8 {
		t.AddIssue(fieldName, ErrPasswordMinLength)
	}
	if !strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		t.AddIssue(fieldName, ErrPasswordUppercase)
	}
	if !strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz") {
		t.AddIssue(fieldName, ErrPasswordLowercase)
	}
	if !strings.ContainsAny(password, "0123456789") {
		t.AddIssue(fieldName, ErrPasswordNumber)
	}
	if !strings.ContainsAny(password, "!@#$%^&*()-_=+[]{}|;:'\",.<>/?") {
		t.AddIssue(fieldName, ErrPasswordSpecialChar)
	}
}

func (t *ValidationTracker) ValidateOptionalField(fieldName string, value *string, minLength int) {
	if value != nil {
		if *value == "" {
			t.AddIssue(fieldName, fmt.Sprintf(ErrFieldEmptyTemplate, formatFieldName(fieldName)))
		} else if minLength > 0 && len(*value) < minLength {
			t.AddIssue(fieldName, fmt.Sprintf("%s must be at least %d characters if provided", strings.ToLower(fieldName), minLength))
		}
	}
}

func (t *ValidationTracker) ValidateIntMinValue(fieldName string, value int64, minValue int64) {

	if value < minValue {
		t.AddIssue(fieldName, fmt.Sprintf("%s must be at least %d", strings.ToLower(fieldName), minValue))
	}
}

func (t *ValidationTracker) ValidateIntMaxValue(fieldName string, value int64, maxValue int64) {

	if value > maxValue {
		t.AddIssue(fieldName, fmt.Sprintf("%s must be at most %d", strings.ToLower(fieldName), maxValue))
	}
}
