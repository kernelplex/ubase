package domain

import (
	"fmt"
	"github.com/kernelplex/ubase/lib/validation"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUserCreateCommand_Validate(t *testing.T) {
	tests := []struct {
		name     string
		command  UserCreateCommand
		expected *validation.ValidationIssues
	}{
		{
			name: "valid command",
			command: UserCreateCommand{
				Email:       "test@example.com",
				Password:    "ValidPass1!",
				FirstName:   "John",
				LastName:    "Doe",
				DisplayName: "JohnD",
			},
			expected: nil,
		},
		{
			name: "invalid email",
			command: UserCreateCommand{
				Email:       "invalid",
				Password:    "ValidPass1!",
				FirstName:   "John",
				LastName:    "Doe",
				DisplayName: "JohnD",
			},
			expected: &validation.ValidationIssues{
				Issues: []validation.ValidationIssue{
					{Field: "Email", Error: []string{validation.ErrEmailInvalid}},
				},
			},
		},
		{
			name: "missing password",
			command: UserCreateCommand{
				Email:       "test@example.com",
				Password:    "",
				FirstName:   "John",
				LastName:    "Doe",
				DisplayName: "JohnD",
			},
			expected: &validation.ValidationIssues{
				Issues: []validation.ValidationIssue{
					{Field: "Password", Error: []string{validation.ErrPasswordRequired}},
				},
			},
		},
		{
			name: "weak password",
			command: UserCreateCommand{
				Email:       "test@example.com",
				Password:    "weak",
				FirstName:   "John",
				LastName:    "Doe",
				DisplayName: "JohnD",
			},
			expected: &validation.ValidationIssues{
				Issues: []validation.ValidationIssue{
					{
						Field: "Password",
						Error: []string{
							validation.ErrPasswordMinLength,
							validation.ErrPasswordUppercase,
							validation.ErrPasswordNumber,
							validation.ErrPasswordSpecialChar,
						},
					},
				},
			},
		},
		{
			name: "missing required fields",
			command: UserCreateCommand{
				Email:       "",
				Password:    "",
				FirstName:   "",
				LastName:    "",
				DisplayName: "",
			},
			expected: &validation.ValidationIssues{
				Issues: []validation.ValidationIssue{
					{Field: "Email", Error: []string{validation.ErrEmailRequired}},
					{Field: "Password", Error: []string{validation.ErrPasswordRequired}},
					{Field: "FirstName", Error: []string{fmt.Sprintf(validation.ErrFieldRequiredTemplate, "first name")}},
					{Field: "LastName", Error: []string{fmt.Sprintf(validation.ErrFieldRequiredTemplate, "last name")}},
					{Field: "DisplayName", Error: []string{fmt.Sprintf(validation.ErrFieldRequiredTemplate, "display name")}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.command.Validate()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUserUpdateCommand_Validate(t *testing.T) {
	validPass := "ValidPass1!"
	empty := ""

	tests := []struct {
		name     string
		command  UserUpdateCommand
		expected *validation.ValidationIssues
	}{
		{
			name:     "no updates - valid",
			command:  UserUpdateCommand{},
			expected: nil,
		},
		{
			name: "valid password update",
			command: UserUpdateCommand{
				Password: &validPass,
			},
			expected: nil,
		},
		{
			name: "invalid password update",
			command: UserUpdateCommand{
				Password: &empty,
			},
			expected: &validation.ValidationIssues{
				Issues: []validation.ValidationIssue{
					{Field: "Password", Error: []string{validation.ErrPasswordEmpty}},
				},
			},
		},
		{
			name: "invalid optional fields",
			command: UserUpdateCommand{
				FirstName:   &empty,
				LastName:    &empty,
				DisplayName: &empty,
			},
			expected: &validation.ValidationIssues{
				Issues: []validation.ValidationIssue{
					{Field: "FirstName", Error: []string{fmt.Sprintf(validation.ErrFieldEmptyTemplate, "first name")}},
					{Field: "LastName", Error: []string{fmt.Sprintf(validation.ErrFieldEmptyTemplate, "last name")}},
					{Field: "DisplayName", Error: []string{fmt.Sprintf(validation.ErrFieldEmptyTemplate, "display name")}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.command.Validate()
			assert.Equal(t, tt.expected, result)
		})
	}
}
