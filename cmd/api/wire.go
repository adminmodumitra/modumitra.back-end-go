package main

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/Danish-13-Khan/backend-cabinet/internal/billing"
	billingpg "github.com/Danish-13-Khan/backend-cabinet/internal/billing/postgres"
	"github.com/Danish-13-Khan/backend-cabinet/internal/config"
	"github.com/Danish-13-Khan/backend-cabinet/internal/device"
	devicepg "github.com/Danish-13-Khan/backend-cabinet/internal/device/postgres"
	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	identpg "github.com/Danish-13-Khan/backend-cabinet/internal/identity/postgres"
	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
	orgpg "github.com/Danish-13-Khan/backend-cabinet/internal/org/postgres"
	"github.com/Danish-13-Khan/backend-cabinet/internal/orglookup"
	"github.com/Danish-13-Khan/backend-cabinet/internal/outbox"
	"github.com/Danish-13-Khan/backend-cabinet/internal/sso"
	ssopg "github.com/Danish-13-Khan/backend-cabinet/internal/sso/postgres"
)

type wired struct {
	ident   *identity.Service
	orgs    *org.Service
	devices *device.Service
	bills   *billing.Service
	sso     *sso.Service
	pub     outbox.Publisher
}

func wire(ctx context.Context, cfg config.Config, gdb *gorm.DB) wired {
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
		orglookup.Adapter{Service: orgs},
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
	return wired{ident: ident, orgs: orgs, devices: devices, bills: bills, sso: ssoSvc, pub: pub}
}
