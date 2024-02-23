package app

import (
	"net/http"

	"github.com/dnsoftware/go-metrics/internal/logger"
	"github.com/dnsoftware/go-metrics/internal/server/collector"
	"github.com/dnsoftware/go-metrics/internal/server/config"
	"github.com/dnsoftware/go-metrics/internal/server/handlers"
	"github.com/dnsoftware/go-metrics/internal/storage"
)

func ServerRun() {
	srvLogger := logger.Log()
	defer srvLogger.Sync()

	cfg := config.NewServerConfig()

	backupStorage, err := storage.NewBackupStorage(cfg.FileStoragePath)
	if err != nil {
		panic(err)
	}

	var (
		repo    collector.ServerStorage
		collect *collector.Collector
	)

	repo, err = storage.NewPostgresqlStorage(cfg.DatabaseDSN)
	if err != nil { // значит база НЕ рабочая - используем Memory
		repo = storage.NewMemStorage()
	}

	collect, err = collector.NewCollector(cfg, repo, backupStorage)
	if err != nil {
		panic(err)
	}

	server := handlers.NewHTTPServer(collect, cfg.CryptoKey)

	err = http.ListenAndServe(cfg.ServerAddress, server.Router)
	if err != nil {
		panic(err)
	}
}
