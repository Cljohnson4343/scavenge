package pgsql

import (
	"database/sql"
	"strings"
)

// Executioner is an interface that is needed for database/sql polymorphism
type Executioner interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// Command is a struct that keeps both a sql script and the associated values
// for the script. It is meant to store everything database/sql package needs to
// execute a sql command.
type Command struct {
	scriptB strings.Builder
	args    []interface{}
}

// Exec executes its sql statement using the given Executioner.
func (cmd *Command) Exec(ex Executioner) (sql.Result, error) {
	return ex.Exec(cmd.Script(), cmd.args...)
}

// AppendArgs appends the given args to the args slice.
func (cmd *Command) AppendArgs(args ...interface{}) {
	cmd.args = append(cmd.args, args...)
}

// AppendScript adds the given str to its sql script and returns the length
// of the sql script and a nil error
func (cmd *Command) AppendScript(str string) (int, error) {
	return cmd.scriptB.WriteString(str)
}

// Script returns the sql script. The returned script does not have the values injected
// yet.
func (cmd *Command) Script() string {
	return cmd.scriptB.String()
}

// Args returns a copy of the args slice.
func (cmd *Command) Args() []interface{} {
	return cmd.args[:]
}
