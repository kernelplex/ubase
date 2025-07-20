package ubstate

type UserState struct {
	Email                 string  `json:"email"`
	PasswordHash          string  `json:"passwordHash"`
	FirstName             string  `json:"firstName"`
	LastName              string  `json:"lastName"`
	VerificationToken     *string `json:"verificationToken,omitempty"`
	Verified              bool    `json:"verified"`
	DisplayName           string  `json:"displayName"`
	ResetToken            *string `json:"resetToken,omitempty"`
	LastLogin             int64   `json:"lastLogin,omitempty"`
	LastLoginAttempt      int64   `json:"lastLoginAttempt,omitempty"`
	FailedLoginAttempts   int64   `json:"failedLoginAttempts,omitempty"`
	TwoFactorSharedSecret *string `json:"twoFactorSharedSecret,omitempty"`
	Roles                 []int64 `json:"roles,omitempty"`
}
