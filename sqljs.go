// +build js

// Package sqljs provides a database/sql-compatible interface to SQL.js (https://github.com/lovasoa/sql.js) for GopherJS.
package sqljs

import (
	"errors"
	"io"

	"database/sql"
	"database/sql/driver"

	"github.com/flimzy/go-sql.js/bindings"
)

type SQLJSDriver struct {
	Reader io.Reader
}

func init() {
	sql.Register("sqljs", &SQLJSDriver{})
}

// Open will open a new, empty database if dsn is the empty string.
// If the DSN string is non-empty, it will attempt to open the file
// previously provided to the SetDB() function.
func (d *SQLJSDriver) Open(dsn string) (driver.Conn, error) {
	var db *bindings.Database
	if d.Reader == nil {
		db = bindings.New()
	} else {
		db = bindings.OpenReader(d.Reader)
	}
	return &SQLJSConn{db}, nil
}

type SQLJSConn struct {
	*bindings.Database
}

func (c *SQLJSConn) Prepare(query string) (driver.Stmt, error) {
	s, err := c.Database.Prepare(query)
	return &SQLJSStmt{s}, err
}

func (c *SQLJSConn) Begin() (driver.Tx, error) {
	return nil, nil
}

func (c *SQLJSConn) Close() error {
	return c.Database.Close()
}

type SQLJSStmt struct {
	*bindings.Statement
}

func (s *SQLJSStmt) Close() (e error) {
	return s.Close()
}

func (s *SQLJSStmt) NumInput() int {
	return -1
}

func (s *SQLJSStmt) Exec(args []driver.Value) (r driver.Result, e error) {
	return &SQLJSResult{}, s.RunParams(valuesToInterface(args))
}

type SQLJSResult struct{}

func (SQLJSResult) LastInsertId() (int64, error) {
	return 0, errors.New("LastInsertId not available")
}

func (SQLJSResult) RowsAffected() (int64, error) {
	return 0, errors.New("RowsAffected not available")
}

func valuesToInterface(args []driver.Value) []interface{} {
	params := make([]interface{}, len(args))
	for i, arg := range args {
		params[i] = interface{}(arg)
	}
	return params
}

func (s *SQLJSStmt) Query(args []driver.Value) (r driver.Rows, e error) {
	if err := s.Bind(valuesToInterface(args)); err != nil {
		return nil, err
	}
	return &SQLJSRows{s.Statement, false, nil}, nil
}

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

func (r *SQLJSRows) Close() (e error) {
	if !r.firstStep {
		r.Reset()
	}
	return nil
}

func (r *SQLJSRows) Columns() []string {
	r.step()
	cols, _ := r.GetColumnNames()
	return cols
}

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
