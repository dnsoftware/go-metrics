package app

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dnsoftware/go-metrics/internal/constants"
)

func TestMetrics(t *testing.T) {
	flags := AgentFlags{flagServerAPI: constants.ServerAPIGRPC}
	_, err := initMetrics(flags)
	assert.NoError(t, err)

	flags = AgentFlags{flagServerAPI: constants.ServerAPIHTTP}
	_, err = initMetrics(flags)
	assert.NoError(t, err)

	flags = AgentFlags{flagServerAPI: "bad"}
	_, err = initMetrics(flags)
	assert.Error(t, err)

	flags = AgentFlags{
		flagServerAPI:      constants.ServerAPIGRPC,
		flagAsymPubKeyPath: "nopath",
	}
	_, err = initMetrics(flags)
	assert.Error(t, err)

}
