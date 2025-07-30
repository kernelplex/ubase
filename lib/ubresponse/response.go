package ubresponse

import (
	"encoding/json"
	"fmt"

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

func (r Response[T]) ToJSON() ([]byte, error) {
	jsonResponse, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}
	return jsonResponse, nil
}

func SuccessAny() Response[any] {
	return Response[any]{
		Status: ubstatus.Success,
	}
}

func Success[T any](data T) Response[T] {
	return Response[T]{
		Status: ubstatus.Success,
		Data:   data,
	}
}

func Error[T any](message string) Response[T] {
	return Response[T]{
		Status:  ubstatus.UnexpectedError,
		Message: message,
	}
}

func ValidationError[T any](issues []ubvalidation.ValidationIssue) Response[T] {
	return Response[T]{
		Status:           ubstatus.ValidationError,
		ValidationIssues: issues,
	}
}

func StatusError[T any](status ubstatus.StatusCode, message string) Response[T] {
	return Response[T]{
		Status:  status,
		Message: message,
	}
}
