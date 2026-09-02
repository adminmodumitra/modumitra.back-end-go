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

	"github.com/Danish-13-Khan/backend-cabinet/internal/billing"
	billingpg "github.com/Danish-13-Khan/backend-cabinet/internal/billing/postgres"
	"github.com/Danish-13-Khan/backend-cabinet/internal/config"
	"github.com/Danish-13-Khan/backend-cabinet/internal/device"
	devicepg "github.com/Danish-13-Khan/backend-cabinet/internal/device/postgres"
	httpapi "github.com/Danish-13-Khan/backend-cabinet/internal/http"
	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	identpg "github.com/Danish-13-Khan/backend-cabinet/internal/identity/postgres"
	"github.com/Danish-13-Khan/backend-cabinet/internal/obs"
	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
	orgpg "github.com/Danish-13-Khan/backend-cabinet/internal/org/postgres"
	"github.com/Danish-13-Khan/backend-cabinet/internal/outbox"
	"github.com/Danish-13-Khan/backend-cabinet/internal/sso"
	ssopg "github.com/Danish-13-Khan/backend-cabinet/internal/sso/postgres"
	"github.com/Danish-13-Khan/backend-cabinet/internal/store"
)

func main() {
	cfg := config.Load()
	if err := cfg.Log.Validate(); err != nil {
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

	obWriter := outbox.NewWriter(gdb)
	clk := identity.SystemClock{}
	ids := identity.NewIDGen()
	ident := identity.NewService(
		identpg.NewUserRepo(gdb),
		identpg.NewSessionRepo(gdb),
		identpg.NewEmailTokenRepo(gdb),
		identpg.NewTx(gdb),
		identity.NewArgon2idHasher(),
		identity.NewJWTTokens(cfg.JWTSecret, identity.Issuer, cfg.AccessTTL, clk),
		identity.NewLogMailer(cfg.Log.Level),
		clk,
		ids,
		identity.Settings{
			TermsVersion:   cfg.TermsVersion,
			PrivacyVersion: cfg.PrivacyVersion,
			AccessTTL:      cfg.AccessTTL,
			RefreshTTL:     cfg.RefreshTTL,
			EmailTokenTTL:  24 * time.Hour,
			JWTSecret:      cfg.JWTSecret,
			Issuer:         identity.Issuer,
		},
		obWriter,
	)

	orgs := org.NewService(
		orgpg.NewOrgRepo(gdb),
		orgpg.NewMemberRepo(gdb),
		orgpg.NewInvitationRepo(gdb),
		identpg.NewUserRepo(gdb),
		orgpg.NewTx(gdb),
		org.NewLogMailer(cfg.Log.Level),
		org.SystemClock{},
		org.NewIDGen(),
		org.NewTokens(),
		org.Settings{TrialDays: cfg.TrialDays},
		obWriter,
	)

	devices := device.NewService(
		devicepg.NewDeviceRepo(gdb),
		devicepg.NewEntitlementRepo(gdb),
		devicepg.NewTx(gdb),
		orgs,
		device.NewHMACSigner(cfg.EntitlementSigningKey, device.KeyIDV1),
		device.SystemClock{},
		device.NewIDGen(),
		obWriter,
	)

	var provider billing.Provider
	if cfg.BillingProvider == billing.ProviderStripe {
		provider = billing.NewStripeProvider()
	} else {
		provider = billing.NewFakeProvider()
	}
	bills := billing.NewService(
		billingpg.NewPlanRepo(gdb),
		billingpg.NewSubscriptionRepo(gdb),
		billingpg.NewInvoiceRepo(gdb),
		billingpg.NewEventRepo(gdb),
		provider,
		orgpg.NewOrgRepo(gdb),
		orgpg.NewMemberRepo(gdb),
		orgpg.NewInvitationRepo(gdb),
		identpg.NewUserRepo(gdb),
		billingpg.NewTx(gdb),
		billing.SystemClock{},
		billing.NewIDGen(),
		billing.Settings{Provider: provider.Name(), WebhookSecret: cfg.BillingWebhookSecret},
		obWriter,
	)

	pub := outbox.NewPublisher(cfg.KafkaEnabled, cfg.KafkaBrokers, cfg.KafkaClientID)
	worker := outbox.NewWorker(outbox.NewStore(gdb), pub, outbox.SystemClock{})
	go worker.Run(ctx)
	log.Info().Bool("kafka_enabled", cfg.KafkaEnabled).Strs("brokers", cfg.KafkaBrokers).Msg("outbox worker started")

	ssoSvc := sso.NewService(
		ssopg.NewIdentityRepo(gdb),
		ssopg.NewStateRepo(gdb),
		ssopg.NewDeviceCodeRepo(gdb),
		ssopg.NewTx(gdb),
		ident,
		sso.LiveProviders(cfg),
		sso.SystemClock{},
		sso.NewIDGen(),
		sso.Settings{
			TermsVersion:    cfg.TermsVersion,
			PrivacyVersion:  cfg.PrivacyVersion,
			PKCESecret:      cfg.JWTSecret,
			VerificationURI: cfg.OAuthDeviceVerifyURI,
			DeviceClientID:  sso.DesktopClientID,
		},
		obWriter,
	)

	engine := httpapi.NewRouter(cfg, st, ident, orgs, devices, bills, ssoSvc)
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
	if closer, ok := pub.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			log.Error().Err(err).Msg("kafka publisher close")
		}
	}
}
