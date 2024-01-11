package app

import (
	"github.com/dnsoftware/go-metrics/internal/agent/domain"
	"github.com/dnsoftware/go-metrics/internal/agent/infrastructure"
	storage "github.com/dnsoftware/go-metrics/internal/storage"
	"math/rand"
	"time"
)

func AgentRun() {
	rand.Seed(time.Now().UnixNano())

	repository := storage.NewMemStorage()

	sender := infrastructure.NewWebSender("http", "localhost:8080")

	metrics := domain.NewMetrics(&repository, &sender)
	metrics.Start()

}
