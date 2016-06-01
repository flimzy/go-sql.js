// +build js

package sqljs

import (
	"errors"
	"io"

	"database/sql"
	"database/sql/driver"

	"github.com/gopherjs/gopherjs/js"
	"github.com/flimzy/go-sql.js/bindings"
)

type SQLJSDriver struct{
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
		db = bindings.OpenReader( d.Reader )
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
	defer func() {
		if r := recover(); r != nil {
			e = r.(*js.Error)
		}
	}()
	if s.Call("free").Bool() {
		return nil
	} else {
		return errors.New("Unknown error freeing statement handle")
	}
}

func (s *SQLJSStmt) NumInput() int {
	return -1
}

func (s *SQLJSStmt) Exec(args []driver.Value) (r driver.Result, e error) {
	defer func() {
		if r := recover(); r != nil {
			e = r.(*js.Error)
		}
	}()
	s.Call("run", args)
	return SQLJSResult{}, nil
}

type SQLJSResult struct{}

func (SQLJSResult) LastInsertId() (int64, error) {
	return 0, errors.New("LastInsertId not available")
}

func (SQLJSResult) RowsAffected() (int64, error) {
	return 0, errors.New("RowsAffected not available")
}

func (s *SQLJSStmt) Query(args []driver.Value) (r driver.Rows, e error) {
	defer func() {
		if r := recover(); r != nil {
			e = r.(*js.Error)
		}
	}()
	if !s.Call("bind", args).Bool() {
		return nil, errors.New("Unexpected error binding values")
	}
	rows := &SQLJSRows{s.Object, false}
	if err := rows.step(); err != nil {
		return nil, err
	}
	return rows, nil
}

type SQLJSRows struct {
	*js.Object
	lastStep bool
}

func (r *SQLJSRows) step() (e error) {
	defer func() {
		if r := recover(); r != nil {
			e = r.(*js.Error)
		}
	}()
	r.lastStep = r.Call("step").Bool()
	return nil
}

func (r *SQLJSRows) Close() (e error) {
	defer func() {
		if r := recover(); r != nil {
			e = r.(*js.Error)
		}
	}()
	r.Call("reset")
	return nil
}

func (r *SQLJSRows) Columns() []string {
	res := r.Call("getColumnNames")
	cols := make([]string, res.Length())
	for i := 0; i < res.Length(); i++ {
		cols[i] = res.Index(i).String()
	}
	return cols
}

func (r *SQLJSRows) Next(dest []driver.Value) (e error) {
	defer func() {
		if r := recover(); r != nil {
			e = r.(*js.Error)
		}
	}()
	if err := r.step(); err != nil {
		return err
	}
	if !r.lastStep {
		return io.EOF
	}
// 	result := r.Call("get")
	return nil
}
