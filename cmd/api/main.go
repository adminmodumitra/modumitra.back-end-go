package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/Danish-13-Khan/backend-cabinet/internal/config"
	httpapi "github.com/Danish-13-Khan/backend-cabinet/internal/http"
	"github.com/Danish-13-Khan/backend-cabinet/internal/obs"
	"github.com/Danish-13-Khan/backend-cabinet/internal/store"
)

// @title Cabinet Studio Backend API
// @version 1.0.0
// @description v1 HTTP API
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by access JWT
func main() {
	cfg := config.Load()
	if err := cfg.Log.Validate(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := cfg.Mail.Validate(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	w, err := obs.OpenSinks(cfg.Log.Sinks)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	obs.InitLoggerWithWriter(w, cfg.Log.Level, cfg.ServiceName)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdownTracer, err := obs.Init(ctx, cfg.ServiceName, cfg.OTLPEndpoint)
	if err != nil {
		log.Warn().Err(err).Msg("otel init failed; continuing with local traces only")
	}
	defer func() {
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if shutdownTracer != nil {
			_ = shutdownTracer(c)
		}
	}()

	st, err := store.OpenAll(ctx, cfg.Datastore)
	if err != nil {
		log.Error().Strs("drivers", cfg.Datastore.EnabledDrivers()).Err(err).Msg("datastore open")
		os.Exit(1)
	}
	defer func() {
		if err := st.Close(); err != nil {
			log.Error().Err(err).Msg("datastore close")
		}
	}()

	drivers := cfg.Datastore.EnabledDrivers()
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
			log.Info().Strs("drivers", drivers).Msg("datastore ping ok")
			break
		} else {
			log.Warn().Strs("drivers", drivers).Err(err).Int("attempt", i+1).Msg("datastore ping failed; retrying")
		}
	}

	gdb := st.Gorm()
	if gdb == nil {
		log.Error().Msg("identity requires postgres; enable SQL in datastore config")
		os.Exit(1)
	}

	svc := wire(ctx, cfg, gdb)
	engine := httpapi.NewRouter(cfg, st, svc.ident, svc.orgs, svc.devices, svc.bills, svc.sso)
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
		log.Info().Str("addr", cfg.HTTPAddr).Strs("drivers", drivers).Msg("listening")
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error().Err(err).Msg("http server")
			os.Exit(1)
		}
	}

	c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(c); err != nil {
		log.Error().Err(err).Msg("http shutdown")
	}
	if closer, ok := svc.pub.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			log.Error().Err(err).Msg("kafka publisher close")
		}
	}
}
