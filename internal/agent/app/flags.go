package app

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"log"
)

var (
	flagRunAddr        string
	flagReportInterval int64
	flagPollInterval   int64
)

// parseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func parseFlags() {

	type Config struct {
		RunAddr        string `env:"ADDRESS"`
		ReportInterval int64  `env:"REPORT_INTERVAL"`
		PollInterval   int64  `env:"POLL_INTERVAL"`
	}

	var cfg Config

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.Int64Var(&flagReportInterval, "r", 10, "report interval")
	flag.Int64Var(&flagPollInterval, "p", 2, "poll interval")

	flag.Parse()

	// переменные окружения
	if cfg.RunAddr != "" {
		flagRunAddr = cfg.RunAddr
	}
	if cfg.ReportInterval != 0 {
		flagReportInterval = cfg.ReportInterval
	}
	if cfg.PollInterval != 0 {
		flagPollInterval = cfg.PollInterval
	}

}
