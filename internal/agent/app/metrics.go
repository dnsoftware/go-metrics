package app

import (
	"errors"
	"fmt"

	"github.com/dnsoftware/go-metrics/internal/agent/domain"
	"github.com/dnsoftware/go-metrics/internal/agent/infrastructure"
	"github.com/dnsoftware/go-metrics/internal/constants"
	"github.com/dnsoftware/go-metrics/internal/crypto"
	"github.com/dnsoftware/go-metrics/internal/logger"
	"github.com/dnsoftware/go-metrics/internal/storage"
)

func initMetrics(flags AgentFlags) (*domain.Metrics, error) {
	repository := storage.NewMemStorage()

	publicCryptoKey, err := crypto.MakePublicKey(flags.flagAsymPubKeyPath)
	if err != nil {
		logger.Log().Error(err.Error() + ": " + flags.flagAsymPubKeyPath)
	}

	// в зависимости от конфига общаемся с сервером по http или gRPC
	var sender domain.MetricsSender
	switch flags.flagServerApi {
	case constants.ServerApiHTTP:
		// для нового API - constants.ApplicationJson (для старого - constants.TextPlain)
		sender = infrastructure.NewWebSender("http", &flags, constants.ApplicationJSON, publicCryptoKey)
	case constants.ServerApiGRPC:
		sender, err = infrastructure.NewGRPCSender(&flags, flags.flagAsymPubKeyPath)
		if err != nil {
			mess := fmt.Sprintf("Ошибка инициализации NewGRPCSender: %v", err.Error())
			return nil, errors.New(mess)
		}
	default:
		mess := fmt.Sprintf("Ошибка запуска AgentRun, неверный протокол общения с сервером: %v", flags.flagServerApi)
		return nil, errors.New(mess)
	}

	metrics := domain.NewMetrics(repository, sender, &flags)

	return &metrics, nil
}
