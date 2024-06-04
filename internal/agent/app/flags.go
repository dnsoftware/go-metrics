package app

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/dnsoftware/go-metrics/internal/constants"
)

type AgentFlags struct {
	flagRunAddr        string
	flagReportInterval int64
	flagPollInterval   int64
	flagCryptoKey      string
	flagRateLimit      int
	flagAsymPubKeyPath string // путь к файлу с публичным асимметричным ключом
	flagGrpcAddress    string // адрес:порт на котором работает gRPC сервер
	flagServerAPI      string // по какому протоколу клиент будет общаться с сервером (http || grpc) (флаг запуска -server-api, переменная окружения SERVER_API)
}

type Config struct {
	RunAddr        string `env:"ADDRESS"`
	ReportInterval int64  `env:"REPORT_INTERVAL"`
	PollInterval   int64  `env:"POLL_INTERVAL"`
	CryptoKey      string `env:"KEY"`
	RateLimit      int    `env:"RATE_LIMIT"`
	AsymPubKeyPath string `env:"CRYPTO_KEY"`   // путь к файлу с публичным асимметричным ключом
	GrpcAddress    string `env:"GRPC_ADDRESS"` // адрес:порт на котором работает gRPC сервер
	ServerApi      string `env:"SERVER_API"`   // "http" или "grpc"
}

// NewAgentFlags обрабатывает аргументы командной строки
// возвращает соответствующую структуру
// а также проверяет переменные окружения и задействует их при наличии
func NewAgentFlags() AgentFlags {

	var (
		cfg    Config
		cfgEnv Config
		flags  AgentFlags
	)

	// настройки из командной строки
	err := env.Parse(&cfgEnv)
	if err != nil {
		log.Fatal(err)
	}

	// конфиг из файла
	var cFile, configFile string
	flag.StringVar(&cFile, "c", "", "json config file path")
	flag.StringVar(&configFile, "config", "", "json config file path")

	flag.StringVar(&flags.flagRunAddr, "a", "", "address and port to run server")
	flag.Int64Var(&flags.flagReportInterval, "r", 0, "report interval")
	flag.Int64Var(&flags.flagPollInterval, "p", 0, "poll interval")
	flag.StringVar(&flags.flagCryptoKey, "k", "", "crypto key")
	flag.IntVar(&flags.flagRateLimit, "l", constants.RateLimit, "poll interval")
	flag.StringVar(&flags.flagAsymPubKeyPath, "crypto-key", "", "asymmetric crypto key")
	flag.StringVar(&flags.flagGrpcAddress, "g", constants.GRPCDefault, "grpc address")
	flag.StringVar(&flags.flagServerAPI, "server-api", constants.ServerAPI, "server protocol")

	flag.Parse()

	// из конфиг файла
	if cFile != "" {
		configFile = cFile
	}

	jsonConf, _ := newJSONConfig(configFile)
	// объединение конфигураций json, флаги, константы, переменные окружения
	flags = consolidateConfig(jsonConf, cfg, flags, cfgEnv)

	return flags
}

func (f *AgentFlags) RunAddr() string {
	return f.flagRunAddr
}

func (f *AgentFlags) ReportInterval() int64 {
	return f.flagReportInterval
}

func (f *AgentFlags) PollInterval() int64 {
	return f.flagPollInterval
}

func (f *AgentFlags) CryptoKey() string {
	return f.flagCryptoKey
}

func (f *AgentFlags) RateLimit() int {
	return f.flagRateLimit
}

func (f *AgentFlags) AsymPubKeyPath() string {
	return f.flagAsymPubKeyPath
}

func (f *AgentFlags) GrpcRunAddr() string {
	return f.flagGrpcAddress
}
