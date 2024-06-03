package handlers

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func NopCloser(r io.Reader) io.ReadCloser {
	return nopCloser{r}
}

func TestGzip(t *testing.T) {

	r := io.NopCloser(strings.NewReader("test gzip"))
	_, err := newCompressReader(r)
	assert.Error(t, err)
}
