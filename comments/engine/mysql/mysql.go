package mysql

import (
	"database/sql"
	"time"

	"github.com/go-sql-driver/mysql"
	"jacobo.tarrio.org/jtweb/comments/engine"
	"jacobo.tarrio.org/jtweb/comments/engine/genericsql"
)

func NewMysqlEngine(connString string) (engine.Engine, error) {
	cfg, err := mysql.ParseDSN(connString)
	if err != nil {
		return nil, err
	}
	cfg.ParseTime = true
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	return genericsql.NewGenericSqlEngine(db), nil
}
