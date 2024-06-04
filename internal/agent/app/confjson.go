package app

import (
	"encoding/json"
	"os"
	"time"

	"github.com/dnsoftware/go-metrics/internal/constants"

	"github.com/dnsoftware/go-metrics/internal/logger"
)

type JSONConfig struct {
	Address           string `json:"address"`
	ReportIntervalStr string `json:"report_interval"`
	ReportInterval    int64
	PollIntervalStr   string `json:"poll_interval"`
	PollInterval      int64
	AsymCryptoKey     string `json:"crypto_key"`
	GrpcAddress       string `json:"grpc_address"`
	ServerApi         string `json:"server_api"` // по какому протоколу клиент будет общаться с сервером (http || grpc) (флаг запуска -server-api, переменная окружения SERVER_API)
}

func newJSONConfig(configFile string) (*JSONConfig, error) {

	if configFile == "" {
		return nil, nil
	}

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

// объединение конфигураций
func consolidateConfig(jsonConf *JSONConfig, cfg Config, flags AgentFlags, cfgEnv Config) AgentFlags {
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

		if jsonConf.GrpcAddress != "" {
			cfg.GrpcAddress = jsonConf.GrpcAddress
		} else {
			cfg.GrpcAddress = constants.GRPCDefault
		}

		if jsonConf.ServerApi != "" {
			cfg.ServerAPI = jsonConf.ServerApi
		} else {
			cfg.ServerAPI = constants.ServerAPI
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
		if flags.flagGrpcAddress == "" {
			flags.flagGrpcAddress = cfg.GrpcAddress
		}
		if flags.flagServerAPI == "" {
			flags.flagServerAPI = cfg.ServerAPI
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
		if flags.flagGrpcAddress == "" {
			flags.flagGrpcAddress = constants.GRPCDefault
		}
		if flags.flagServerAPI == "" {
			flags.flagServerAPI = constants.ServerAPI
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

	if cfgEnv.GrpcAddress != "" {
		flags.flagGrpcAddress = cfgEnv.GrpcAddress
	}

	if cfgEnv.ServerAPI != "" {
		flags.flagServerAPI = cfgEnv.ServerAPI
	}

	return flags
}
