package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dnsoftware/go-metrics/internal/constants"
)

func TestLogger(t *testing.T) {

	lg := Log()
	assert.NotNil(t, lg)

	_, err := createLogger(constants.LogFile, constants.LogLevel)
	assert.NoError(t, err)

	_, err = createLogger("", constants.LogLevel)
	assert.Error(t, err)
}
