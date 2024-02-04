package app

import (
	"github.com/dnsoftware/go-metrics/internal/logger"
	"github.com/dnsoftware/go-metrics/internal/server/collector"
	"github.com/dnsoftware/go-metrics/internal/server/config"
	"github.com/dnsoftware/go-metrics/internal/server/handlers"
	"github.com/dnsoftware/go-metrics/internal/storage"
	"net/http"
)

func ServerRun() {

	srvLogger := logger.Log()
	defer srvLogger.Sync()

	cfg := config.NewServerConfig()

	repository := storage.NewMemStorage()
	backupStorage, err := storage.NewBackupStorage(cfg.FileStoragePath)
	if err != nil {
		panic(err)
	}

	pgStorage, err := storage.NewPostgresqlStorage(cfg.DatabaseDSN)
	if err != nil {
		panic(err)
	}

	collect, err := collector.NewCollector(cfg, repository, backupStorage, pgStorage)
	if err != nil {
		panic(err)
	}

	server := handlers.NewHTTPServer(collect)
	err = http.ListenAndServe(cfg.ServerAddress, server.Router)
	if err != nil {
		panic(err)
	}

}
