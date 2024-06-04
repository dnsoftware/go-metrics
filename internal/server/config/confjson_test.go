package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJsonConfig(t *testing.T) {

	configFile := "../../../cmd/server/config.json"

	_, err := newJSONConfigServer(configFile)
	assert.NoError(t, err)

	_, err = newJSONConfigServer("bad")
	assert.Error(t, err)

	_, err = newJSONConfigServer("./confjson_test.go")
	assert.Error(t, err)

	configFile = "../../../cmd/server/_config.json"
	_, err = newJSONConfigServer(configFile)
	assert.Error(t, err)

}

func TestConsolidate(t *testing.T) {
	jsonConf := &JSONConfig{
		ServerAddress:    "",
		RestoreSaved:     false,
		StoreIntervalStr: "",
		StoreInterval:    0,
		FileStoragePath:  "",
		DatabaseDSN:      "",
		AsymCertKeyPath:  "",
		AsymPrivKeyPath:  "",
		TrustedSubnet:    "",
		GrpcAddress:      "",
	}

	cfg := &ServerConfig{
		ServerAddress:   "",
		StoreInterval:   0,
		FileStoragePath: "",
		RestoreSaved:    false,
		DatabaseDSN:     "",
		CryptoKey:       "",
		AsymCertKeyPath: "",
		AsymPrivKeyPath: "",
		TrustedSubnet:   "",
		GrpcAddress:     "",
	}

	sf := serverFlags{
		serverAddress:   "",
		storeInterval:   0,
		fileStoragePath: "",
		restoreSaved:    false,
		databaseDSN:     "",
		cryptoKey:       "",
		asymCertKeyPath: "",
		asymPrivKeyPath: "",
		trustedSubnet:   "",
		grpcAddress:     "",
	}

	cfg = consolidateConfigServer(jsonConf, cfg, sf)

	assert.Equal(t, "localhost:8080", cfg.ServerAddress)

	jsonConf.ServerAddress = ":8080"
	jsonConf.StoreInterval = 11
	jsonConf.FileStoragePath = "path"
	cfg = consolidateConfigServer(jsonConf, cfg, sf)
	assert.Equal(t, ":8080", cfg.ServerAddress)
	assert.Equal(t, int64(11), cfg.StoreInterval)
	assert.Equal(t, "path", cfg.FileStoragePath)

	jsonConf.RestoreSaved = true
	cfg = consolidateConfigServer(jsonConf, cfg, sf)
	assert.Equal(t, false, cfg.RestoreSaved)
	jsonConf.DatabaseDSN = "dsn"
	cfg = consolidateConfigServer(jsonConf, cfg, sf)
	assert.Equal(t, "dsn", cfg.DatabaseDSN)
	jsonConf.AsymCertKeyPath = "key"
	cfg = consolidateConfigServer(jsonConf, cfg, sf)
	assert.Equal(t, "key", cfg.AsymCertKeyPath)
	jsonConf.AsymPrivKeyPath = "key"
	cfg = consolidateConfigServer(jsonConf, cfg, sf)
	assert.Equal(t, "key", cfg.AsymPrivKeyPath)
	jsonConf.TrustedSubnet = "127.0.0.1/24"
	cfg = consolidateConfigServer(jsonConf, cfg, sf)
	assert.Equal(t, "127.0.0.1/24", cfg.TrustedSubnet)
	jsonConf.GrpcAddress = ":8090"
	cfg = consolidateConfigServer(jsonConf, cfg, sf)
	assert.Equal(t, ":8090", cfg.GrpcAddress)

	jsonConf = nil
	sf.restoreSaved = false
	cfg = consolidateConfigServer(jsonConf, cfg, sf)
	assert.Equal(t, true, cfg.RestoreSaved)
	sf.grpcAddress = ""
	cfg = consolidateConfigServer(jsonConf, cfg, sf)
	assert.Equal(t, ":8090", cfg.GrpcAddress)
}
