// +build js

package sqljs

import (
	"bytes"
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
func New() *Database {
	return &Database{ js.Global.Get("SQL").Get("Database").New() }
}

// OpenReader opens an existing database, referenced by the passed io.Reader
func OpenReader(r io.Reader) *Database {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	db := js.Global.Get("SQL").Get("Database").New( []uint8(buf.Bytes()) )
	return &Database{ db }
}

// Run will execute one or more SQL queries (separated by ';'), ignoring the rows it returns
//
// See also: http://lovasoa.github.io/sql.js/documentation/class/Database.html#run-dynamic
func (d *Database) Run(query string) (e error) {
	defer func() {
		if r := recover(); r != nil {
			e = r.(*js.Error)
		}
	}()
	d.Call("run", query)
	return nil
}

// RunParams will execute a single SQL query, along with placeholder parameters, ignoring what it returns
//
// See also: http://lovasoa.github.io/sql.js/documentation/class/Database.html#run-dynamic
func (d *Database) RunParams(query string, params []interface{}) (e error) {
	defer func() {
		if r := recover(); r != nil {
			e = r.(*js.Error)
		}
	}()
	d.Call("run", query, params)
	return nil
}

func (d *Database) Close() (e error) {
	defer func() {
		if r := recover(); r != nil {
			e = r.(*js.Error)
		}
	}()
	d.Call("close")
	return nil
}

// exec(sql) http://lovasoa.github.io/sql.js/documentation/class/Database.html#exec-dynamic
// each(sql, params, callback, done) http://lovasoa.github.io/sql.js/documentation/class/Database.html#each-dynamic
// prepare(sql, params) http://lovasoa.github.io/sql.js/documentation/class/Database.html#prepare-dynamic
// export() http://lovasoa.github.io/sql.js/documentation/class/Database.html#export-dynamic
