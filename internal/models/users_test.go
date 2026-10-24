package models

import (
	"testing"

	"snippetbox.dheeraj.net/internal/assert"
)

// TestUserModelExists is an integration test for the database-backed existence
// query, including valid, zero, and unknown IDs.
func TestUserModelExists(t *testing.T) {
	tests := []struct {
        name   string
        userID int
        want   bool
    }{
        {
            name:   "Valid ID",
            userID: 1,
            want:   true,
        },
        {
            name:   "Zero ID",
            userID: 0,
            want:   false,
        },
        {
            name:   "Non-existent ID",
            userID: 2,
            want:   false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
			// Database integration tests are optional in the short test suite.
			if testing.Short() {
        		t.Skip("models: skipping integration test")
    		}
			
            db := newTestDB(t)

            m := UserModel{db}

            exists, err := m.Exists(tt.userID)
            assert.Equal(t, exists, tt.want)
            assert.Nil(t, err)
        })
    }
}