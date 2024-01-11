package app

import (
	"github.com/dnsoftware/go-metrics/internal/agent/domain"
	"github.com/dnsoftware/go-metrics/internal/agent/infrastructure"
	storage "github.com/dnsoftware/go-metrics/internal/storage"
)

func AgentRun() {

	repository := storage.NewMemStorage()

	sender := infrastructure.NewWebSender("http", "localhost:8080")

	metrics := domain.NewMetrics(&repository, &sender)
	metrics.Start()

}
