package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJsonConfig(t *testing.T) {

	configFile := "../../../cmd/server/config.json"

	_, err := newJSONConfigServer(configFile)
	assert.NoError(t, err)

}
