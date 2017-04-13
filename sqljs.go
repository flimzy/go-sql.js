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
	"fmt"
	"io"

	"database/sql"
	"database/sql/driver"

	"github.com/flimzy/go-sql.js/bindings"
)

var readers map[string]io.Reader

// Driver struct. To load an existing database, you must register a new instance
// of this driver, with an io.Reader pointing to the SQLite3 database file.  See
// Open() for an example.
type SQLJSDriver struct{}

func init() {
	sql.Register("sqljs", &SQLJSDriver{})
}

func AddReader(name string, reader io.Reader) error {
	if readers == nil {
		readers = make(map[string]io.Reader)
	}
	if _, ok := readers[name]; ok {
		return fmt.Errorf("Reader `%s` already registered", name)
	}
	readers[name] = reader
	return nil
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
//    db := sql.Open("sqljs-reader","")
func (d *SQLJSDriver) Open(dsn string) (driver.Conn, error) {
	var db *bindings.Database
	if dsn == "" {
		db = bindings.New()
	} else {
		reader, ok := readers[dsn]
		if !ok {
			return nil, fmt.Errorf("reader `%s` does not exist; all AddReader() first", reader)
		}
		delete(readers, dsn)
		db = bindings.OpenReader(reader)
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
func (s *SQLJSStmt) Close() error {
	if s.Free() {
		return nil
	}
	return errors.New("Error freeing statement memory")
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
	return &SQLJSRows{s.Statement, nil, []string{}, nil}, nil
}

// Rows struct.
type SQLJSRows struct {
	*bindings.Statement
	prevStep *prevStep
	cols     []string
	err      error
}

type prevStep struct {
	ok  bool
	err error
}

func (r *SQLJSRows) step() (bool, error) {
	if r.prevStep == nil {
		ok, err := r.Step()
		r.prevStep = &prevStep{ok, err}
	}
	return r.prevStep.ok, r.prevStep.err
}

// Close closes the Rows iterator.
func (r *SQLJSRows) Close() error {
	r.Reset()
	return nil
}

func (r *SQLJSRows) setColumns() {
	if len(r.cols) > 0 {
		return
	}
	if ok, err := r.step(); err != nil {
		r.err = err
		return
	} else if !ok {
		r.err = errors.New("cannot read column names; nothing to fetch")
		return
	}
	cols, err := r.GetColumnNames()
	r.cols = cols
	r.err = err
}

// Columns returns the names of the columns.
func (r *SQLJSRows) Columns() []string {
	r.setColumns()
	return r.cols
}

// Next is called to populate the next row of data into the provided slice.
func (r *SQLJSRows) Next(dest []driver.Value) error {
	if err := r.err; err != nil {
		r.err = nil
		return err
	}
	r.setColumns()
	ok, err := r.step()
	r.prevStep = nil
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
