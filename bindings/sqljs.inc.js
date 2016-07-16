if ( typeof($global.SQL) === 'undefined' ) {
    try {
        $global.SQL = require('sql.js');
    } catch(e) {
        throw("Cannot find global SQL object. Did you load sql.js?");
    }
}
