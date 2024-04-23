package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBackupStorage_Save(t *testing.T) {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)

	dbFile := exPath + "/test_backup"
	bs, _ := NewBackupStorage(dbFile)

	testStr := "text data"
	bs.Save(testStr)
	bs.backupFile.Close()

	bs, _ = NewBackupStorage(dbFile)
	backup, err := bs.Load()
	if err != nil {
		fmt.Println(err)
	}

	os.Remove(dbFile)
	assert.Equal(t, testStr, backup)
}
