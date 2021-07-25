package archiver_test

import (
	"backup"
	"testing"
)

func TestBackup(t *testing.T) {
	err := backup.ZIP.Archive("backup", "result")
	if err != nil {
		t.Error(err)
	}
}
