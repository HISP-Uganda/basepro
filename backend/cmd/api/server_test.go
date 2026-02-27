package main

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestRunServerGracefulShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdownSignal := make(chan struct{})
	serve := func() error {
		<-shutdownSignal
		return http.ErrServerClosed
	}
	shutdown := func(context.Context) error {
		close(shutdownSignal)
		return nil
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- runServerWithHooks(ctx, 2*time.Second, serve, shutdown)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("runServerWithHooks returned error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for server shutdown")
	}
}
