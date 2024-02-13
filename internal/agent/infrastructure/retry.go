package infrastructure

import (
	"github.com/dnsoftware/go-metrics/internal/constants"
	"net/http"
	"strings"
	"time"
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
			resp, err = client.Do(r)
			if err == nil {
				break
			}
		}
		if err != nil {
			return err
		}
	}

	resp.Body.Close()

	return nil
}
