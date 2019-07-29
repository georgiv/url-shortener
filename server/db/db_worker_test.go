package db_test

import (
	"os"
	"testing"
)

func TestNewWorker(t *testing.T) {
	original, _ := os.Getwd()
	t.Logf("CWD: %v", original)

	os.Chdir("../../testing")
	dir, _ := os.Getwd()
	t.Logf("CWD: %v", dir)

	// db, err := db.NewWorker(0)

	os.Chdir(original)
	dir, _ = os.Getwd()
	t.Logf("CWD: %v", dir)
}
