package storage

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPostgresqlStorage(t *testing.T) {

	ddsn := os.Getenv("DATABASE_DSN")
	if ddsn == "" {
		ddsn = "postgres://praktikum:praktikum@127.0.0.2:5532/praktikum?sslmode=disable"
	}

	pgStorage, err := NewPostgresqlStorage(ddsn)
	assert.NoError(t, err)

	ctx := context.Background()
	err = pgStorage.CreateDatabaseTables(ctx)

	assert.NoError(t, err)

}
