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
}

// serverFlags флаги конфигурации
type serverFlags struct {
	serverAddress   string
	storeInterval   int64
	fileStoragePath string
	restoreSaved    bool
	databaseDSN     string
	cryptoKey       string
}

func NewServerConfig() *ServerConfig {
	cfg := &ServerConfig{}
	sf := serverFlags{}

	err := env.Parse(cfg)
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&sf.serverAddress, "a", constants.ServerDefault, "server endpoint")
	flag.Int64Var(&sf.storeInterval, "i", constants.StoreInterval, "store interval")
	flag.StringVar(&sf.fileStoragePath, "f", constants.FileStoragePath, "file store path")
	flag.BoolVar(&sf.restoreSaved, "r", constants.RestoreSaved, "to restore?")
	flag.StringVar(&sf.databaseDSN, "d", "", "data source name")
	flag.StringVar(&sf.cryptoKey, "k", "", "crypto key")
	flag.Parse()

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

	return cfg
}
