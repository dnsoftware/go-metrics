package app

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/dnsoftware/go-metrics/internal/agent/domain"
	"github.com/dnsoftware/go-metrics/internal/agent/infrastructure"
	"github.com/dnsoftware/go-metrics/internal/constants"
	"github.com/dnsoftware/go-metrics/internal/storage"
)

func AgentRun() {
	flags := NewAgentFlags()

	repository := storage.NewMemStorage()

	// для нового API - constants.ApplicationJson (для старого - constants.TextPlain)
	sender := infrastructure.NewWebSender("http", &flags, constants.ApplicationJSON)

	metrics := domain.NewMetrics(repository, &sender, &flags)

	go func() {
		http.ListenAndServe(constants.AgentPprofAddr, nil) // запускаем сервер для pprof
	}()

	metrics.Start()
}
