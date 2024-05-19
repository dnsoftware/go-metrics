package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJsonConfig(t *testing.T) {

	configFile := "../../../cmd/agent/config.json"

	_, err := newJSONConfig(configFile)
	assert.NoError(t, err)

}