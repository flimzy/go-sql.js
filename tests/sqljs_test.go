// +build js

package test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"database/sql"
	"github.com/flimzy/go-sql.js"
)

func TestOpenEmpty(t *testing.T) {
	db, err := sql.Open("sqljs", "")
	if err != nil {
		t.Fatalf("Error opening empty database: %s", err)
	}

	if _, err := db.Prepare("an invalid statement"); err.Error() != "JavaScript error: near \"an\": syntax error" {
		t.Fatalf("Error preparing statement: %s", err)
	}

	stmt, err := db.Prepare("SELECT 1 AS foo")
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

func TestOpenExisting(t *testing.T) {
	reader, _ := OpenTestDb(t)
	sqljs.AddReader("test.db", reader)
	db, err := sql.Open("sqljs", "test.db")
	if err != nil {
		t.Fatalf("Error opening existing database: %s", err)
	}

	stmt, err := db.Prepare("INSERT INTO test (id,name) VALUES (?,?)")
	if err != nil {
		t.Fatalf("Error preparing statement: %s", err)
	}

	result, err := stmt.Exec(3, "John")
	if err != nil {
		t.Fatalf("Error execing statement: %s", err)
	}
	if _, err := result.LastInsertId(); err.Error() != "LastInsertId not available" {
		t.Fatalf("Unexpected error calling LastInsertId: %s", err)
	}
	ra, err := result.RowsAffected()
	if ra != 1 {
		t.Fatalf("Expected 1 modified row, got %d", ra)
	}
	if err != nil {
		t.Fatalf("Unexpected error calling RowsAffected: %s", err)
	}

	stmt, err = db.Prepare("SELECT * FROM test WHERE id=?")
	if err != nil {
		t.Fatalf("Error preparing SELECT: %s", err)
	}
	rows, err := stmt.Query(2)
	if err != nil {
		t.Fatalf("Error executing query: %s", err)
	}
	cols, err := rows.Columns()
	if err != nil {
		t.Fatalf("Error fetching columns: %s", err)
	}
	if cols[0] != "id" || cols[1] != "name" {
		t.Fatalf("Unexpected columns: %s, %s", cols[0], cols[1])
	}
	for rows.Next() {
		var id int
		var name string
		err := rows.Scan(&id, &name)
		if err != nil {
			t.Fatalf("Error scanning row: %s", err)
		}
		if id != 2 || name != "Alice" {
			t.Fatalf("Unexpected results: id = %d, name = %s", id, name)
		}
	}
}

func OpenTestDb(t *testing.T) (io.Reader, []byte) {
	file, err := os.Open("../bindings/test.db")
	if err != nil {
		t.Fatalf("Error opening db file: %s", err)
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(file)
	byteArray := buf.Bytes()
	if err := file.Close(); err != nil {
		t.Fatalf("Error closing db file: %s", err)
	}
	return bytes.NewReader(byteArray), byteArray
}
