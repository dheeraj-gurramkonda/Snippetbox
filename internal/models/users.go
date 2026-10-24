package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

// UserModelInterface describes the account operations needed by HTTP handlers.
type UserModelInterface interface {
    Insert(name, email, password string) error
    Authenticate(email, password string) (int, error)
    Exists(id int) (bool, error)
}

// Users represents the user fields returned from storage. Password data is
// stored as a bcrypt hash rather than as the original password.
type Users struct {
	ID               int
	Name             string
	Email            string
	Hashed_passwords []byte
	Created          time.Time
}

// UserModel stores account records and authenticates users against MySQL.
type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) Insert(name, email, password string) error {
	// Hash passwords before they reach persistent storage; bcrypt includes a
	// per-password salt in the resulting hash.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := `INSERT INTO users (name, email, hashed_password, created)
    VALUES(?, ?, ?, UTC_TIMESTAMP())`

	_, err = m.DB.Exec(stmt, name, email, string(hashedPassword))
	if err != nil {
		var mySQLErr *mysql.MySQLError
		switch {
		case errors.As(err, &mySQLErr) && mySQLErr.Number == 1062 && strings.Contains(mySQLErr.Message, "users_uc_email"):
			// Convert this specific unique-email violation into a stable model
			// error that handlers can map to a form message.
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil

}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	var id int
	var hashedPassword []byte

	stmt := "SELECT id, hashed_password FROM users WHERE email = ?"
	err := m.DB.QueryRow(stmt, email).Scan(&id, &hashedPassword)
	if err != nil {
		switch {
			case errors.Is(err, sql.ErrNoRows):
				// Keep unknown addresses indistinguishable from wrong passwords.
				return 0, ErrInvalidCredentials
			default:
				return 0, err
		}
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		switch {
			case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
				// Both credential failure paths intentionally use one public error.
				return 0, ErrInvalidCredentials
			default:
				return 0, err
		}
	}
	
	return id, nil
}

// Exists reports whether an account ID is still present in the database.
func (m *UserModel) Exists(id int) (bool, error) {
	var exists bool

	stmt := "SELECT EXISTS(SELECT true FROM users WHERE id = ?)"
	err := m.DB.QueryRow(stmt, id).Scan(&exists)
	if err != nil {
		return false, err
	}
	
	return exists, nil
}
