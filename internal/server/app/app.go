// Package app Приложение Сервер
package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/dnsoftware/go-metrics/internal/crypto"
	"github.com/dnsoftware/go-metrics/internal/logger"
	"github.com/dnsoftware/go-metrics/internal/server/collector"
	"github.com/dnsoftware/go-metrics/internal/server/config"
	"github.com/dnsoftware/go-metrics/internal/server/handlers"
	"github.com/dnsoftware/go-metrics/internal/storage"
)

func ServerRun() error {

	cfg := config.NewServerConfig()

	backupStorage, err := storage.NewBackupStorage(cfg.FileStoragePath)
	if err != nil {
		return err
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
		return err
	}

	privateCryptoKey, err := crypto.MakePrivateKey(cfg.AsymPrivKeyPath)
	if err != nil {
		logger.Log().Error(err.Error())
	}

	// http server
	server := handlers.NewServer(collect, cfg.CryptoKey, privateCryptoKey, cfg.TrustedSubnet)
	srv := &http.Server{Addr: cfg.ServerAddress, Handler: server.Router}

	// grpc server
	// определяем порт для сервера
	listen, err := net.Listen("tcp", cfg.GrpcAddress)
	if err != nil {
		logger.Log().Fatal(err.Error())
	}
	grpcServer, err := handlers.NewGRPCServer(collect, cfg.CryptoKey, cfg.AsymCertKeyPath, cfg.AsymPrivKeyPath, cfg.TrustedSubnet)
	if err != nil {
		logger.Log().Fatal(err.Error())
	}
	fmt.Println("Сервер gRPC начал работу")
	// получаем запрос gRPC
	go func() {
		if err = grpcServer.Serve(listen); err != nil {
			logger.Log().Fatal(err.Error())
		}
	}()

	// через этот канал сообщим основному потоку, что соединения закрыты
	idleConnsClosed := make(chan struct{})
	// канал для перенаправления прерываний
	// поскольку нужно отловить всего одно прерывание,
	// ёмкости 1 для канала будет достаточно
	sigint := make(chan os.Signal, 1)
	// регистрируем перенаправление прерываний
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	// запускаем горутину обработки пойманных прерываний
	go func() {
		// читаем из канала прерываний
		// поскольку нужно прочитать только одно прерывание,
		// можно обойтись без цикла
		<-sigint
		// получили сигнал os.Interrupt, запускаем процедуру graceful shutdown
		if err = srv.Shutdown(context.Background()); err != nil {
			// ошибки закрытия Listener
			logger.Log().Error("HTTP server Shutdown: " + err.Error())
		}
		fmt.Println("\nhttp server shutdown gracefully")

		// корректное завершение работы gRPC сервера
		grpcServer.GracefulStop()
		fmt.Println("grpc server shutdown gracefully")

		// сообщаем основному потоку,
		// что все сетевые соединения обработаны и закрыты
		close(idleConnsClosed)
	}()

	if err2 := srv.ListenAndServe(); err2 != http.ErrServerClosed {
		logger.Log().Fatal("HTTP server ListenAndServe: " + err2.Error())
	}

	// ждём завершения процедуры graceful shutdown
	<-idleConnsClosed
	// получили оповещение о завершении
	// здесь можно освобождать ресурсы перед выходом,
	// например закрыть соединение с базой данных,
	// закрыть открытые файлы
	fmt.Println("Server Shutdown gracefully")

	return nil // нормальное завершение
}
