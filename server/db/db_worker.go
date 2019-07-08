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
	Find(stmtID string, param string) (string, error)
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

	dbWorker := &db{con: con}

	dbWorker.statements = make(map[string]*sql.Stmt)

	urlByIDStmt, err := dbWorker.prepareStmt("select original_url from url where id like ?")
	if err != nil {
		return nil, err
	}
	dbWorker.statements["id_to_url"] = urlByIDStmt

	idByURLstmt, err := dbWorker.prepareStmt("select id from url where original_url like ?")
	if err != nil {
		return nil, err
	}
	dbWorker.statements["url_to_id"] = idByURLstmt

	return dbWorker, nil
}

func (dbWorker *db) Find(stmtID string, param string) (string, error) {
	stmt, ok := dbWorker.statements[stmtID]
	if !ok {
		return "", nil
	}

	res, err := dbWorker.query(stmt, param)
	if err != nil {
		return "", err
	}

	return res, nil
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

func (dbWorker *db) prepareStmt(query string) (*sql.Stmt, error) {
	stmt, err := dbWorker.con.Prepare(query)
	if err != nil {
		return nil, err
	}

	return stmt, nil
}

func (dbWorker *db) query(stmt *sql.Stmt, param string) (string, error) {
	// var res string
	// err := stmt.QueryRow(param).Scan(&res)
	// if err != nil {
	// 	return "", err
	// }

	// return res, nil

	var res string
	rows, err := stmt.Query(param)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&res)
		if err != nil {
			return "", err
		}
	}

	err = rows.Err()
	if err != nil {
		return "", err
	}

	return res, nil
}
