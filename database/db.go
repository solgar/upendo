package database

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

var (
	databases map[string]*sql.DB = make(map[string]*sql.DB)
)

func ConnectToDb(user, pass, dbname string) (*sql.DB, error) {
	key := user + ":" + pass + "@/" + dbname

	db, ok := databases[key]
	if ok {
		return db, nil
	}

	db, err := sql.Open("mysql", key)
	return db, err
}
