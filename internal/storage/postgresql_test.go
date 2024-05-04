package storage

import (
	"context"
	"testing"

	"github.com/dnsoftware/go-metrics/internal/server/config"

	"github.com/stretchr/testify/assert"
)

func TestNewPostgresqlStorage(t *testing.T) {

	cfg := config.NewServerConfig()

	//ddsn := os.Getenv("DATABASE_DSN")
	ddsn := cfg.DatabaseDSN
	if ddsn == "" {
		//ddsn = "postgres://praktikum:praktikum@127.0.0.1:5532/praktikum?sslmode=disable"
		ddsn = "postgres://postgres:postgres@postgres:5432/praktikum?sslmode=disable"
	}

	pgStorage, err := NewPostgresqlStorage(ddsn)
	assert.NoError(t, err)

	ctx := context.Background()
	err = pgStorage.CreateDatabaseTables(ctx)

	assert.NoError(t, err)

}
