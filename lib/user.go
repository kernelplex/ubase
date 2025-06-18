package ubase

import (
	"github.com/kernelplex/evercore/base"
	"github.com/kernelplex/ubase/lib/validation"
)

type UserState struct {
	Email                 string  `json:"email"`
	PasswordHash          string  `json:"passwordHash"`
	FirstName             string  `json:"firstName"`
	LastName              string  `json:"lastName"`
	DisplayName           string  `json:"displayName"`
	ResetToken            *string `json:"resetToken,omitempty"`
	LastLogin             int64   `json:"lastLogin,omitempty"`
	LastLoginAttempt      int64   `json:"lastLoginAttempt,omitempty"`
	FailedLoginAttempts   int64   `json:"failedLoginAttempts,omitempty"`
	TwoFactorSharedSecret *string `json:"twoFactorSharedSecret,omitempty"`
	Roles                 []int64 `json:"roles,omitempty"`
}

// evercore:aggregate
type UserAggregate struct {
	evercore.StateAggregate[UserState]
}

type Response struct {
	Status           string                       `json:"status"`
	Message          string                       `json:"message"`
	ValidationIssues *validation.ValidationIssues `json:"validationIssues,omitempty"`
}

type UserCreateCommand struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	DisplayName string `json:"displayName"`
}

type UserCreateResponse struct {
	Response `json:"response"`
	Id       int64 `json:"id"`
}

func (r UserCreateCommand) Validate() *validation.ValidationIssues {
	var issues []validation.ValidationIssue

	issues = append(issues, validation.ValidateEmail(r.Email)...)

	if r.Password != "" {
		if passwordErrors := validation.ValidatePasswordComplexity(r.Password); len(passwordErrors) > 0 {
			issues = append(issues, validation.ValidationIssue{
				Field: "Password",
				Error: passwordErrors,
			})
		}
	} else {
		issues = append(issues, validation.ValidationIssue{
			Field: "Password",
			Error: []string{"password is required"},
		})
	}
	issues = append(issues, validation.ValidateField("FirstName", r.FirstName, true, 0)...)
	issues = append(issues, validation.ValidateField("LastName", r.LastName, true, 0)...)
	issues = append(issues, validation.ValidateField("DisplayName", r.DisplayName, true, 0)...)

	if len(issues) == 0 {
		return nil
	}
	return &validation.ValidationIssues{Issues: issues}
}

type UserUpdateCommand struct {
	Id          int64   `json:"id"`
	Password    *string `json:"password,omitempty"`
	FirstName   *string `json:"firstName,omitempty"`
	LastName    *string `json:"lastName,omitempty"`
	DisplayName *string `json:"displayName,omitempty"`
}

func (r UserUpdateCommand) Validate() *validation.ValidationIssues {
	var issues []validation.ValidationIssue

	if r.Password != nil {
		if *r.Password == "" {
			issues = append(issues, validation.ValidationIssue{
				Field: "Password",
				Error: []string{"password cannot be empty if provided"},
			})
		} else if passwordErrors := validation.ValidatePasswordComplexity(*r.Password); len(passwordErrors) > 0 {
			issues = append(issues, validation.ValidationIssue{
				Field: "Password",
				Error: passwordErrors,
			})
		}
	}

	issues = append(issues, validation.ValidateOptionalField("FirstName", r.FirstName, 0)...)
	issues = append(issues, validation.ValidateOptionalField("LastName", r.LastName, 0)...)
	issues = append(issues, validation.ValidateOptionalField("DisplayName", r.DisplayName, 0)...)

	if len(issues) == 0 {
		return nil
	}
	return &validation.ValidationIssues{Issues: issues}
}

type UserUpdatedResponse struct {
	Response
}

type UserLoginCommand struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserLoginResponse struct {
	Response
	UserId      int64   `json:"userId"`
	LastName    *string `json:"lastName,omitempty"`
	FirstName   *string `json:"firstName,omitempty"`
	DisplayName *string `json:"displayName,omitempty"`
	Email       *string `json:"email,omitempty"`
}

type UserSetRolesComand struct {
	Id      int64   `json:"id"`
	RoleIds []int64 `json:"roleIds"`
}

type UserSetRolesResponse struct {
	Response `json:"response"`
}
