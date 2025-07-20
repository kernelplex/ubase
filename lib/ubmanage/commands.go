package ubmanage

import (
	"github.com/kernelplex/ubase/lib/ubstatus"
	"github.com/kernelplex/ubase/lib/ubvalidation"
)

// Response is a generic response type that can be used to return data from a command or query.
type Response[T any] struct {
	Status           ubstatus.StatusCode            `json:"status"`
	Message          string                         `json:"message,omitempty"`
	ValidationIssues []ubvalidation.ValidationIssue `json:"validationIssues,omitempty"`
	Data             T                              `json:"data,omitempty"`
}
