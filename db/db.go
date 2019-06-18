package db

import (
	"database/sql"
	"fmt"
	"log"

	// TODO look into whether this blank import is necessary. GoLint seems to
	// have a problem with it.
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

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

	err = initStatements(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Shutdown closes all db connections
func Shutdown(db *sql.DB) {

	for _, v := range stmtMap {
		v.Close()
	}

	err := db.Close()
	if err != nil {
		panic(err)
	}

}

// InitDB initializes a db and returns the db. The caller is responible for closing the
// db.
func InitDB(env string) *sql.DB {
	var dbConfig = new(Config)
	dbConfig.DBName = viper.GetString(fmt.Sprintf("database.%s.dbname", env))
	dbConfig.Host = viper.GetString(fmt.Sprintf("database.%s.host", env))
	dbConfig.Port = viper.GetInt(fmt.Sprintf("database.%s.port", env))
	dbConfig.Password = viper.GetString(fmt.Sprintf("database.%s.password", env))
	dbConfig.User = viper.GetString(fmt.Sprintf("database.%s.user", env))

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.DBName)

	database, err := NewDB(psqlInfo)
	if err != nil {
		log.Panicf("Error initializing the db: %s\n", err.Error())
	}

	return database
}
