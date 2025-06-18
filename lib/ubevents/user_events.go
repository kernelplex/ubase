package ubevents

// evercore:state-event
type UserCreatedEvent struct {
	Email        *string `json:"email"`
	PasswordHash *string `json:"passwordHash"`
	FirstName    *string `json:"firstName"`
	LastName     *string `json:"lastName"`
	DisplayName  *string `json:"displayName"`
}

// evercore:state-event
type UserUpdatedEvent struct {
	PasswordHash *string `json:"passwordHash"`
	FirstName    *string `json:"firstName"`
	LastName     *string `json:"lastName"`
	DisplayName  *string `json:"displayName"`
}

// evercore:state-event
type UserLoginFailedEvent struct {
	LastLoginAttempt    int64 `json:"lastLoginAttempt"`
	FailedLoginAttempts int64 `json:"failedLoginAttempts"`
}

// evercore:state-event
type UserLoginSucceededEvent struct {
	LastLoginAttempt    int64 `json:"lastLoginAttempt"`
	FailedLoginAttempts int64 `json:"failedLoginAttempts"`
}

// evercore:state-event
type UserRolesUpdatedEvent struct {
	Roles []int64 `json:"roles"`
}
