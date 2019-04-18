package db

import (
	"database/sql"
	"strings"
)

// SQLCommand is a struct that keeps both a sql script and the associated values
// for the script. It is meant to store everything database/sql package needs to
// execute a sql command.
type SQLCommand struct {
	scriptB strings.Builder
	args    []interface{}
}

// Exec executes its sql statement using the given Executioner.
func (sqlStmnt *SQLCommand) Exec(ex Executioner) (sql.Result, error) {
	return ex.Exec(sqlStmnt.Script(), sqlStmnt.args...)
}

// AppendArgs appends the given args to the args slice.
func (sqlStmnt *SQLCommand) AppendArgs(args ...interface{}) {
	sqlStmnt.args = append(sqlStmnt.args, args...)
}

// AppendScript adds the given str to its sql script and returns the length
// of the sql script and a nil error
func (sqlStmnt *SQLCommand) AppendScript(str string) (int, error) {
	return sqlStmnt.scriptB.WriteString(str)
}

// Script returns the sql script. The returned script does not have the values injected
// yet.
func (sqlStmnt *SQLCommand) Script() string {
	return sqlStmnt.scriptB.String()
}

// Args returns a copy of the args slice.
func (sqlStmnt *SQLCommand) Args() []interface{} {
	return sqlStmnt.args[:]
}
