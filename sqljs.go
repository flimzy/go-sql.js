// +build js

package sqljs

import (
	"bytes"
	"errors"
	"io"

	"database/sql"
	"database/sql/driver"

	"github.com/gopherjs/gopherjs/js"
)

var dbReader io.Reader

func SetDBReader(file io.Reader) {
	dbReader = file
}

type SQLJSDriver struct{}

func init() {
	sql.Register("sqljs", &SQLJSDriver{})
}

// Open will open a new, empty database if dsn is the empty string.
// If the DSN string is non-empty, it will attempt to open the file
// previously provided to the SetDB() function.
func (d *SQLJSDriver) Open(dsn string) (driver.Conn, error) {
	if dsn == "" {
		db := js.Global.Get("SQL").Get("Database").New()
		return &SQLJSConn{db}, nil
	}
	if dbReader == nil {
		return nil, errors.New("You must call SetDBReader() first")
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(dbReader)
	db := js.Global.Get("SQL").Get("Database").New([]uint8(buf.Bytes()))
	dbReader = nil // Make sure we don't accidentally re-use the same DBReader
	return &SQLJSConn{db}, nil
}

type SQLJSConn struct {
	*js.Object
}

func (c *SQLJSConn) Prepare(query string) (d driver.Stmt, e error) {
	defer func() {
		if r := recover(); r != nil {
			e = r.(*js.Error)
		}
	}()
	s := c.Call("prepare", query)
	return &SQLJSStmt{s}, nil
}

func (c *SQLJSConn) Begin() (driver.Tx, error) {
	return nil, nil
}

func (c *SQLJSConn) Close() (e error) {
	defer func() {
		if r := recover(); r != nil {
			e = r.(*js.Error)
		}
	}()
	c.Call("close")
	return nil
}

type SQLJSStmt struct {
	*js.Object
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
	result := r.Call("get")

}
