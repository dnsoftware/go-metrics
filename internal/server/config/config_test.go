package config

import (
	"fmt"
	"testing"

	"github.com/dnsoftware/go-metrics/internal/constants"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {

	cfg := NewServerConfig()

	assert.Equal(t, constants.ServerDefault, cfg.ServerAddress)
	assert.Greater(t, cfg.StoreInterval, int64(0))
	assert.Equal(t, constants.FileStoragePath, cfg.FileStoragePath)

	fmt.Println(cfg)
}
