package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

func runServer(ctx context.Context, srv *http.Server, shutdownTimeout time.Duration) error {
	return runServerWithHooks(ctx, shutdownTimeout, srv.ListenAndServe, func(ctx context.Context) error {
		return srv.Shutdown(ctx)
	})
}

func runServerWithHooks(
	ctx context.Context,
	shutdownTimeout time.Duration,
	serve func() error,
	shutdown func(context.Context) error,
) error {
	errCh := make(chan error, 1)
	go func() {
		err := serve()
		if err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server shutdown: %w", err)
		}
		return <-errCh
	}
}
