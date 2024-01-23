package app

import (
	"flag"
	"github.com/dnsoftware/go-metrics/internal/constants"
	"github.com/dnsoftware/go-metrics/internal/logger"
	"github.com/dnsoftware/go-metrics/internal/server/collector"
	"github.com/dnsoftware/go-metrics/internal/server/handlers"
	"github.com/dnsoftware/go-metrics/internal/storage"
	"net/http"
	"os"
)

func ServerRun() {

	srvLogger := logger.Log()
	defer srvLogger.Sync()

	endpoint := flag.String("a", constants.ServerDefault, "server endpoint")
	flag.Parse()

	runAddr := os.Getenv("ADDRESS")
	if runAddr != "" {
		endpoint = &runAddr
	}

	repository := storage.NewMemStorage()

	collect := collector.NewCollector(&repository)

	server := handlers.NewHTTPServer(collect)
	err := http.ListenAndServe(*endpoint, server.Router)
	if err != nil {
		panic(err)
	}

}
