package db

import (
	"database/sql"
	"strings"

	_ "github.com/lib/pq"
)

// SQLStatement is a struct that keeps both a sql script and the associated values
// for the script. It is meant to store everything database/sql package needs to
// execute a sql command.
type SQLStatement struct {
	script strings.Builder
	args   []interface{}
}

// Executioner is an interface that is needed for database/sql polymorphism
type Executioner interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// Exec executes its sql statement using the given Executioner.
func (sqlStmnt *SQLStatement) Exec(ex Executioner) (sql.Result, error) {
	return ex.Exec(sqlStmnt.Script(), sqlStmnt.args...)
}

// AppendArgs appends the given args to the args slice.
func (sqlStmnt *SQLStatement) AppendArgs(args ...interface{}) {
	sqlStmnt.args = append(sqlStmnt.args, args...)
}

// AppendScript adds the given str to its sql script and returns the length
// of the sql script and a nil error
func (sqlStmnt *SQLStatement) AppendScript(str string) (int, error) {
	return sqlStmnt.script.WriteString(str)
}

// Script returns the sql script. The returned script does not have the values injected
// yet.
func (sqlStmnt *SQLStatement) Script() string {
	return sqlStmnt.script.String()
}

func (sqlStmnt *SQLStatement) Args() []interface{} {
	return sqlStmnt.args
}

// Config is a custom type to store info used to configure postgresql db
type Config struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
}

// NewDB returns a newly initialized database that uses the given config.
func NewDB(dataSourceName string) (*sql.DB, error) {
	var err error
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
