package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/joho/godotenv"

	"github.com/stretchr/testify/require"

	"github.com/dnsoftware/go-metrics/internal/constants"

	"github.com/dnsoftware/go-metrics/internal/server/config"
)

func setup() (*PgStorage, error) {

	godotenv.Load("../../.env_test")

	cfg := config.NewServerConfig()
	ddsn := cfg.DatabaseDSN
	if ddsn == "" {
		ddsn = constants.TestDSN
	}

	pgStorage, err := NewPostgresqlStorage(ddsn)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	err = pgStorage.CreateDatabaseTables(ctx)
	if err != nil {
		return nil, err
	}

	return pgStorage, nil
}

func TestNewPostgresqlStorage(t *testing.T) {
	_, err := setup()
	require.NoError(t, err)
}

func TestSetGauge(t *testing.T) {
	ctx := context.Background()
	pgs, _ := setup()

	var testVal float64 = 123.456
	err := pgs.SetGauge(ctx, "test245", testVal)
	assert.NoError(t, err)

	val, err := pgs.GetGauge(ctx, "test245")
	assert.NoError(t, err)
	assert.Equal(t, testVal, val)
}
