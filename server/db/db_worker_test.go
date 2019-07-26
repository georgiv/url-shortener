package db_test

import (
	"os"
	"testing"
)

func TestNewWorkerValid(t *testing.T) {
	dir, _ := os.Getwd()
	t.Logf("CWD: %v", dir)
}
