package app

import (
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestApp(t *testing.T) {
	shutdown := make(chan error)
	go func() {
		shutdown <- ServerRun()
	}()

	time.Sleep(3 * time.Second)
	syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	errShutdown := <-shutdown

	assert.NoError(t, errShutdown)
}
