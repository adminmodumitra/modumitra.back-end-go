package identity

import (
	"errors"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/outbox"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

const minPasswordLen = 12

type Service struct {
	users    UserRepo
	sessions SessionRepo
	emails   EmailTokenRepo
	tx       TxRunner
	hasher   Hasher
	tokens   Tokens
	mailer   Mailer
	clock    Clock
	ids      IDGen
	cfg      Settings
	outbox   outbox.Writer
}

func NewService(
	users UserRepo,
	sessions SessionRepo,
	emails EmailTokenRepo,
	tx TxRunner,
	hasher Hasher,
	tokens Tokens,
	mailer Mailer,
	clock Clock,
	ids IDGen,
	cfg Settings,
	ob outbox.Writer,
) *Service {
	if cfg.Issuer == "" {
		cfg.Issuer = Issuer
	}
	if cfg.AccessTTL <= 0 {
		cfg.AccessTTL = 15 * time.Minute
	}
	if cfg.RefreshTTL <= 0 {
		cfg.RefreshTTL = 30 * 24 * time.Hour
	}
	if cfg.EmailTokenTTL <= 0 {
		cfg.EmailTokenTTL = 24 * time.Hour
	}
	if cfg.OTPTTL <= 0 {
		cfg.OTPTTL = 10 * time.Minute
	}
	if clock == nil {
		clock = SystemClock{}
	}
	if ids == nil {
		ids = NewIDGen()
	}
	return &Service{
		users:    users,
		sessions: sessions,
		emails:   emails,
		tx:       tx,
		hasher:   hasher,
		tokens:   tokens,
		mailer:   mailer,
		clock:    clock,
		ids:      ids,
		cfg:      cfg,
		outbox:   ob,
	}
}

// NewMemoryService builds a service with in-memory repos and a recording mailer.
func NewMemoryService(cfg Settings) (*Service, *RecordingMailer, *Memory) {
	mem := NewMemory()
	mail := &RecordingMailer{}
	clock := SystemClock{}
	ids := NewIDGen()
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = "test-secret-for-cabinet-studio-hs256"
	}
	tok := NewJWTTokens(cfg.JWTSecret, cfg.Issuer, cfg.AccessTTL, clock)
	svc := NewService(mem.Users(), mem.Sessions(), mem.EmailTokens(), mem, PlainHasher{}, tok, mail, clock, ids, cfg, outbox.NewMemoryWriter())
	return svc, mail, mem
}

func (s *Service) expiresIn() int {
	return int(s.cfg.AccessTTL.Seconds())
}

func mapRepoErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrNotFound) {
		return utils.E(utils.CodeNotFound)
	}
	return utils.From(err)
}
