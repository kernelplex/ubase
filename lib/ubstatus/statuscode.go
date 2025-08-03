package ubstatus

type StatusCode string

const (
	Success         StatusCode = "success"
	PartialSuccess  StatusCode = "partial_success"
	NotFound        StatusCode = "not_found"
	NotAuthorized   StatusCode = "not_authorized"
	AlreadyExists   StatusCode = "already_exists"
	ValidationError StatusCode = "validation_error"
	UnexpectedError StatusCode = "unexpected_error"
)
