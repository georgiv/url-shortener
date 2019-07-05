package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type DbWorker interface {
	Persist() error
	Shutdown()
}

type db struct {
	con        *sql.DB
	statements map[string]*sql.Stmt
}

type DbConfig struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	User        string `json:"user"`
	Password    string `json:"password"`
	DbName      string `json:"db_name"`
	MaxOpenCons int    `json:"max_open_cons"`
	MaxIdleCons int    `json:"max_idle_cons"`
}

func NewDbWorker() (DbWorker, error) {
	f, err := os.Open("res/db_config.json")
	if err != nil {
		return nil, err
	}

	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var dbConfig DbConfig
	json.Unmarshal(b, &dbConfig)

	con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:%v)/%v", dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.DbName))
	if err != nil {
		return nil, err
	}

	con.SetMaxOpenConns(dbConfig.MaxOpenCons)
	con.SetMaxIdleConns(dbConfig.MaxOpenCons)
	con.SetConnMaxLifetime(time.Hour)

	err = con.Ping()
	if err != nil {
		return nil, err
	}

	return &db{con: con}, nil
}

func (dbWorker *db) Persist() error {
	return nil
}

func (dbWorker *db) Shutdown() {
	log.Println("Shutting down DB pool...")

	err := dbWorker.con.Close()
	if err != nil {
		log.Println(fmt.Sprintf("Shutting down DB pool failed: %v", err))
		return
	}

	log.Println("DB pool successfully shut down")
}

func (dbWorker *db) GetURL(id string) {

}

func (dbWorker *db) prepareURLByIDStmt() (*sql.Stmt, error) {
	return dbWorker.prepareStmt("select original_url from url where id like ?")
}

func (dbWorker *db) prepareIDByURLStmt() (*sql.Stmt, error) {
	return dbWorker.prepareStmt("select id from url where original_url like ?")
}

func (dbWorker *db) prepareStmt(query string) (*sql.Stmt, error) {
	stmt, err := dbWorker.con.Prepare(query)
	if err != nil {
		return nil, err
	}

	return stmt, nil
}

func (dbWorker *db) query(stmt *sql.Stmt, p string) (string, error) {
	var res string
	err := stmt.QueryRow(p).Scan(&res)
	if err != nil {
		return "", err
	}

	return res, nil
}
