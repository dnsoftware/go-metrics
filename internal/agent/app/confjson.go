package app

import (
	"encoding/json"
	"os"
	"time"

	"github.com/dnsoftware/go-metrics/internal/logger"
)

type JSONConfig struct {
	Address           string `json:"address"`
	ReportIntervalStr string `json:"report_interval"`
	ReportInterval    int64
	PollIntervalStr   string `json:"poll_interval"`
	PollInterval      int64
	AsymCryptoKey     string `json:"crypto_key"`
}

func newJSONConfig(configFile string) (*JSONConfig, error) {

	conf, err := os.Open(configFile)
	if err != nil {
		logger.Log().Error("json config file open error: " + err.Error())
		return nil, err
	}
	defer conf.Close()

	cfg := JSONConfig{}
	jsonParser := json.NewDecoder(conf)
	if err = jsonParser.Decode(&cfg); err != nil {
		logger.Log().Error("json config parse error error: " + err.Error())
		return nil, err
	}

	d, err := time.ParseDuration(cfg.ReportIntervalStr)
	if err != nil {
		logger.Log().Error("parse report interval error: " + err.Error())
		return nil, err
	}
	cfg.ReportInterval = int64(d.Seconds())

	p, err := time.ParseDuration(cfg.PollIntervalStr)
	if err != nil {
		logger.Log().Error("parse poll interval error: " + err.Error())
		return nil, err
	}
	cfg.PollInterval = int64(p.Seconds())

	return &cfg, err
}
