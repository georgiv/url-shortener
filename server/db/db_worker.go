package db

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type DbConfiguration struct {
	Host     string
	Port     int
	User     string
	Password string
}

type DbWorker interface {
	Persist() error
}

type db struct {
	con *sql.DB
}

func NewDbWorker() (DbWorker, error) {
	con, err := sql.Open("mysql", "cranki:abcd1234@tcp(127.0.0.1:3306)/url_shortener")
	if err != nil {
		return nil, err
	}

	return &db{con: con}, nil
}

func (*db) Persist() error {
	return nil
}
