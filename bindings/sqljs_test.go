package bindings

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"
)

func TestDb(t *testing.T) {
	db := New()

	if err := db.Run("CREATE TABLE foo (x int)"); err != nil {
		t.Fatalf("Error creating table: %s", err)
	}
	if err := db.Run("INSERT INTO foo (x) VALUES (1),(2),(3)"); err != nil {
		t.Fatalf("Error inserting: %s", err)
	}
	if modified := db.GetRowsModified(); modified != 3 {
		t.Fatalf("Unexpected number of rows modified: %i intead of 3", modified)
	}

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

	stmt, err := db.Prepare("SELECT * FROM test ORDER BY id")
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

	colNames, err := stmt.GetColumnNames()
	if err != nil {
		t.Fatalf("Error fetching column names: %s", err)
	}
	if n := len(colNames); n != 2 {
		t.Fatalf("Unexpected number of clumns found: %d", n)
	}
	if colNames[0] != "id" || colNames[1] != "name" {
		t.Fatalf("Unexpected column names found")
	}

	stmt, err = db.Prepare("SELECT name FROM test WHERE id=?")
	if err != nil {
		t.Fatalf("Error preparing statement with placeholder: %s", err)
	}
	results, err = stmt.GetParams([]interface{}{2})
	if err != nil {
		t.Fatalf("Error calling Get([2]): %s", err)
	}
	if v := results[0].(string); v != "Alice" {
		t.Fatalf("Unexpected value fetched: %s", v)
	}

	stmt.Reset()
	if err := stmt.Bind([]interface{}{1}); err != nil {
		t.Fatalf("Error binding: %s", err)
	}

	stmt.Step()
	results, err = stmt.Get()
	if err != nil {
		t.Fatalf("Error calling Get(): %s", err)
	}
	if v := results[0].(string); v != "Bob" {
		t.Fatalf("Unexpected value fetched: %s", v)
	}

	// Bind named parameters
	stmt, err = db.Prepare("SELECT name FROM test WHERE id=$id")
	if err != nil {
		t.Fatalf("Error preparing statement with placeholder: %s", err)
	}
	if err := stmt.BindNamed(map[string]interface{}{"$id": 1}); err != nil {
		t.Fatalf("Error binding named parameters: %s", err)
	}
	stmt.Step()
	result, err := stmt.GetAsMap()
	if err != nil {
		t.Fatalf("Error calling Get(): %s", err)
	}
	if v := result["name"].(string); v != "Bob" {
		t.Fatalf("Error fetching with GetAsMap: %s", v)
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

func TestReader2(t *testing.T) {
	file, _ := OpenTestDb(t)
	db := OpenReader(file)

	result, err := db.Exec("SELECT * FROM test ORDER BY id; SELECT name FROM test ORDER BY name")
	if err != nil {
		t.Fatalf("Error with Exec(): %s", err)
	}

	expected := []Result{
		Result{
			Columns: []string{"id", "name"},
			Values: [][]interface{}{
				[]interface{}{
					float64(1),
					string("Bob"),
				},
				[]interface{}{
					float64(2),
					string("Alice"),
				},
			},
		},
		Result{
			Columns: []string{"name"},
			Values: [][]interface{}{
				[]interface{}{
					string("Alice"),
				},
				[]interface{}{
					string("Bob"),
				},
			},
		},
	}

	if !reflect.DeepEqual(expected, result) {
		fmt.Printf("Values don't match\n")
		fmt.Printf("Expected: %v\n", expected)
		fmt.Printf("  Result: %v\n", result)
		t.Fatal()
	}
	if err := db.Close(); err != nil {
		t.Fatalf("Error closing DB: %s", err)
	}

}
