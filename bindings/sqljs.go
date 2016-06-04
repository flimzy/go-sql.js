// +build js

// Package bindings provides minimal GopherJS bindings around the SQL.js (https://github.com/lovasoa/sql.js)
package bindings

import (
	"bytes"
	"errors"
	"io"

	"github.com/gopherjs/gopherjs/js"
)

type Database struct {
	*js.Object
}

type Statement struct {
	*js.Object
}

// New returns a new database by creating a new one in memory
//
// See http://lovasoa.github.io/sql.js/documentation/class/Database.html#constructor-dynamic
func New() *Database {
	return &Database{js.Global.Get("SQL").Get("Database").New()}
}

// OpenReader opens an existing database, referenced by the passed io.Reader
//
// See http://lovasoa.github.io/sql.js/documentation/class/Database.html#constructor-dynamic
func OpenReader(r io.Reader) *Database {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	db := js.Global.Get("SQL").Get("Database").New([]uint8(buf.Bytes()))
	return &Database{db}
}

func captureError(fn func()) (e error) {
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case *js.Error:
				e = r.(*js.Error)
			case error:
				e = r.(error)
			}
		}
	}()
	fn()
	return nil
}

// Run will execute one or more SQL queries (separated by ';'), ignoring the rows it returns
//
// See http://kripken.github.io/sql.js/documentation/class/Database.html#run-dynamic
func (d *Database) Run(query string) (e error) {
	return captureError(func() {
		d.Call("run", query)
	})
}

// RunParams will execute a single SQL query, along with placeholder parameters, ignoring what it returns
//
// See http://kripken.github.io/sql.js/documentation/class/Database.html#run-dynamic
func (d *Database) RunParams(query string, params []interface{}) (e error) {
	return captureError(func() {
		d.Call("run", query, params)
	})
}

// Export the contents of the database to an io.Reader
//
// See http://kripken.github.io/sql.js/documentation/class/Database.html#export-dynamic
func (d *Database) Export() io.Reader {
	array := d.Call("export").Interface()
	return bytes.NewReader([]byte(array.([]uint8)))
}

// Close the database and all associated prepared statements.
//
// See http://kripken.github.io/sql.js/documentation/class/Database.html#close-dynamic
func (d *Database) Close() (e error) {
	return captureError(func() {
		d.Call("close")
	})
}

func (d *Database) prepare(query string, params interface{}) (*Statement, error) {
	var s *js.Object
	err := captureError(func() {
		s = d.Call("prepare", query, params)
	})
	return &Statement{s}, err
}

// Prepare an SQL statement
//
// See http://kripken.github.io/sql.js/documentation/class/Database.html#prepare-dynamic
func (d *Database) Prepare(query string) (s *Statement, e error) {
	return d.prepare(query, nil)
}

// Prepare an SQL statement, with array of parameters
//
// See http://kripken.github.io/sql.js/documentation/class/Database.html#prepare-dynamic
func (d *Database) PrepareParams(query string, params []interface{}) (s *Statement, e error) {
	return d.prepare(query, params)
}

// Prepare an SQL statement, with named parameters
//
// See http://kripken.github.io/sql.js/documentation/class/Database.html#prepare-dynamic
func (d *Database) PrepareNamedParams(query string, params map[string]interface{}) (s *Statement, e error) {
	return d.prepare(query, params)
}

// GetRowsModified returns the number of rows modified, inserted or deleted by
// the most recently completed INSERT, UPDATE or DELETE statement. Executing
// any other type of SQL statement does not modify the value returned by this
// function.
//
// See http://kripken.github.io/sql.js/documentation/class/Database.html#getRowsModified-dynamic
func (d *Database) GetRowsModified() int64 {
	return d.Call("getRowsModified").Int64()
}

type Result struct {
	Columns []string
	Values  [][]interface{}
}

// Exec will execute an SQL query, and return the result.
//
// This is a wrapper around Database.Prepare(), Statement.Step(),
// Statement.Get(), and Statement.Free().
//
// The result is an array of Result elements. There are as many result elements
// as the number of statements in your sql string (statements are separated by
// a semicolon).
//
// See http://kripken.github.io/sql.js/documentation/class/Database.html#exec-dynamic
func (d *Database) Exec(query string) (r []Result, e error) {
	var result *js.Object
	e = captureError(func() {
		result = d.Call("exec", query)
	})
	if e != nil {
		return
	}
	r = make([]Result, result.Length())
	for i := 0; i < result.Length(); i++ {
		cols := result.Index(i).Get("columns")
		rows := result.Index(i).Get("values")
		r[i].Columns = make([]string, cols.Length())
		for j := 0; j < cols.Length(); j++ {
			r[i].Columns[j] = cols.Index(j).String()
		}
		r[i].Values = make([][]interface{}, rows.Length())
		for j := 0; j < rows.Length(); j++ {
			vals := rows.Index(j)
			r[i].Values[j] = make([]interface{}, vals.Length())
			for k := 0; k < vals.Length(); k++ {
				r[i].Values[j][k] = vals.Index(k).Interface()
			}
		}
	}
	return r, nil
}

// Step executes the statement if necessary, and fetches the next line of the result which
// can be retrieved with Get().
//
// See http://kripken.github.io/sql.js/documentation/class/Statement.html#step-dynamic
func (s *Statement) Step() (ok bool, e error) {
	err := captureError(func() {
		ok = s.Call("step").Bool()
	})
	return ok, err
}

