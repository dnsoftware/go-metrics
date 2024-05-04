package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dnsoftware/go-metrics/internal/constants"

	"github.com/dnsoftware/go-metrics/internal/server/config"
)

func setup(t *testing.T) (*PgStorage, error) {
	cfg := config.NewServerConfig()

	ddsn := cfg.DatabaseDSN
	if ddsn == "" {
		ddsn = constants.TestDSN
	}

	pgStorage, err := NewPostgresqlStorage(ddsn)
	require.NoError(t, err)

	ctx := context.Background()
	err = pgStorage.CreateDatabaseTables(ctx)
	require.NoError(t, err)

	return pgStorage, nil
}

func TestNewPostgresqlStorage(t *testing.T) {
	ctx := context.Background()
	db, _ := setup(t)

	db.GetAll(ctx)
}
