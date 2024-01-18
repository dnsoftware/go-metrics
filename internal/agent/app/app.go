package app

import (
	"github.com/dnsoftware/go-metrics/internal/agent/domain"
	"github.com/dnsoftware/go-metrics/internal/agent/infrastructure"
	storage "github.com/dnsoftware/go-metrics/internal/storage"
)

func AgentRun() {

	flags := NewAgentFlags()

	repository := storage.NewMemStorage()

	sender := infrastructure.NewWebSender("http", &flags)

	metrics := domain.NewMetrics(&repository, &sender, &flags)
	metrics.Start()

}
