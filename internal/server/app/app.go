package app

import (
	"github.com/dnsoftware/go-metrics/internal/server/collector"
	"github.com/dnsoftware/go-metrics/internal/server/handlers"
	"github.com/dnsoftware/go-metrics/internal/storage"
	"net/http"
)

func ServerRun() {

	repository := storage.NewMemStorage()

	collect := collector.NewCollector(&repository)

	server := handlers.NewHTTPServer(collect)
	err := http.ListenAndServe(":8080", server.Router)
	if err != nil {
		panic(err)
	}

}
