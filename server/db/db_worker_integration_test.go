package db_test

import (
	"os"
	"strings"
	"testing"

	"github.com/georgiv/url-shortener/server/db"
	"github.com/georgiv/url-shortener/testdata"
	_ "github.com/go-sql-driver/mysql"
)

func TestNewWorker(t *testing.T) {
	test := func() {
		db, err := db.NewWorker(7)

		defer db.Shutdown()

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}

	testdata.Execute(t, test)
}

func TestNewWorkerMissigConfig(t *testing.T) {
	test := func() {
		err := os.Chdir("./res")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		db, err := db.NewWorker(7)

		if db != nil {
			t.Errorf("Expected nil, received %v", db)
		}

		if err == nil {
			t.Errorf("Expected error, received nil")
		}

		_, ok := err.(*os.PathError)

		if !ok {
			t.Errorf("Expected *os.PathError, received %T", err)
		}
	}

	testdata.Execute(t, test)
}

func TestNewWorkerWrongCredentials(t *testing.T) {
	test := func() {
		err := os.Chdir("../incorrect")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		db, err := db.NewWorker(7)

		if db != nil {
			t.Errorf("Expected nil, received %v", db)
		}

		if err == nil {
			t.Errorf("Expected error, received nil")
		}
	}

	testdata.Execute(t, test)
}

func TestRegister(t *testing.T) {
	test := func() {
		db, err := db.NewWorker(7)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		defer db.Shutdown()

		id := "cranki"
		url := "http://testurl.com"
		err = db.Register(id, url)
		if err != nil {
			t.Errorf("Expected nil, received %v", err)
		}

		dbId, dbUrl := testdata.GetEntry(t, id)
		if dbId != id {
			t.Errorf("Expected %s, received %s", id, dbId)
		}
		if dbUrl != url {
			t.Errorf("Expected %s, received %s", url, dbUrl)
		}

		testdata.DeleteAll(t)
	}

	testdata.Execute(t, test)
}

func TestRegisterDuplicateId(t *testing.T) {
	test := func() {
		db, err := db.NewWorker(7)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		defer db.Shutdown()

		id := "cranki"
		url := "http://testurl.com"
		url2 := "https://testurl.com"
		err = db.Register(id, url)
		if err != nil {
			t.Errorf("Expected nil, received %v", err)
		}

		err = db.Register(id, url2)
		if err == nil {
			t.Errorf("Expected error, received nil")
		}

		if !strings.Contains(err.Error(), "Error 1062") {
			t.Errorf("Expected error Error 1062, received %v", err)
		}

		testdata.DeleteAll(t)
	}

	testdata.Execute(t, test)
}

func TestRegisterDuplicateUrl(t *testing.T) {
	test := func() {
		db, err := db.NewWorker(7)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		defer db.Shutdown()

		id := "cranki"
		id2 := "tester"
		url := "http://testurl.com"

		err = db.Register(id, url)
		if err != nil {
			t.Errorf("Expected nil, received %v", err)
		}

		err = db.Register(id2, url)
		if err == nil {
			t.Errorf("Expected error, received nil")
		}

		if !strings.Contains(err.Error(), "Error 1062") {
			t.Errorf("Expected error Error 1062, received %v", err)
		}

		testdata.DeleteAll(t)
	}

	testdata.Execute(t, test)
}

func TestFind(t *testing.T) {
	test := func() {
		db, err := db.NewWorker(7)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		defer db.Shutdown()

		id := "cranki"
		url := "http://testurl.com"

		testdata.AddEntry(t, id, url, 604800)

		dbId, dbUrl, err := db.Find("id_to_url", id)
		if dbId != id {
			t.Errorf("Expected %s, received %s", id, dbId)
		}
		if dbUrl != url {
			t.Errorf("Expected %s, received %s", url, dbUrl)
		}

		dbId, dbUrl, err = db.Find("url_to_id", url)
		if dbId != id {
			t.Errorf("Expected %s, received %s", id, dbId)
		}
		if dbUrl != url {
			t.Errorf("Expected %s, received %s", url, dbUrl)
		}

		testdata.DeleteAll(t)
	}

	testdata.Execute(t, test)
}

func TestFindNonExistingEntry(t *testing.T) {
	test := func() {
		db, err := db.NewWorker(7)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		defer db.Shutdown()

		id := "cranki"
		url := "http://testurl.com"

		dbId, dbUrl, err := db.Find("id_to_url", id)
		if dbId != "" {
			t.Errorf("Expected empty string, received %s", dbId)
		}
		if dbUrl != "" {
			t.Errorf("Expected empty string, received %s", dbUrl)
		}

		dbId, dbUrl, err = db.Find("url_to_id", url)
		if dbId != "" {
			t.Errorf("Expected empty string, received %s", dbId)
		}
		if dbUrl != "" {
			t.Errorf("Expected empty string, received %s", dbUrl)
		}
	}

	testdata.Execute(t, test)
}

func TestShutdown(t *testing.T) {
	test := func() {
		db, err := db.NewWorker(7)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		db.Shutdown()

		err = db.Register("cranki", "http://testurl.com")
		if err == nil {
			t.Errorf("Expected error, received nil")
		}
	}

	testdata.Execute(t, test)
}
