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
	}

	var (
		cfg   Config
		flags AgentFlags
	)

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&flags.flagRunAddr, "a", constants.ServerDefault, "address and port to run server")
	flag.Int64Var(&flags.flagReportInterval, "r", constants.ReportInterval, "report interval")
	flag.Int64Var(&flags.flagPollInterval, "p", constants.PollInterval, "poll interval")
	flag.StringVar(&flags.flagCryptoKey, "k", "", "crypto key")
	flag.IntVar(&flags.flagRateLimit, "l", constants.RateLimit, "poll interval")

	flag.Parse()

	// переменные окружения
	if cfg.RunAddr != "" {
		flags.flagRunAddr = cfg.RunAddr
	}

	if cfg.ReportInterval != 0 {
		flags.flagReportInterval = cfg.ReportInterval
	}

	if cfg.PollInterval != 0 {
		flags.flagPollInterval = cfg.PollInterval
	}

	if cfg.CryptoKey != "" {
		flags.flagCryptoKey = cfg.CryptoKey
	}

	if cfg.RateLimit != 0 {
		flags.flagRateLimit = cfg.RateLimit
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
