package ubase

import (
	"fmt"
	"github.com/kernelplex/ubase/lib/ubvalidation"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUserCreateCommand_Validate(t *testing.T) {
	tests := []struct {
		name     string
		command  UserCreateCommand
		expected *ubvalidation.ValidationIssues
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
			expected: &ubvalidation.ValidationIssues{
				Issues: []ubvalidation.ValidationIssue{
					{Field: "Email", Error: []string{ubvalidation.ErrEmailInvalid}},
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
			expected: &ubvalidation.ValidationIssues{
				Issues: []ubvalidation.ValidationIssue{
					{Field: "Password", Error: []string{ubvalidation.ErrPasswordRequired}},
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
			expected: &ubvalidation.ValidationIssues{
				Issues: []ubvalidation.ValidationIssue{
					{
						Field: "Password",
						Error: []string{
							ubvalidation.ErrPasswordMinLength,
							ubvalidation.ErrPasswordUppercase,
							ubvalidation.ErrPasswordNumber,
							ubvalidation.ErrPasswordSpecialChar,
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
			expected: &ubvalidation.ValidationIssues{
				Issues: []ubvalidation.ValidationIssue{
					{Field: "Email", Error: []string{ubvalidation.ErrEmailRequired}},
					{Field: "Password", Error: []string{ubvalidation.ErrPasswordRequired}},
					{Field: "FirstName", Error: []string{fmt.Sprintf(ubvalidation.ErrFieldRequiredTemplate, "first name")}},
					{Field: "LastName", Error: []string{fmt.Sprintf(ubvalidation.ErrFieldRequiredTemplate, "last name")}},
					{Field: "DisplayName", Error: []string{fmt.Sprintf(ubvalidation.ErrFieldRequiredTemplate, "display name")}},
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
		expected *ubvalidation.ValidationIssues
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
			expected: &ubvalidation.ValidationIssues{
				Issues: []ubvalidation.ValidationIssue{
					{Field: "Password", Error: []string{ubvalidation.ErrPasswordEmpty}},
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
			expected: &ubvalidation.ValidationIssues{
				Issues: []ubvalidation.ValidationIssue{
					{Field: "FirstName", Error: []string{fmt.Sprintf(ubvalidation.ErrFieldEmptyTemplate, "first name")}},
					{Field: "LastName", Error: []string{fmt.Sprintf(ubvalidation.ErrFieldEmptyTemplate, "last name")}},
					{Field: "DisplayName", Error: []string{fmt.Sprintf(ubvalidation.ErrFieldEmptyTemplate, "display name")}},
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
