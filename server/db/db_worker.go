// Package db provides interface for managing the
// lifecycle of the URL registrations in the database
// layer.
// The underlying DB is MySQL as configurations are
// specified in the ../../db_config.json file.
//
// Copyright 2019 cranki. All rights reserved.
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

// Worker exports API for selecting and inserting entries
// in the underlying DB.
// It runs additional worker in the background which scans
// the database and cleans up the expired entries.
type Worker interface {
	// Selects entry from the database based in search criteria.
	// Params:
	//   - stmtID: string value specifying a conversion. Possible
	//     values are "id_to_url" (accepts id and returns url)
	//     and "url_to_id" (accepts url and returns url)
	//   - param: id or url. In case stmtID is "id_to_url" then
	//     param should be id. In case stmtID is "url_to_id" then
	//     param should be url.
	// Returns id and url in case a result is found. In case of
	// no match, empty strings are returned along with nil value
	// for an error. In case of an error, it is returned along
	// with empty strings for id and url.
	// If the result is entry which has already expired the return
	// values are empty strings.
	Find(stmtID string, param string) (id string, url string, err error)

	// Inserts new URL alias based on provided id and url. This
	// method does not make preliminary checks if entry with the
	// same id or url exists and in case such is provided it will
	// and error.
	Register(id string, url string) (err error)

	// Closes the DB pool and all statements and perform all
	// necessary cleanups of resources. In case of an error,
	// it is only logged properly, but not returned.
	Shutdown()
}

// NewWorker creates and returns instance satisfying the Worker
// interface
// Params:
//   - expiration: integer representing the period in days for
//     running cleanup service which removes the entries which
//     expired
func NewWorker(expiration int) (worker Worker, err error) {
	f, err := os.Open("res/db_config.json")
	if err != nil {
		return
	}

	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return
	}

	var config config
	err = json.Unmarshal(b, &config)
	if err != nil {
		return
	}

	con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:%v)/%v",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DbName))
	if err != nil {
		return
	}

	con.SetMaxOpenConns(config.MaxOpenCons)
	con.SetMaxIdleConns(config.MaxOpenCons)
	con.SetConnMaxLifetime(time.Hour)

	err = con.Ping()
	if err != nil {
		return
	}

	expirationSec := expiration * 24 * 60 * 60

	dbWorker := &db{con: con, expiration: expirationSec}

	dbWorker.statements = make(map[string]*sql.Stmt)

	urlByIDStmt, err := dbWorker.prepareStmt("SELECT id, original_url, expiration_time FROM url WHERE id LIKE ?")
	if err != nil {
		return
	}
	dbWorker.statements["id_to_url"] = urlByIDStmt

	idByURLstmt, err := dbWorker.prepareStmt("SELECT id, original_url, expiration_time FROM url WHERE original_url LIKE ?")
	if err != nil {
		return
	}
	dbWorker.statements["url_to_id"] = idByURLstmt

	cleaner := time.NewTicker(time.Duration(expirationSec) * time.Second)
	// cleaner = time.NewTicker(5 * time.Second)
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

	worker = dbWorker

	return
}

type db struct {
	con           *sql.DB
	statements    map[string]*sql.Stmt
	expiration    int
	cleanerHandle chan struct{}
}

type config struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	User        string `json:"user"`
	Password    string `json:"password"`
	DbName      string `json:"db_name"`
	MaxOpenCons int    `json:"max_open_cons"`
	MaxIdleCons int    `json:"max_idle_cons"`
}

func (worker *db) Find(stmtID string, param string) (id string, url string, err error) {
	stmt, ok := worker.statements[stmtID]
	if !ok {
		log.Panicf("Missing statement ID: %v", stmtID)
		return
	}

	id, url, expirationTime, err := worker.query(stmt, param)
	if err != nil {
		return
	}

	if expirationTime != 0 {
		now := int(time.Now().Unix())
		if now >= expirationTime {
			log.Printf("Deleting expired entry {%v: %v}...", id, url)
			err = worker.unregister(id)
			if err != nil {
				log.Panicf("Deleting expired entry {%v: %v} failed: %v", id, url, err)
			}

			log.Printf("Expired entry {%v: %v} successfully deleted", id, url)
			return
		}
	}

	return
}

func (worker *db) Register(id string, url string) (err error) {
	tx, err := worker.con.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(fmt.Sprintf("INSERT INTO url(id, original_url, creation_time, expiration_time) VALUES(?, ?, UNIX_TIMESTAMP(), UNIX_TIMESTAMP() + %v)", worker.expiration))
	if err != nil {
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(id, url)
	if err != nil {
		return
	}

	err = tx.Commit()
	if err != nil {
		return
	}

	return
}

func (worker *db) Shutdown() {
	log.Println("Shutting down DB pool...")

	close(worker.cleanerHandle)

	for k, v := range worker.statements {
		err := v.Close()
		if err != nil {
			log.Printf("Closing %v statement failed: %v", k, err)
		}
	}

	err := worker.con.Close()
	if err != nil {
		log.Printf("Shutting down DB pool failed: %v", err)
		return
	}

	log.Println("DB pool successfully shut down")
}

func (worker *db) prepareStmt(query string) (stmt *sql.Stmt, err error) {
	stmt, err = worker.con.Prepare(query)
	return
}

func (worker *db) query(stmt *sql.Stmt, param string) (id string,
	url string,
	expirationTime int,
	err error) {

	rows, err := stmt.Query(param)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&id, &url, &expirationTime)
		if err != nil {
			return
		}
	}

	err = rows.Err()
	if err != nil {
		return
	}

	return
}

func (worker *db) unregister(id string) (err error) {
	tx, err := worker.con.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("DELETE FROM url WHERE id LIKE ?")
	if err != nil {
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return
	}

	err = tx.Commit()
	if err != nil {
		return
	}

	return
}

func (worker *db) clean() {
	var (
		id             string
		url            string
		expirationTime int
	)

	stmt, err := worker.prepareStmt("SELECT id, original_url, expiration_time FROM url")
	if err != nil {
		log.Printf("Error while running cleaner for expired entries: %v", err)
		return
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		log.Printf("Error while running cleaner for expired entries: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&id, &url, &expirationTime)
		if err != nil {
			log.Printf("Error while running cleaner for expired entries: %v", err)
			return
		}

		now := int(time.Now().Unix())
		if now >= expirationTime {
			log.Printf("Deleting expired entry {%v: %v}...", id, url)
			err := worker.unregister(id)
			if err != nil {
				log.Printf("Error while running cleaner for expired entries: %v", err)
				return
			}
		}
	}

	err = rows.Err()
	if err != nil {
		log.Printf("Error while running cleaner for expired entries: %v", err)
		return
	}
}
