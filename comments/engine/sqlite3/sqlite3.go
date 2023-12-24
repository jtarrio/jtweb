package sqlite3

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"jacobo.tarrio.org/jtweb/comments/engine"
	"jacobo.tarrio.org/jtweb/comments/engine/genericsql"
)

func NewSqlite3Engine(connString string) (engine.Engine, error) {
	db, err := sql.Open("sqlite3", connString)
	if err != nil {
		return nil, err
	}
	return genericsql.NewGenericSqlEngine(db), nil
}
