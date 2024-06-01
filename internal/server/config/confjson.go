package config

import (
	"encoding/json"
	"os"
	"time"

	"github.com/dnsoftware/go-metrics/internal/logger"
)

type JSONConfig struct {
	ServerAddress    string `json:"address"`
	RestoreSaved     bool   `json:"restore"`
	StoreIntervalStr string `json:"store_interval"`
	StoreInterval    int64
	FileStoragePath  string `json:"store_file"`
	DatabaseDSN      string `json:"database_dsn"`
	AsymCertKeyPath  string `json:"crypto_cert"`
	AsymPrivKeyPath  string `json:"crypto_key"`
	TrustedSubnet    string `json:"trusted_subnet"`
	GrpcAddress      string `json:"grpc_address"`
}

func newJSONConfigServer(configFile string) (*JSONConfig, error) {

	conf, err := os.Open(configFile)
	if err != nil {
		logger.Log().Error("json config file open error: " + err.Error())
		return nil, err
	}
	defer conf.Close()

	cfg := JSONConfig{}
	jsonParser := json.NewDecoder(conf)
	if err = jsonParser.Decode(&cfg); err != nil {
		logger.Log().Error("server json config parse error error: " + err.Error())
		return nil, err
	}

	d, err := time.ParseDuration(cfg.StoreIntervalStr)
	if err != nil {
		logger.Log().Error("server parse report interval error: " + err.Error())
		return nil, err
	}
	cfg.StoreInterval = int64(d.Seconds())

	return &cfg, err
}
