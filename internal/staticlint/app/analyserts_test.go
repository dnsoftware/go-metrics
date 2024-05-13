package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateAnalyzers(t *testing.T) {
	a := Analyzers()
	assert.Equal(t, 200, len(a))
}
