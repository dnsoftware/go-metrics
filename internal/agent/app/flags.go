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
}

// NewAgentFlags обрабатывает аргументы командной строки
// возвращает соответствующую структуру
// а также проверяет переменные окружения и задействует их при наличии
func NewAgentFlags() AgentFlags {
	type Config struct {
		RunAddr        string `env:"ADDRESS"`
		ReportInterval int64  `env:"REPORT_INTERVAL"`
		PollInterval   int64  `env:"POLL_INTERVAL"`
		CryptoKey      string `env:"KEY"`
		RateLimit      int    `env:"RATE_LIMIT"`
		AsymPubKeyPath string `env:"CRYPTO_KEY"` // путь к файлу с публичным асимметричным ключом
	}

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

	flag.Parse()

	// из конфиг файла
	if cFile != "" {
		configFile = cFile
	}

	jsonConf, err := newJsonConfig(configFile)
	if jsonConf != nil {
		if jsonConf.Address != "" {
			cfg.RunAddr = jsonConf.Address
		} else {
			cfg.RunAddr = constants.ServerDefault
		}

		if jsonConf.ReportInterval != 0 {
			cfg.ReportInterval = jsonConf.ReportInterval
		} else {
			cfg.ReportInterval = constants.ReportInterval
		}

		if jsonConf.PollInterval != 0 {
			cfg.PollInterval = jsonConf.PollInterval
		} else {
			cfg.PollInterval = constants.PollInterval
		}

		if jsonConf.AsymCryptoKey != "" {
			cfg.CryptoKey = jsonConf.AsymCryptoKey
		} else {
			cfg.CryptoKey = constants.CryptoPublicFilePath
		}

		if flags.flagRunAddr == "" {
			flags.flagRunAddr = cfg.RunAddr
		}
		if flags.flagReportInterval == 0 {
			flags.flagReportInterval = cfg.ReportInterval
		}
		if flags.flagPollInterval == 0 {
			flags.flagPollInterval = cfg.PollInterval
		}
		if flags.flagCryptoKey == "" {
			flags.flagCryptoKey = cfg.CryptoKey
		}

	} else {
		if flags.flagRunAddr == "" {
			flags.flagRunAddr = constants.ServerDefault
		}
		if flags.flagReportInterval == 0 {
			flags.flagReportInterval = constants.ReportInterval
		}
		if flags.flagPollInterval == 0 {
			flags.flagPollInterval = constants.PollInterval
		}
		if flags.flagCryptoKey == "" {
			flags.flagCryptoKey = constants.CryptoPublicFilePath
		}
	}

	// переменные окружения
	if cfgEnv.RunAddr != "" {
		flags.flagRunAddr = cfgEnv.RunAddr
	}

	if cfgEnv.ReportInterval != 0 {
		flags.flagReportInterval = cfgEnv.ReportInterval
	}

	if cfgEnv.PollInterval != 0 {
		flags.flagPollInterval = cfgEnv.PollInterval
	}

	if cfgEnv.CryptoKey != "" {
		flags.flagCryptoKey = cfgEnv.CryptoKey
	}

	if cfgEnv.RateLimit != 0 {
		flags.flagRateLimit = cfgEnv.RateLimit
	}

	if cfgEnv.AsymPubKeyPath != "" {
		flags.flagAsymPubKeyPath = cfgEnv.AsymPubKeyPath
	}

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
