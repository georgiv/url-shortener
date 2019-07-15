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
	Find(stmtID string, param string) (string, string, error)
	Register(id string, url string) error
	Shutdown()
}

type db struct {
	con           *sql.DB
	statements    map[string]*sql.Stmt
	expiration    int
	cleanerHandle chan struct{}
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

func NewDbWorker(expiration int) (DbWorker, error) {
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
	err = json.Unmarshal(b, &dbConfig)
	if err != nil {
		return nil, err
	}

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

	expirationSec := expiration * 24 * 60 * 60

	dbWorker := &db{con: con, expiration: expirationSec}

	dbWorker.statements = make(map[string]*sql.Stmt)

	urlByIDStmt, err := dbWorker.prepareStmt("SELECT id, original_url, expiration_time FROM url WHERE id LIKE ?")
	if err != nil {
		return nil, err
	}
	dbWorker.statements["id_to_url"] = urlByIDStmt

	idByURLstmt, err := dbWorker.prepareStmt("SELECT id, original_url, expiration_time FROM url WHERE original_url LIKE ?")
	if err != nil {
		return nil, err
	}
	dbWorker.statements["url_to_id"] = idByURLstmt

	cleaner := time.NewTicker(10 * time.Second)
	cleanerHandle := make(chan struct{})

	dbWorker.cleanerHandle = cleanerHandle

	go func() {
		for {
			select {
			case <-cleaner.C:
				log.Println("Running scheduled cleaner for expired entries...")
				dbWorker.clean()
				log.Println("Scheduled cleaner for expired entries completed")
			case <-cleanerHandle:
				log.Println("Closing cleaner for expired entries...")
				cleaner.Stop()
				log.Println("Cleaner for expired entries successfully closed")
				return
			}
		}
	}()

	return dbWorker, nil
}

func (dbWorker *db) Find(stmtID string, param string) (string, string, error) {
	stmt, ok := dbWorker.statements[stmtID]
	if !ok {
		return "", "", nil
	}

	id, url, expirationTime, err := dbWorker.query(stmt, param)
	if err != nil {
		return "", "", err
	}

	if expirationTime != 0 {
		now := int(time.Now().Unix())
		if now >= expirationTime {
			log.Printf("Deleting expired entry {%v: %v}...", id, url)
			err := dbWorker.unregister(id)
			if err == nil {
				log.Printf("Expired entry {%v: %v} successfully deleted", id, url)
				return "", "", nil
			} else {
				log.Printf("Deleting expired entry {%v: %v} failed: %v", id, url, err)
			}
		}
	}

	return id, url, nil
}

func (dbWorker *db) Register(id string, url string) error {
	tx, err := dbWorker.con.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(fmt.Sprintf("INSERT INTO url(id, original_url, creation_time, expiration_time) VALUES(?, ?, UNIX_TIMESTAMP(), UNIX_TIMESTAMP() + %v)", dbWorker.expiration))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id, url)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (dbWorker *db) Shutdown() {
	log.Println("Shutting down DB pool...")

	close(dbWorker.cleanerHandle)

	for k, v := range dbWorker.statements {
		err := v.Close()
		if err != nil {
			log.Printf("Closing %v statement failed: %v", k, err)
		}
	}

	err := dbWorker.con.Close()
	if err != nil {
		log.Printf("Shutting down DB pool failed: %v", err)
		return
	}

	log.Println("DB pool successfully shut down")
}

func (dbWorker *db) prepareStmt(query string) (*sql.Stmt, error) {
	stmt, err := dbWorker.con.Prepare(query)
	if err != nil {
		return nil, err
	}

	return stmt, nil
}

func (dbWorker *db) query(stmt *sql.Stmt, param string) (string, string, int, error) {
	var (
		id             string
		URL            string
		expirationTime int
	)
	rows, err := stmt.Query(param)
	if err != nil {
		return "", "", 0, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&id, &URL, &expirationTime)
		if err != nil {
			return "", "", 0, err
		}
	}

	err = rows.Err()
	if err != nil {
		return "", "", 0, err
	}

	return id, URL, expirationTime, nil
}

func (dbWorker *db) unregister(id string) error {
	tx, err := dbWorker.con.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("DELETE FROM url WHERE id LIKE ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (dbWorker *db) clean() {
}
