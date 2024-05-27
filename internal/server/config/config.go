// Package config Конфигурация
package config

import (
	"flag"
	"log"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/dnsoftware/go-metrics/internal/constants"
)

// ServerConfig конфигурационные параметры сервера
type ServerConfig struct {
	ServerAddress   string `env:"ADDRESS"`
	StoreInterval   int64  `env:"STORE_INTERVAL" envDefault:"-1"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"none"`
	RestoreSaved    bool   `env:"RESTORE" envDefault:"true"`
	DatabaseDSN     string `env:"DATABASE_DSN" envDefault:""`
	CryptoKey       string `env:"KEY" envDefault:""`
	AsymPrivKeyPath string `env:"CRYPTO_KEY"` // путь к файлу с приватным асимметричным ключом
	TrustedSubnet   string `env:"TRUSTED_SUBNET"`
}

// serverFlags флаги конфигурации
type serverFlags struct {
	serverAddress   string
	storeInterval   int64
	fileStoragePath string
	restoreSaved    bool
	databaseDSN     string
	cryptoKey       string
	asymPrivKeyPath string // путь к файлу с приватным асимметричным ключом
	trustedSubnet   string
}

func NewServerConfig() *ServerConfig {
	cfg := &ServerConfig{}
	sf := serverFlags{}

	err := env.Parse(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// конфиг из файла
	var cFile, configFile string
	flag.StringVar(&cFile, "c", "", "json server config file path")
	flag.StringVar(&configFile, "config", "", "json server config file path")

	flag.StringVar(&sf.serverAddress, "a", "", "server endpoint")
	flag.Int64Var(&sf.storeInterval, "i", 0, "store interval")
	flag.StringVar(&sf.fileStoragePath, "f", "", "file store path")
	flag.BoolVar(&sf.restoreSaved, "r", true, "to restore?")
	flag.StringVar(&sf.databaseDSN, "d", "", "data source name")
	flag.StringVar(&sf.cryptoKey, "k", "", "crypto key")
	flag.StringVar(&sf.asymPrivKeyPath, "crypto-key", constants.CryptoPrivateFilePath, "asymmetric crypto key")
	flag.StringVar(&sf.trustedSubnet, "t", constants.TrustedSubnet, "trusted subnet")
	flag.Parse()

	// из конфиг файла
	if cFile != "" {
		configFile = cFile
	}

	jsonConf, _ := newJSONConfigServer(configFile)
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
		if sf.asymPrivKeyPath == "" {
			sf.asymPrivKeyPath = constants.CryptoPrivateFilePath
		}
		if sf.trustedSubnet == "" {
			sf.trustedSubnet = constants.TrustedSubnet
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

	if cfg.AsymPrivKeyPath == "" {
		cfg.AsymPrivKeyPath = sf.asymPrivKeyPath
	}

	if cfg.TrustedSubnet == "" {
		cfg.TrustedSubnet = sf.trustedSubnet
	}

	return cfg
}
