package models

import (
	"errors"
)

// Sentinel errors describe expected model outcomes so callers can branch with
// errors.Is without depending on database-driver error details.
var (
	// ErrNoRecord means a requested record was not found or is no longer visible.
	ErrNoRecord = errors.New("models: no matching record found")

	// ErrInvalidCredentials deliberately covers both unknown users and bad passwords.
	ErrInvalidCredentials = errors.New("models: invalid credentials")

	// ErrDuplicateEmail signals that an account already uses the submitted address.
	ErrDuplicateEmail = errors.New("models: duplicate email")
)
