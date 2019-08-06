package testdata

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
)

func Execute(t *testing.T, f func()) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	err = os.Chdir("../../testdata/correct")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	f()

	err = os.Chdir(cwd)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func GetEntry(t *testing.T, param string) (id string, url string) {
	con, err := sql.Open("mysql", "testuser:abcd1234@tcp(localhost:3306)/url_shortener_test")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer con.Close()

	stmt, err := con.Prepare("SELECT id, original_url FROM url WHERE id LIKE ?")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(param)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&id, &url)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	}

	err = rows.Err()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	return
}

func AddEntry(t *testing.T, id string, url string, expiration int) {
	con, err := sql.Open("mysql", "testuser:abcd1234@tcp(localhost:3306)/url_shortener_test")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer con.Close()

	tx, err := con.Begin()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(fmt.Sprintf("INSERT INTO url(id, original_url, creation_time, expiration_time) VALUES(?, ?, UNIX_TIMESTAMP(), UNIX_TIMESTAMP() + %v)", expiration))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(id, url)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func DeleteAll(t *testing.T) {
	con, err := sql.Open("mysql", "testuser:abcd1234@tcp(localhost:3306)/url_shortener_test")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer con.Close()

	tx, err := con.Begin()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("TRUNCATE TABLE url")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}
