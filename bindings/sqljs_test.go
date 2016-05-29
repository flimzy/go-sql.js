package sqljs

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestDb(t *testing.T) {
	db := New()

	if err := db.Close(); err != nil {
		t.Fatalf("Error closing DB: %s", err)
	}
}

func TestReader(t *testing.T) {
	file, originalArray := OpenTestDb(t)
	db := OpenReader(file)

	exp := db.Export()
	buf := new(bytes.Buffer)
	buf.ReadFrom(exp)
	if !bytes.Equal(buf.Bytes(), originalArray) {
		t.Fatalf("Exported and imported databases are not the same!")
	}

	stmt, err := db.Prepare("SELECT * FROM test ORDER BY id", []interface{}{})
	if err != nil {
		t.Fatalf("Error preparing statement: %s", err)
	}
	if succ, err := stmt.Step(); err != nil {
		t.Fatalf("Error stepping through statement: %s", err)
	} else if succ != true {
		t.Fatal("Step failed")
	}
	results, err := stmt.Get()
	if err != nil {
		t.Fatalf("Error calling Get(): %s", err)
	}
	if v := int(results[0].(float64)); v != 1 {
		t.Fatalf("Unexpected value fetched: %d", v)
	}
	if v := results[1].(string); v != "Bob" {
		t.Fatalf("Unexpected value fetched: %s", v)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("Error closing DB: %s", err)
	}
}

func OpenTestDb(t *testing.T) (io.Reader, []byte) {
	file, err := os.Open("test.db")
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
