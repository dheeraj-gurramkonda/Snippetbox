package models

import (
	"database/sql"
	"errors"
	"time"
)

// SnippetModelInterface lets handlers use either the database-backed model or
// a test double with the same snippet operations.
type SnippetModelInterface interface {
    Insert(title string, content string, expires int) (int, error)
    Get(id int) (Snippet, error)
    Latest() ([]Snippet, error)
}

// Snippet is a stored piece of text and the time range during which it is visible.
type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

// SnippetModel performs snippet queries using the shared database connection.
type SnippetModel struct {
	DB *sql.DB
}

// Insert saves a new snippet record and returns the generated database ID.
func (m *SnippetModel) Insert(title string, content string, expires int) (int, error) {
	// Calculate creation and expiry timestamps in MySQL using UTC.
	stmt := `INSERT INTO snippets (title, content, created, expires) 
	VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`

	result, err := m.DB.Exec(stmt, title, content, expires)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, nil
	}

	return int(id), nil
}

// Get retrieves a single non-expired snippet by its ID and returns any not-found error.
func (m *SnippetModel) Get(id int) (Snippet, error) {
	// Expired snippets are treated like missing records, so callers never expose
	// them through the view endpoint.
	stmt := `SELECT id, title, content, created, expires FROM snippets
    WHERE expires > UTC_TIMESTAMP() AND id = ?`

	row := m.DB.QueryRow(stmt, id)

	var s Snippet

	err := row.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return Snippet{}, ErrNoRecord
		default:
			return Snippet{}, err
		}
	}
	return s, nil
}

// Latest returns up to ten non-expired snippets sorted by newest first.
func (m *SnippetModel) Latest() ([]Snippet, error) {
	// Keep the home page bounded and show the newest active records first.
	stmt := `SELECT id, title, content, created, expires FROM snippets
    WHERE expires > UTC_TIMESTAMP() ORDER BY id DESC LIMIT 10`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	// A nil slice is returned when there are no matching rows; templates handle
	// that as the empty-list case.
	var snippets []Snippet

	for rows.Next() {
		var s Snippet
		err := rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
		if err != nil {
			return nil, err
		}

		snippets = append(snippets, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return snippets, nil
}
