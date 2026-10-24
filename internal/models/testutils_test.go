package models

import (
	"database/sql"
	"os"
	"testing"
)

// newTestDB prepares the integration-test schema and registers teardown so
// database fixtures are removed even when a test fails.
func newTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("mysql", "test_web:pass@/test_snippetbox?parseTime=true&multiStatements=true")
    if err != nil {
        t.Fatal(err)
    }

	script, err := os.ReadFile("./testdata/setup.sql")
    if err != nil {
        db.Close()
        t.Fatal(err)
    }

	_, err = db.Exec(string(script))
    if err != nil {
        db.Close()
        t.Fatal(err)
    }

	t.Cleanup(func() {
		// Keep the connection open until the teardown script has dropped tables.
        defer db.Close()

        script, err := os.ReadFile("./testdata/teardown.sql")
        if err != nil {
            t.Fatal(err)
        }
        _, err = db.Exec(string(script))
        if err != nil {
            t.Fatal(err)
        }
    })

	return db
}