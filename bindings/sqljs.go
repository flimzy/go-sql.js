// +build js

package sqljs

import (
	"bytes"
	"github.com/gopherjs/gopherjs/js"
	"io"
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
// See http://lovasoa.github.io/sql.js/documentation/class/Database.html#run-dynamic
func (d *Database) Run(query string) (e error) {
	return captureError(func() {
		d.Call("run", query)
	})
}

// RunParams will execute a single SQL query, along with placeholder parameters, ignoring what it returns
//
// See http://lovasoa.github.io/sql.js/documentation/class/Database.html#run-dynamic
func (d *Database) RunParams(query string, params []interface{}) (e error) {
	return captureError(func() {
		d.Call("run", query, params)
	})
}

// Export the contents of the database to an io.Reader
//
// See http://lovasoa.github.io/sql.js/documentation/class/Database.html#export-dynamic
func (d *Database) Export() io.Reader {
	array := d.Call("export").Interface()
	return bytes.NewReader([]byte(array.([]uint8)))
}

// Close the database and all associated prepared statements.
//
// See http://lovasoa.github.io/sql.js/documentation/class/Database.html#close-dynamic
func (d *Database) Close() (e error) {
	return captureError(func() {
		d.Call("close")
	})
}

// Prepare an SQL statement
//
// See http://lovasoa.github.io/sql.js/documentation/class/Database.html#prepare-dynamic
func (d *Database) Prepare(query string) (s *Statement, e error) {
	var stmt *js.Object
	err := captureError(func() {
		stmt = d.Call("prepare", query)
	})
	return &Statement{stmt}, err
}

// Prepare an SQL statement
//
// See http://lovasoa.github.io/sql.js/documentation/class/Database.html#prepare-dynamic
func (d *Database) PrepareParams(query string, params []interface{}) (s *Statement, e error) {
	var stmt *js.Object
	err := captureError(func() {
		stmt = d.Call("prepare", query, params)
	})
	return &Statement{stmt}, err
}

// Unimplemented Database methods:
// exec(sql) http://lovasoa.github.io/sql.js/documentation/class/Database.html#exec-dynamic
// each(sql, params, callback, done) http://lovasoa.github.io/sql.js/documentation/class/Database.html#each-dynamic

// Step executes the statement if necessary, and fetches the next line of the result which
// can be retrieved with Get().
//
// See http://lovasoa.github.io/sql.js/documentation/class/Statement.html#step-dynamic
func (s *Statement) Step() (tf bool, e error) {
	var success bool
	err := captureError(func() {
		success = s.Call("step").Bool()
	})
	return success, err
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
// See http://lovasoa.github.io/sql.js/documentation/class/Statement.html#get-dynamic
func (s *Statement) Get() (r []interface{}, e error) {
	return s.get(nil)
}

// GetParams will get one row of results of a statement after binding the parameters and executing the statement.
//
// See http://lovasoa.github.io/sql.js/documentation/class/Statement.html#get-dynamic
func (s *Statement) GetParams(params []interface{}) (r []interface{}, e error) {
	return s.get(params)
}

// GetNamedParams will get one row of results of a statement after binding the parameters and executing the statement.
//
// See http://lovasoa.github.io/sql.js/documentation/class/Statement.html#get-dynamic
func (s *Statement) GetNamedParams(params map[string]interface{}) (r []interface{}, e error) {
	return s.get(params)
}

// GetColumnNames list of column names of a row of result of a statement.
//
// See http://lovasoa.github.io/sql.js/documentation/class/Statement.html#getColumnNames-dynamic
func (s *Statement) GetColumnNames() (c []string, e error) {
	cols := s.Call("getColumnNames")
	c = make([]string, cols.Length())
	for i := 0; i < cols.Length(); i++ {
		c[i] = cols.Index(i).String()
	}
	return c, nil
}

// Bind values to parameters, after having reset the statement.
//
// See http://lovasoa.github.io/sql.js/documentation/class/Statement.html#bind-dynamic
func (s *Statement) Bind(params []interface{}) (tf bool, e error) {
	err := captureError(func() {
		tf = s.Call("bind", params).Bool()
	})
	return tf, err
}

// BindName binds values to named parameters, after having reset the statement.
//
// See http://lovasoa.github.io/sql.js/documentation/class/Statement.html#bind-dynamic
func (s *Statement) BindName(params map[string]interface{}) (tf bool, e error) {
	err := captureError(func() {
		tf = s.Call("bind", params).Bool()
	})
	return tf, err
}

// Reset a statement, so that it's parameters can be bound to new values. It
// also clears all previous bindings, freeing the memory used by bound parameters.
//
// See http://lovasoa.github.io/sql.js/documentation/class/Statement.html#reset-dynamic
func (s *Statement) Reset() {
	s.Call("reset")
}

func (s *Statement) Freemem() {
	s.Call("freemem")
}

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
// See http://lovasoa.github.io/sql.js/documentation/class/Statement.html#getAsObject-dynamic
func (s *Statement) GetAsMap() (m map[string]interface{}, e error) {
	return s.getAsMap(nil)
}

// GetAsMapParams will get one row of result as a javascript object, associating
// column names with their value in the current row, after binding the parameters
// and executing the statement
//
// See http://lovasoa.github.io/sql.js/documentation/class/Statement.html#getAsObject-dynamic
func (s *Statement) GetAsMapParams(params []interface{}) (m map[string]interface{}, e error) {
	return s.getAsMap(params)
}

// GetAsMapNamedParams will get one row of result as a javascript object, associating
// column names with their value in the current row, after binding the parameters
// and executing the statement
//
// See http://lovasoa.github.io/sql.js/documentation/class/Statement.html#getAsObject-dynamic
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
// See http://lovasoa.github.io/sql.js/documentation/class/Statement.html#run-dynamic
func (s *Statement) Run() (e error) {
	return s.run(nil)
}

// RunParams is shorthand for Bind() + Step() + Reset(). Bind the values, execute the
// statement, ignoring the rows it returns, and resets it
//
// See http://lovasoa.github.io/sql.js/documentation/class/Statement.html#run-dynamic
func (s *Statement) RunParams(params []interface{}) (e error) {
	return s.run(params)
}

// RunNamedParams is shorthand for Bind() + Step() + Reset(). Bind the values, execute the
// statement, ignoring the rows it returns, and resets it
//
// See http://lovasoa.github.io/sql.js/documentation/class/Statement.html#run-dynamic
func (s *Statement) RunNamedParams(params map[string]interface{}) (e error) {
	return s.run(params)
}