func (s *Statement) get(params interface{}) (r []interface{}, e error) {
	err := captureError(func() {
		results := s.Call("get", params)
		r = make([]interface{}, results.Length())
		for i := 0; i < results.Length(); i++ {
			r[i] = results.Index(i).Interface()
		}
	})
	return r, err
}

// Get one row of results of a statement. Step() must have been called first.
//
// See http://kripken.github.io/sql.js/documentation/class/Statement.html#get-dynamic
func (s *Statement) Get() (r []interface{}, e error) {
	return s.get(nil)
}

// GetParams will get one row of results of a statement after binding the parameters and executing the statement.
//
// See http://kripken.github.io/sql.js/documentation/class/Statement.html#get-dynamic
func (s *Statement) GetParams(params []interface{}) (r []interface{}, e error) {
	return s.get(params)
}

// GetNamedParams will get one row of results of a statement after binding the parameters and executing the statement.
//
// See http://kripken.github.io/sql.js/documentation/class/Statement.html#get-dynamic
func (s *Statement) GetNamedParams(params map[string]interface{}) (r []interface{}, e error) {
	return s.get(params)
}

// GetColumnNames list of column names of a row of result of a statement.
//
// See http://kripken.github.io/sql.js/documentation/class/Statement.html#getColumnNames-dynamic
func (s *Statement) GetColumnNames() (c []string, e error) {
	cols := s.Call("getColumnNames")
	c = make([]string, cols.Length())
	for i := 0; i < cols.Length(); i++ {
		c[i] = cols.Index(i).String()
	}
	return c, nil
}

func (s *Statement) bind(params interface{}) (e error) {
	var tf bool
	err := captureError(func() {
		tf = s.Call("bind", params).Bool()
	})
	if err != nil {
		return err
	}
	if !tf {
		return errors.New("Unknown error binding parameters")
	}
	return nil
}

// Bind values to parameters, after having reset the statement.
//
// See http://kripken.github.io/sql.js/documentation/class/Statement.html#bind-dynamic
func (s *Statement) Bind(params []interface{}) (e error) {
	return s.bind(params)
}

// BindNamed binds values to named parameters, after having reset the statement.
//
// See http://kripken.github.io/sql.js/documentation/class/Statement.html#bind-dynamic
func (s *Statement) BindNamed(params map[string]interface{}) (e error) {
	return s.bind(params)
}

// Reset a statement, so that it's parameters can be bound to new values. It
// also clears all previous bindings, freeing the memory used by bound parameters.
//
// See http://kripken.github.io/sql.js/documentation/class/Statement.html#reset-dynamic
func (s *Statement) Reset() {
	s.Call("reset")
}

// Freemem frees memory allocated during paramater binding.
//
// See http://kripken.github.io/sql.js/documentation/class/Statement.html#freemem-dynamic
func (s *Statement) Freemem() {
	s.Call("freemem")
}

// Free frees any memory used by the statement.
//
// See http://kripken.github.io/sql.js/documentation/class/Statement.html#free-dynamic
func (s *Statement) Free() bool {
	return s.Call("free").Bool()
}

func (s *Statement) getAsMap(params interface{}) (m map[string]interface{}, e error) {
	err := captureError(func() {
		o := s.Call("getAsObject", params)
		m = make(map[string]interface{}, o.Length())
		for _, key := range js.Keys(o) {
			m[key] = o.Get(key).Interface()
		}
	})
	return m, err
}

// GetAsMap will get one row of result as a javascript object, associating
// column names with their value in the current row.
//
// See http://kripken.github.io/sql.js/documentation/class/Statement.html#getAsObject-dynamic
func (s *Statement) GetAsMap() (m map[string]interface{}, e error) {
	return s.getAsMap(nil)
}

// GetAsMapParams will get one row of result as a javascript object, associating
// column names with their value in the current row, after binding the parameters
// and executing the statement
//
// See http://kripken.github.io/sql.js/documentation/class/Statement.html#getAsObject-dynamic
func (s *Statement) GetAsMapParams(params []interface{}) (m map[string]interface{}, e error) {
	return s.getAsMap(params)
}

// GetAsMapNamedParams will get one row of result as a javascript object, associating
// column names with their value in the current row, after binding the parameters
// and executing the statement
//
// See http://kripken.github.io/sql.js/documentation/class/Statement.html#getAsObject-dynamic
func (s *Statement) GetAsMapNamedParams(params map[string]interface{}) (m map[string]interface{}, e error) {
	return s.getAsMap(params)
}

func (s *Statement) run(params interface{}) (e error) {
	return captureError(func() {
		s.Call("run", params)
	})
}

// Run is shorthand for Bind() + Step() + Reset(). Bind the values, execute the
// statement, ignoring the rows it returns, and resets it
//
// See http://kripken.github.io/sql.js/documentation/class/Statement.html#run-dynamic
func (s *Statement) Run() (e error) {
	return s.run(nil)
}

// RunParams is shorthand for Bind() + Step() + Reset(). Bind the values, execute the
// statement, ignoring the rows it returns, and resets it
//
// See http://kripken.github.io/sql.js/documentation/class/Statement.html#run-dynamic
func (s *Statement) RunParams(params []interface{}) (e error) {
	return s.run(params)
}

// RunNamedParams is shorthand for Bind() + Step() + Reset(). Bind the values, execute the
// statement, ignoring the rows it returns, and resets it
//
// See http://kripken.github.io/sql.js/documentation/class/Statement.html#run-dynamic
func (s *Statement) RunNamedParams(params map[string]interface{}) (e error) {
	return s.run(params)
}
