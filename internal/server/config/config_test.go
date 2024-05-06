package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	cfg := NewServerConfig()

	assert.Equal(t, "localhost:8080", cfg.ServerAddress)
	assert.Greater(t, cfg.StoreInterval, int64(0))
	assert.Equal(t, "/tmp/metrics-db.json", cfg.FileStoragePath)

	fmt.Println(cfg)
}
