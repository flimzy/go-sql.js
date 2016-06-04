This package provides minimal GopherJS bindings around [SQL.js](https://github.com/kripken/sql.js).

The SQL.js API is pretty simple to begin with, so the direct mapping to Go methods is also fairly straight forward.

Database object methods:

SQL.js          | go-sql.js
----------------|-------------------------
constructor     | New(), OpenReader()
exec            | Exec()
each            | --
prepare         | Prepare(), PrepareParams()
export          | Export()
close           | Close()
getRowsModified | GetRowsModified()
create_function | --

Statement object methods:

SQL.js         | go-sql.js
---------------|-------------------------
bind           | Bind(), BindNamed()
step           | Step()
get            | Get(), GetParams(), GetNamedParams()
getColumnNames | GetColumnNames()
getAsObject    | GetAsMap(), GetAsMapParams(), GetAsMapNamedParams()
run            | Run()
reset          | Reset()
freemem        | Freemem()
free           | Free()
