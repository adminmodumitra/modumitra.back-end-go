package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/config"
	httpapi "github.com/Danish-13-Khan/backend-cabinet/internal/http"
	"github.com/Danish-13-Khan/backend-cabinet/internal/obs"
	"github.com/Danish-13-Khan/backend-cabinet/internal/store"
)

func main() {
	cfg := config.Load()
	obs.InitLogger(cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdownTracer, err := obs.Init(ctx, cfg.ServiceName, cfg.OTLPEndpoint)
	if err != nil {
		slog.Warn("otel init failed; continuing with local traces only", "error", err)
	}
	defer func() {
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if shutdownTracer != nil {
			_ = shutdownTracer(c)
		}
	}()

	st, err := store.Open(ctx, cfg.Datastore)
	if err != nil {
		slog.Error("datastore open", "driver", cfg.Datastore.NormalizedDriver(), "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := st.Close(); err != nil {
			slog.Error("datastore close", "error", err)
		}
	}()

	driver := cfg.Datastore.NormalizedDriver()
	wait := time.NewTimer(0)
	defer wait.Stop()
	for i := 0; i < 20; i++ {
		if i > 0 {
			wait.Reset(500 * time.Millisecond)
			select {
			case <-ctx.Done():
				return
			case <-wait.C:
			}
		}
		if err := st.Ping(ctx); err == nil {
			slog.Info("datastore ping ok", "driver", driver)
			break
		} else {
			slog.Warn("datastore ping failed; retrying", "driver", driver, "error", err, "attempt", i+1)
		}
	}

	engine := httpapi.NewRouter(cfg, st)
	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", cfg.HTTPAddr, "driver", driver)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server", "error", err)
			os.Exit(1)
		}
	}

	c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(c); err != nil {
		slog.Error("http shutdown", "error", err)
	}
}
