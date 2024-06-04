package app

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dnsoftware/go-metrics/internal/constants"
)

func TestMetrics(t *testing.T) {
	flags := AgentFlags{flagServerApi: constants.ServerApiGRPC}
	_, err := initMetrics(flags)
	assert.NoError(t, err)

	flags = AgentFlags{flagServerApi: constants.ServerApiHTTP}
	_, err = initMetrics(flags)
	assert.NoError(t, err)

	flags = AgentFlags{flagServerApi: "bad"}
	_, err = initMetrics(flags)
	assert.Error(t, err)

	flags = AgentFlags{
		flagServerApi:      constants.ServerApiGRPC,
		flagAsymPubKeyPath: "nopath",
	}
	_, err = initMetrics(flags)
	assert.Error(t, err)

}
