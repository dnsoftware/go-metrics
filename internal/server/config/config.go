// Package config Конфигурация
package config

import (
	"flag"
	"log"

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
	AsymCertKeyPath string `env:"CRYPTO_CERT"` // путь к файлу с публичным асимметричным ключом
	AsymPrivKeyPath string `env:"CRYPTO_KEY"`  // путь к файлу с приватным асимметричным ключом
	TrustedSubnet   string `env:"TRUSTED_SUBNET"`
	GrpcAddress     string `env:"GRPC_ADDRESS"` // адрес:порт на котором работает gRPC сервер
}

// serverFlags флаги конфигурации
type serverFlags struct {
	serverAddress   string
	storeInterval   int64
	fileStoragePath string
	restoreSaved    bool
	databaseDSN     string
	cryptoKey       string
	asymCertKeyPath string // путь к файлу с публичным асимметричным ключом (сертификат)
	asymPrivKeyPath string // путь к файлу с приватным асимметричным ключом
	trustedSubnet   string
	grpcAddress     string // адрес:порт на котором работает gRPC сервер
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
	flag.StringVar(&sf.asymCertKeyPath, "crypto-cert", constants.CryptoPublicFilePath, "asymmetric public crypto key")
	flag.StringVar(&sf.asymPrivKeyPath, "crypto-key", constants.CryptoPrivateFilePath, "asymmetric crypto key")
	flag.StringVar(&sf.trustedSubnet, "t", constants.TrustedSubnet, "trusted subnet")
	flag.StringVar(&sf.grpcAddress, "g", constants.GRPCDefault, "grpc address")
	flag.Parse()

	// из конфиг файла
	if cFile != "" {
		configFile = cFile
	}

	jsonConf, _ := newJSONConfigServer(configFile)
	// объединение конфигураций json, флаги, константы, переменные окружения
	cfg = consolidateConfigServer(jsonConf, cfg, sf)

	return cfg
}
