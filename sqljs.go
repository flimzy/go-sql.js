// +build js

// Package sqljs provides a database/sql-compatible interface to SQL.js (https://github.com/lovasoa/sql.js) for GopherJS.
//
// SQL.js does not provide anything like a complete SQLite3 API, and this module even less so. This module exists
// for one primary purpose: To be able to read SQLite3 databases from within a browser. For such purposes, only
// a small subset of features is considered useful.  To this end, this module is tested only for reading
// databases. Although writes are supported, there is currently no way to export the database using this
// package. Also, transactions are not supported.
package sqljs

import (
	"errors"
	"io"

	"database/sql"
	"database/sql/driver"

	"github.com/flimzy/go-sql.js/bindings"
)

// Driver struct. To load an existing database, you must register a new instance
// of this driver, with an io.Reader pointing to the SQLite3 database file.  See
// Open() for an example.
type SQLJSDriver struct {
	Reader io.Reader
}

func init() {
	sql.Register("sqljs", &SQLJSDriver{})
}

// Open will a new database instance. By default, it will create a new database
// in memory. To open an existing database, you must first register a new
// instance as the driver. The DSN string is always ignored.
//
// Example:
//
//    driver := &sqljs.SQLJSDriver{}
//    sql.Register("sqljs-reader", driver)
//    file, _ := os.Open("/path/to/database.db")
//    driver.Reader, _ = file
//    db := sql.Open("sqlite-reader","")
func (d *SQLJSDriver) Open(dsn string) (driver.Conn, error) {
	var db *bindings.Database
	if d.Reader == nil {
		db = bindings.New()
	} else {
		db = bindings.OpenReader(d.Reader)
	}
	return &SQLJSConn{db}, nil
}

// Connection struct
type SQLJSConn struct {
	*bindings.Database
}

// Prepare the query string. Return a new statement handle.
func (c *SQLJSConn) Prepare(query string) (driver.Stmt, error) {
	s, err := c.Database.Prepare(query)
	return &SQLJSStmt{s, c.Database}, err
}

// Begin a transaction -- not supported (will always return an error)
func (c *SQLJSConn) Begin() (driver.Tx, error) {
	return nil, errors.New("Transactions not supported")
}

// Close the database and free memory.
func (c *SQLJSConn) Close() error {
	return c.Database.Close()
}

// Statement struct.
type SQLJSStmt struct {
	*bindings.Statement
	db *bindings.Database // So we can call GetRowsModified()
}

// Close the statement handler.
func (s *SQLJSStmt) Close() (e error) {
	return s.Close()
}

// NumInput is unsupported. It will always return -1.
func (s *SQLJSStmt) NumInput() int {
	return -1
}

// Exec executes a query that does not return any rows.
func (s *SQLJSStmt) Exec(args []driver.Value) (r driver.Result, e error) {
	err := s.RunParams(valuesToInterface(args))
	return &SQLJSResult{
		s.db.GetRowsModified(),
	}, err
}

// Result struct.
type SQLJSResult struct {
	rowsAffected int64
}

// LastInsertId is not supported. It will always return an error.
func (SQLJSResult) LastInsertId() (int64, error) {
	return 0, errors.New("LastInsertId not available")
}

// RowsAffected is not supported. It will always return an error.
func (s *SQLJSResult) RowsAffected() (int64, error) {
	return s.rowsAffected, nil
}

func valuesToInterface(args []driver.Value) []interface{} {
	params := make([]interface{}, len(args))
	for i, arg := range args {
		params[i] = interface{}(arg)
	}
	return params
}

// Query executes a query that may return rows, such as a SELECT.
func (s *SQLJSStmt) Query(args []driver.Value) (r driver.Rows, e error) {
	if err := s.Bind(valuesToInterface(args)); err != nil {
		return nil, err
	}
	return &SQLJSRows{s.Statement, false, nil}, nil
}

// Rows struct.
type SQLJSRows struct {
	*bindings.Statement
	firstStep bool
	lastStep  *lastStep
}

type lastStep struct {
	ok  bool
	err error
}

func (r *SQLJSRows) step() (bool, error) {
	if r.lastStep == nil {
		ok, err := r.Step()
		r.firstStep = true
		r.lastStep = &lastStep{ok, err}
	}
	return r.lastStep.ok, r.lastStep.err
}

// Close closes the Rows iterator.
func (r *SQLJSRows) Close() (e error) {
	if !r.firstStep {
		r.Reset()
	}
	return nil
}

// Columns returns the names of the columns.
func (r *SQLJSRows) Columns() []string {
	r.step()
	cols, _ := r.GetColumnNames()
	return cols
}

// Next is called to populate the next row of data into the provided slice.
func (r *SQLJSRows) Next(dest []driver.Value) (e error) {
	ok, err := r.step()
	r.lastStep = nil
	if err != nil {
		return err
	}
	if !ok {
		return io.EOF
	}
	result, err := r.Get()
	if err != nil {
		return err
	}
	for i, _ := range dest {
		dest[i] = result[i]
	}
	return nil
}
