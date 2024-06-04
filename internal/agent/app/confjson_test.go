package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJsonConfig(t *testing.T) {

	configFile := "../../../cmd/agent/config.json"

	_, err := newJSONConfig(configFile)
	assert.NoError(t, err)

	configFile = "../../../cmd/agent/_config.json"
	_, err = newJSONConfig(configFile)
	assert.Error(t, err)

	configFile = "../../../cmd/agent/main.go"
	_, err = newJSONConfig(configFile)
	assert.Error(t, err)

	configFile = "../../../cmd/agent/test_config.json"
	_, err = newJSONConfig(configFile)
	require.Error(t, err)

	configFile = "../../../cmd/agent/test2_config.json"
	_, err = newJSONConfig(configFile)
	assert.Error(t, err)

}

func TestConsolidate(t *testing.T) {
	jsonConf := &JSONConfig{
		Address:           "",
		ReportIntervalStr: "",
		ReportInterval:    0,
		PollIntervalStr:   "",
		PollInterval:      0,
		AsymCryptoKey:     "",
		GrpcAddress:       "",
		ServerAPI:         "",
	}

	cfg := Config{
		RunAddr:        "",
		ReportInterval: 0,
		PollInterval:   0,
		CryptoKey:      "",
		RateLimit:      0,
		AsymPubKeyPath: "",
		GrpcAddress:    "",
		ServerAPI:      "",
	}

	flags := AgentFlags{
		flagRunAddr:        "",
		flagReportInterval: 0,
		flagPollInterval:   0,
		flagCryptoKey:      "",
		flagRateLimit:      0,
		flagAsymPubKeyPath: "",
		flagGrpcAddress:    "",
		flagServerAPI:      "",
	}

	cfgEnv := Config{
		RunAddr:        "",
		ReportInterval: 0,
		PollInterval:   0,
		CryptoKey:      "",
		RateLimit:      0,
		AsymPubKeyPath: "",
		GrpcAddress:    "",
		ServerAPI:      "",
	}

	newFlags := consolidateConfig(jsonConf, cfg, flags, cfgEnv)

	assert.Equal(t, "localhost:8080", newFlags.flagRunAddr)
	assert.Equal(t, int64(10), newFlags.flagReportInterval)
	assert.Equal(t, int64(2), newFlags.flagPollInterval)
	assert.Equal(t, "", newFlags.flagCryptoKey)
	assert.Equal(t, 0, newFlags.flagRateLimit)
	assert.Equal(t, "", newFlags.flagAsymPubKeyPath)
	assert.Equal(t, "127.0.0.1:8090", newFlags.flagGrpcAddress)
	assert.Equal(t, "http", newFlags.flagServerAPI)

	flags.flagRunAddr = ""
	jsonConf.Address = "localhost:8081"
	newFlags = consolidateConfig(jsonConf, cfg, flags, cfgEnv)
	assert.Equal(t, "localhost:8081", newFlags.flagRunAddr)

	flags.flagReportInterval = 0
	jsonConf.ReportInterval = 12
	newFlags = consolidateConfig(jsonConf, cfg, flags, cfgEnv)
	assert.Equal(t, int64(12), newFlags.flagReportInterval)

	flags.flagPollInterval = 0
	jsonConf.PollInterval = 13
	newFlags = consolidateConfig(jsonConf, cfg, flags, cfgEnv)
	assert.Equal(t, int64(13), newFlags.flagPollInterval)

	flags.flagCryptoKey = ""
	jsonConf.AsymCryptoKey = "crkey"
	newFlags = consolidateConfig(jsonConf, cfg, flags, cfgEnv)
	assert.Equal(t, "crkey", newFlags.flagCryptoKey)

	flags.flagGrpcAddress = ""
	jsonConf.GrpcAddress = ":8099"
	newFlags = consolidateConfig(jsonConf, cfg, flags, cfgEnv)
	assert.Equal(t, ":8099", newFlags.flagGrpcAddress)

	flags.flagServerAPI = ""
	jsonConf.ServerAPI = "grpc"
	newFlags = consolidateConfig(jsonConf, cfg, flags, cfgEnv)
	assert.Equal(t, "grpc", newFlags.flagServerAPI)

	jsonConf = nil
	flags.flagGrpcAddress = ""
	newFlags = consolidateConfig(jsonConf, cfg, flags, cfgEnv)
	assert.Equal(t, "127.0.0.1:8090", newFlags.flagGrpcAddress)

	flags.flagServerAPI = ""
	newFlags = consolidateConfig(jsonConf, cfg, flags, cfgEnv)
	assert.Equal(t, "http", newFlags.flagServerAPI)

	cfgEnv = Config{
		RunAddr:        ":8080",
		ReportInterval: 14,
		PollInterval:   14,
		CryptoKey:      "superkey",
		RateLimit:      14,
		AsymPubKeyPath: "/path",
		GrpcAddress:    ":8090",
		ServerAPI:      "grpc",
	}

	newFlags = consolidateConfig(jsonConf, cfg, flags, cfgEnv)
	assert.Equal(t, ":8080", newFlags.flagRunAddr)
	assert.Equal(t, int64(14), newFlags.flagReportInterval)
	assert.Equal(t, int64(14), newFlags.flagPollInterval)
	assert.Equal(t, "superkey", newFlags.flagCryptoKey)
	assert.Equal(t, 14, newFlags.flagRateLimit)
	assert.Equal(t, "/path", newFlags.flagAsymPubKeyPath)
	assert.Equal(t, ":8090", newFlags.flagGrpcAddress)
	assert.Equal(t, "grpc", newFlags.flagServerAPI)

}
