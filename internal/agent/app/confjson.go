package app

import (
	"encoding/json"
	"os"
	"time"

	"github.com/dnsoftware/go-metrics/internal/logger"
)

type jsonParseConfig struct {
	Address        string `json:"address"`
	ReportInterval string `json:"report_interval"`
	PollInterval   string `json:"poll_interval"`
	AsymCryptoKey  string `json:"crypto_key"`
}

type JsonConfig struct {
	Address        string
	ReportInterval int64
	PollInterval   int64
	AsymCryptoKey  string
}

func newJsonConfig(configFile string) (*JsonConfig, error) {

	conf, err := os.Open(configFile)
	if err != nil {
		logger.Log().Error("json config file open error: " + err.Error())
		return nil, err
	}
	defer conf.Close()

	cfgParse := jsonParseConfig{}
	cfg := JsonConfig{}
	jsonParser := json.NewDecoder(conf)
	if err = jsonParser.Decode(&cfgParse); err != nil {
		logger.Log().Error("json config parse error error: " + err.Error())
		return nil, err
	}

	cfg.Address = cfgParse.Address
	d, err := time.ParseDuration(cfgParse.ReportInterval)
	if err != nil {
		logger.Log().Error("parse report interval error: " + err.Error())
		return nil, err
	}
	cfg.ReportInterval = int64(d.Seconds())

	p, err := time.ParseDuration(cfgParse.PollInterval)
	if err != nil {
		logger.Log().Error("parse poll interval error: " + err.Error())
		return nil, err
	}
	cfg.PollInterval = int64(p.Seconds())
	cfg.AsymCryptoKey = cfgParse.AsymCryptoKey

	return &cfg, err
}
