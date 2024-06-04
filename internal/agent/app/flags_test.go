package app

import (
	"fmt"
	"testing"

	"github.com/dnsoftware/go-metrics/internal/constants"
	"github.com/stretchr/testify/assert"
)

func TestRestoreFromDump(t *testing.T) {

	flags := NewAgentFlags()
	fmt.Println(flags)

	assert.Equal(t, constants.ServerDefault, flags.RunAddr())
	assert.Equal(t, constants.ReportInterval, flags.ReportInterval())
	assert.Equal(t, constants.PollInterval, flags.PollInterval())
	assert.Equal(t, "", flags.CryptoKey())
	assert.Equal(t, constants.RateLimit, flags.RateLimit())
	assert.Equal(t, "", flags.AsymPubKeyPath())
}
