[![Build Status](https://travis-ci.org/flimzy/go-sql.js.svg?branch=master)](https://travis-ci.org/flimzy/go-sql.js) [![GoDoc](https://godoc.org/github.com/flimzy/go-sql.js?status.png)](http://godoc.org/github.com/flimzy/go-sql.js)

This package provides [GopherJS](http://www.gopherjs.org/) bindings around [SQL.js](https://github.com/kripken/sql.js), and a [database/sql/driver](https://golang.org/pkg/database/sql/driver/) implementation, for use with the standard Go Database driver infrastructure.

SQL.js is SQLite compiled to JavaScript through Emscripten, which can run in the browser.  If your goal is to use SQLite from Go, you should use [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3) or [go-sqlite](https://github.com/mxk/go-sqlite) instead. If your intention is to use SQLite within GopherJS running on node.js, you should probably use the [sqlite3](https://www.npmjs.com/package/sqlite3) package instead (which as far as I know, has no GopherJS bindings at the moment).

To be clear: You should only use this package if you are writing code for GopherJS which must run in the browser.

This does not support storing databases on the filesystem--it only supports in-memory databases (which may be imported from binary blobs).  The database/sql driver also does not support transactions (what value would they be in an in-memory, in-browser database, anyway?)

Build instructions
------------------
As this package provides bindings for a JavaScript package, naturally the JavaScript must be installed to successfully use these bindings.  In your GopherJS package, which depends on this one, you can add a `package.json` which includes `sql.js` as a dependency, then run `npm install` prior to building the GopherJS package.
