package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/dnsoftware/go-metrics/internal/constants"

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

	if configFile == "" {
		return nil, nil
	}

	conf, err := os.Open(configFile)
	if err != nil {
		logger.Log().Error(fmt.Sprintf("json config file %v open error: %v", configFile, err.Error()))
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

// объединение конфигураций
func consolidateConfigServer(jsonConf *JSONConfig, cfg *ServerConfig, sf serverFlags) *ServerConfig {
	if jsonConf != nil {
		if jsonConf.ServerAddress != "" {
			cfg.ServerAddress = jsonConf.ServerAddress
		} else {
			cfg.ServerAddress = constants.ServerDefault
		}

		if jsonConf.StoreInterval != 0 {
			cfg.StoreInterval = jsonConf.StoreInterval
		} else {
			cfg.StoreInterval = constants.StoreInterval
		}

		if jsonConf.FileStoragePath != "" {
			cfg.FileStoragePath = jsonConf.FileStoragePath
		} else {
			cfg.FileStoragePath = constants.FileStoragePath
		}

		if !jsonConf.RestoreSaved {
			cfg.RestoreSaved = jsonConf.RestoreSaved
		} else {
			cfg.RestoreSaved = constants.RestoreSaved
		}

		if jsonConf.DatabaseDSN != "" {
			cfg.DatabaseDSN = jsonConf.DatabaseDSN
		}

		if jsonConf.AsymCertKeyPath != "" {
			cfg.AsymCertKeyPath = jsonConf.AsymCertKeyPath
		} else {
			cfg.AsymCertKeyPath = constants.CryptoPublicFile
		}

		if jsonConf.AsymPrivKeyPath != "" {
			cfg.AsymPrivKeyPath = jsonConf.AsymPrivKeyPath
		} else {
			cfg.AsymPrivKeyPath = constants.CryptoPrivateFilePath
		}

		if jsonConf.TrustedSubnet != "" {
			cfg.TrustedSubnet = jsonConf.TrustedSubnet
		} else {
			cfg.TrustedSubnet = constants.TrustedSubnet
		}

		if jsonConf.GrpcAddress != "" {
			cfg.GrpcAddress = jsonConf.GrpcAddress
		} else {
			cfg.GrpcAddress = constants.GRPCDefault
		}

	} else {
		if sf.serverAddress == "" {
			sf.serverAddress = constants.ServerDefault
		}
		if sf.storeInterval == 0 {
			sf.storeInterval = constants.StoreInterval
		}
		if sf.fileStoragePath == "" {
			sf.fileStoragePath = constants.FileStoragePath
		}
		if !sf.restoreSaved {
			sf.restoreSaved = constants.RestoreSaved
		}
		if sf.asymCertKeyPath == "" {
			sf.asymCertKeyPath = constants.CryptoPublicFilePath
		}
		if sf.asymPrivKeyPath == "" {
			sf.asymPrivKeyPath = constants.CryptoPrivateFilePath
		}
		if sf.trustedSubnet == "" {
			sf.trustedSubnet = constants.TrustedSubnet
		}
		if sf.grpcAddress == "" {
			sf.grpcAddress = constants.GRPCDefault
		}
	}

	// если какого-то параметра нет в переменных окружения - берем значение флага, а если и флага нет - берем по умолчанию
	if cfg.ServerAddress == "" {
		cfg.ServerAddress = sf.serverAddress
	}

	if cfg.StoreInterval == -1 {
		cfg.StoreInterval = sf.storeInterval
	}

	if cfg.FileStoragePath == "none" {
		cfg.FileStoragePath = sf.fileStoragePath
	}

	if _, ok := os.LookupEnv(constants.RestoreSavedEnv); !ok {
		cfg.RestoreSaved = sf.restoreSaved
	}

	if cfg.DatabaseDSN == "" {
		cfg.DatabaseDSN = sf.databaseDSN
	}

	if cfg.CryptoKey == "" {
		cfg.CryptoKey = sf.cryptoKey
	}

	if cfg.AsymCertKeyPath == "" {
		cfg.AsymCertKeyPath = sf.asymCertKeyPath
	}

	if cfg.AsymPrivKeyPath == "" {
		cfg.AsymPrivKeyPath = sf.asymPrivKeyPath
	}

	if cfg.TrustedSubnet == "" {
		cfg.TrustedSubnet = sf.trustedSubnet
	}

	if cfg.GrpcAddress == "" {
		cfg.GrpcAddress = sf.grpcAddress
	}

	return cfg
}
