package infrastructure

import (
	"net/http"
	"strings"
	"time"

	"github.com/dnsoftware/go-metrics/internal/constants"
)

// retryRequest retriable error HTTP запрос
// durations - срез периодов, через которые делается повторная попытка
func retryRequest(r *http.Request) error {
	client := &http.Client{}
	durations := strings.Split(constants.HTTPAttemtPeriods, ",")

	resp, err := client.Do(r)
	if err != nil {
		for _, duration := range durations {
			d, _ := time.ParseDuration(duration)
			time.Sleep(d)

			respRetry, errRetry := client.Do(r)
			if errRetry == nil {
				respRetry.Body.Close()
				return nil
			}
		}

		return err
	}

	resp.Body.Close()

	return nil
}
