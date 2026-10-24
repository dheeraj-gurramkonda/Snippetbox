package mocks

import (
    "time"

    "snippetbox.dheeraj.net/internal/models"
)

// mockSnippet is the predictable record returned by the snippet test double.
var mockSnippet = models.Snippet{
    ID:      1,
    Title:   "An old silent pond",
    Content: "An old silent pond...",
    Created: time.Now(),
    Expires: time.Now(),
}

// SnippetModel implements the snippet model interface without a database.
type SnippetModel struct{}

// Insert returns a stable ID because tests do not persist records.
func (m *SnippetModel) Insert(title string, content string, expires int) (int, error) {
    return 2, nil
}

// Get returns the fixture for ID 1 and a model not-found error for other IDs.
func (m *SnippetModel) Get(id int) (models.Snippet, error) {
    switch id {
    case 1:
        return mockSnippet, nil
    default:
        return models.Snippet{}, models.ErrNoRecord
    }
}

// Latest returns one fixture so handlers can render their non-empty state.
func (m *SnippetModel) Latest() ([]models.Snippet, error) {
    return []models.Snippet{mockSnippet}, nil
}
