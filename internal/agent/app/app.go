package app

import (
	"github.com/dnsoftware/go-metrics/internal/agent/domain"
	"github.com/dnsoftware/go-metrics/internal/agent/infrastructure"
	storage "github.com/dnsoftware/go-metrics/internal/storage"
)

func AgentRun() {

	parseFlags()

	repository := storage.NewMemStorage()

	sender := infrastructure.NewWebSender("http", flagRunAddr)

	metrics := domain.NewMetrics(&repository, &sender, flagPollInterval, flagReportInterval)
	metrics.Start()

}
