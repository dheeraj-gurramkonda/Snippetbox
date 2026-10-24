package mocks

import (
    "snippetbox.dheeraj.net/internal/models"
)

// UserModel is an in-memory test double with deterministic success and failure
// cases for signup, login, and session authentication tests.
type UserModel struct{}

func (m *UserModel) Insert(name, email, password string) error {
    // This address exercises the duplicate-email response path in handlers.
    switch email {
    case "dupe@example.com":
        return models.ErrDuplicateEmail
    default:
        return nil
    }
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
    // Only this credential pair represents a successful login in tests.
    if email == "alice@example.com" && password == "pa$$word" {
        return 1, nil
    }

    return 0, models.ErrInvalidCredentials
}

// Exists recognizes the single fixture user used by authenticated tests.
func (m *UserModel) Exists(id int) (bool, error) {
    switch id {
    case 1:
        return true, nil
    default:
        return false, nil
    }
}