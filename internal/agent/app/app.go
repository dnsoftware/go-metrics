// Package app Приложение Агента
package app

import (
	_ "net/http/pprof"

	"github.com/dnsoftware/go-metrics/internal/logger"
)

func AgentRun() {
	metrics, err := initMetrics(NewAgentFlags())
	if err != nil {
		logger.Log().Fatal(err.Error())
	}

	metrics.Start()
}
