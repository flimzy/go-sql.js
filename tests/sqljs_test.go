// +build js

package test

import (
	"testing"

	"database/sql"
	_ "github.com/flimzy/go-sql.js"
)

func TestOpenEmpty(t *testing.T) {
	db, err := sql.Open("sqljs", "")
	if err != nil {
		t.Fatalf("Error opening empty database: %s", err)
	}
	
	if _,err := db.Prepare("an invalid statement"); err.Error() != "JavaScript error: near \"an\": syntax error" {
		t.Fatalf("Error preparing statement: %s", err)
	}

	stmt,err := db.Prepare("SELECT 1 AS foo")
	if err != nil {
		t.Fatalf("Error preparing statement: %s", err)
	}
	rows, err := stmt.Query()
	if err != nil {
		t.Fatalf("Error executing query: %s", err)
	}
	cols, err := rows.Columns()
	if err != nil {
		t.Fatalf("Error fetching column names: %s", err)
	}
	if len(cols) != 1 {
		t.Fatalf("%d columns returned, expected 1", len(cols))
	}
	if cols[0] != "foo" {
		t.Fatalf("Unknown column: %s", cols[0])
	}

	if err := stmt.Close(); err != nil {
		t.Fatalf("Error closing statement handle: %s", err)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("Error closing empty database: %s", err)
	}
}
