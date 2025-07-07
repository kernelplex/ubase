package ubase

import (
	"github.com/kernelplex/evercore/base"
	_ "github.com/kernelplex/ubase/lib/evercoregen"
	"github.com/kernelplex/ubase/lib/ubstate"
	"github.com/kernelplex/ubase/lib/ubvalidation"
)

// evercore:aggregate
type UserAggregate struct {
	evercore.StateAggregate[ubstate.UserState]
}

type Response struct {
	Status           string                         `json:"status"`
	Message          string                         `json:"message"`
	ValidationIssues *ubvalidation.ValidationIssues `json:"validationIssues,omitempty"`
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

func (r UserCreateCommand) Validate() *ubvalidation.ValidationIssues {
	var issues []ubvalidation.ValidationIssue

	issues = append(issues, ubvalidation.ValidateEmail(r.Email)...)

	if r.Password != "" {
		if passwordErrors := ubvalidation.ValidatePasswordComplexity(r.Password); len(passwordErrors) > 0 {
			issues = append(issues, ubvalidation.ValidationIssue{
				Field: "Password",
				Error: passwordErrors,
			})
		}
	} else {
		issues = append(issues, ubvalidation.ValidationIssue{
			Field: "Password",
			Error: []string{"password is required"},
		})
	}
	issues = append(issues, ubvalidation.ValidateField("FirstName", r.FirstName, true, 0)...)
	issues = append(issues, ubvalidation.ValidateField("LastName", r.LastName, true, 0)...)
	issues = append(issues, ubvalidation.ValidateField("DisplayName", r.DisplayName, true, 0)...)

	if len(issues) == 0 {
		return nil
	}
	return &ubvalidation.ValidationIssues{Issues: issues}
}

type UserUpdateCommand struct {
	Id          int64   `json:"id"`
	Password    *string `json:"password,omitempty"`
	FirstName   *string `json:"firstName,omitempty"`
	LastName    *string `json:"lastName,omitempty"`
	DisplayName *string `json:"displayName,omitempty"`
}

func (r UserUpdateCommand) Validate() *ubvalidation.ValidationIssues {
	var issues []ubvalidation.ValidationIssue

	if r.Password != nil {
		if *r.Password == "" {
			issues = append(issues, ubvalidation.ValidationIssue{
				Field: "Password",
				Error: []string{"password cannot be empty if provided"},
			})
		} else if passwordErrors := ubvalidation.ValidatePasswordComplexity(*r.Password); len(passwordErrors) > 0 {
			issues = append(issues, ubvalidation.ValidationIssue{
				Field: "Password",
				Error: passwordErrors,
			})
		}
	}

	issues = append(issues, ubvalidation.ValidateOptionalField("FirstName", r.FirstName, 0)...)
	issues = append(issues, ubvalidation.ValidateOptionalField("LastName", r.LastName, 0)...)
	issues = append(issues, ubvalidation.ValidateOptionalField("DisplayName", r.DisplayName, 0)...)

	if len(issues) == 0 {
		return nil
	}
	return &ubvalidation.ValidationIssues{Issues: issues}
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
