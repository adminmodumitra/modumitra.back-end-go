package identity

import (
	"context"
	"errors"
	"log/slog"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

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
	svc := NewService(mem.Users(), mem.Sessions(), mem.EmailTokens(), mem, PlainHasher{}, tok, mail, clock, ids, cfg)
	return svc, mail, mem
}

func (s *Service) expiresIn() int {
	return int(s.cfg.AccessTTL.Seconds())
}

func normalizeEmail(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func validEmail(raw string) bool {
	s := strings.TrimSpace(raw)
	if len(s) < 3 || len(s) > 254 {
		return false
	}
	addr, err := mail.ParseAddress(s)
	if err != nil || addr.Address == "" {
		return false
	}
	return true
}

func validDisplayName(s string) bool {
	s = strings.TrimSpace(s)
	n := utf8.RuneCountInString(s)
	return n >= 1 && n <= 80
}

func validOTP(otp string) bool {
	if len(otp) != 6 {
		return false
	}
	for i := 0; i < 6; i++ {
		if otp[i] < '0' || otp[i] > '9' {
			return false
		}
	}
	return true
}

func (s *Service) Register(ctx context.Context, in RegisterInput) (*User, error) {
	email := strings.TrimSpace(in.Email)
	display := strings.TrimSpace(in.DisplayName)
	if !validEmail(email) || !validDisplayName(display) || in.Password == "" || in.TermsVersion == "" || in.PrivacyVersion == "" {
		return nil, utils.E(utils.CodeValidation)
	}
	if in.TermsVersion != s.cfg.TermsVersion || in.PrivacyVersion != s.cfg.PrivacyVersion {
		return nil, utils.E(utils.CodeTermsOutdated)
	}
	if len(in.Password) < minPasswordLen {
		return nil, utils.E(utils.CodeWeakPassword)
	}

	hash, err := s.hasher.Hash(ctx, in.Password)
	if err != nil {
		return nil, utils.Generic(err)
	}

	now := s.clock.Now()
	u := &User{
		ID:              s.ids.User(),
		Email:           email,
		EmailNormalized: normalizeEmail(email),
		PasswordHash:    hash,
		DisplayName:     display,
		TermsVersion:    in.TermsVersion,
		PrivacyVersion:  in.PrivacyVersion,
		Status:          StatusActive,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	rawTok, tokHash := s.tokens.NewEmailToken(PrefixVerify)
	et := &EmailToken{
		ID:        s.ids.Token(),
		UserID:    u.ID,
		Purpose:   PurposeEmailVerify,
		TokenHash: tokHash,
		ExpiresAt: now.Add(s.cfg.EmailTokenTTL),
		CreatedAt: now,
	}

	err = s.tx.InTx(ctx, func(ctx context.Context) error {
		if err := s.users.Insert(ctx, u); err != nil {
			return err
		}
		return s.emails.Insert(ctx, et)
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}

	if err := s.mailer.SendVerification(ctx, u.Email, rawTok); err != nil {
		slog.ErrorContext(ctx, "send verification mail", "user_id", u.ID, "error", err)
	}
	slog.InfoContext(ctx, "auth.register", "user_id", u.ID)
	out := *u
	out.PasswordHash = ""
	return &out, nil
}

func (s *Service) VerifyEmail(ctx context.Context, raw string) (*User, error) {
	tok, err := s.lookupEmailToken(ctx, raw, PrefixVerify, PurposeEmailVerify)
	if err != nil {
		return nil, err
	}
	now := s.clock.Now()
	var user *User
	err = s.tx.InTx(ctx, func(ctx context.Context) error {
		if err := s.emails.Consume(ctx, tok.ID, now); err != nil {
			return err
		}
		if err := s.users.SetEmailVerified(ctx, tok.UserID, now); err != nil {
			return err
		}
		u, err := s.users.GetByID(ctx, tok.UserID)
		if err != nil {
			return err
		}
		user = u
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, utils.E(utils.CodeInvalidToken)
		}
		return nil, mapRepoErr(err)
	}
	slog.InfoContext(ctx, "auth.email_verified", "user_id", user.ID)
	user.PasswordHash = ""
	return user, nil
}

func (s *Service) ResendVerification(ctx context.Context, email string) error {
	if !validEmail(email) {
		return utils.E(utils.CodeValidation)
	}
	u, err := s.users.GetByEmailNormalized(ctx, normalizeEmail(email))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil
		}
		return mapRepoErr(err)
	}
	if u.EmailVerified() {
		return nil
	}
	now := s.clock.Now()
	rawTok, tokHash := s.tokens.NewEmailToken(PrefixVerify)
	et := &EmailToken{
		ID:        s.ids.Token(),
		UserID:    u.ID,
		Purpose:   PurposeEmailVerify,
		TokenHash: tokHash,
		ExpiresAt: now.Add(s.cfg.EmailTokenTTL),
		CreatedAt: now,
	}
	if err := s.emails.Insert(ctx, et); err != nil {
		return mapRepoErr(err)
	}
	if err := s.mailer.SendVerification(ctx, u.Email, rawTok); err != nil {
		slog.ErrorContext(ctx, "send verification mail", "user_id", u.ID, "error", err)
	}
	return nil
}

func (s *Service) Login(ctx context.Context, in LoginInput) (*User, *TokenPair, error) {
	if !validEmail(in.Email) || in.Password == "" {
		return nil, nil, utils.E(utils.CodeValidation)
	}
	if dn := strings.TrimSpace(in.DeviceName); dn != "" && !validDisplayName(dn) {
		return nil, nil, utils.E(utils.CodeValidation)
	}

	u, err := s.users.GetByEmailNormalized(ctx, normalizeEmail(in.Email))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			s.hasher.DummyVerify(ctx)
			return nil, nil, utils.E(utils.CodeInvalidCredentials)
		}
		return nil, nil, mapRepoErr(err)
	}
	if u.Suspended() {
		return nil, nil, utils.E(utils.CodeAccountSuspended)
	}
	if u.PasswordHash == "" {
		s.hasher.DummyVerify(ctx)
		return nil, nil, utils.E(utils.CodeInvalidCredentials)
	}
	ok, verr := s.hasher.Verify(ctx, u.PasswordHash, in.Password)
	if verr != nil {
		return nil, nil, utils.Generic(verr)
	}
	if !ok {
		return nil, nil, utils.E(utils.CodeInvalidCredentials)
	}

	pair, err := s.issueSession(ctx, u.ID, strings.TrimSpace(in.DeviceName))
	if err != nil {
		return nil, nil, err
	}
	if err := s.users.SetLastLogin(ctx, u.ID, s.clock.Now()); err != nil {
		slog.ErrorContext(ctx, "set last_login_at", "user_id", u.ID, "error", err)
	}
	slog.InfoContext(ctx, "auth.login", "user_id", u.ID)
	u.PasswordHash = ""
	return u, pair, nil
}

// RequestOTP starts (or resends) a one-time login. Unknown emails are a no-op
// so callers can always return 202 without leaking account existence.
func (s *Service) RequestOTP(ctx context.Context, email string) error {
	if !validEmail(email) {
		return utils.E(utils.CodeValidation)
	}
	u, err := s.users.GetByEmailNormalized(ctx, normalizeEmail(email))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil
		}
		return mapRepoErr(err)
	}
	now := s.clock.Now()
	otp := NewOTP()
	et := &EmailToken{
		ID:        s.ids.Token(),
		UserID:    u.ID,
		Purpose:   PurposeOTPLogin,
		TokenHash: s.tokens.HashOTP(u.ID, otp),
		ExpiresAt: now.Add(s.cfg.OTPTTL),
		CreatedAt: now,
	}
	err = s.tx.InTx(ctx, func(ctx context.Context) error {
		if err := s.emails.InvalidateOpen(ctx, u.ID, PurposeOTPLogin, now); err != nil {
			return err
		}
		return s.emails.Insert(ctx, et)
	})
	if err != nil {
		return mapRepoErr(err)
	}
	if err := s.mailer.SendOTP(ctx, u.Email, otp); err != nil {
		slog.ErrorContext(ctx, "send otp mail", "user_id", u.ID, "error", err)
	}
	slog.InfoContext(ctx, "auth.otp_requested", "user_id", u.ID)
	return nil
}

func (s *Service) VerifyOTP(ctx context.Context, email, otp, deviceName string) (*User, *TokenPair, error) {
	if !validEmail(email) {
		return nil, nil, utils.E(utils.CodeValidation)
	}
	if !validOTP(otp) {
		return nil, nil, utils.E(utils.CodeValidation)
	}
	if dn := strings.TrimSpace(deviceName); dn != "" && !validDisplayName(dn) {
		return nil, nil, utils.E(utils.CodeValidation)
	}

	u, err := s.users.GetByEmailNormalized(ctx, normalizeEmail(email))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			s.hasher.DummyVerify(ctx)
			return nil, nil, utils.E(utils.CodeInvalidCredentials)
		}
		return nil, nil, mapRepoErr(err)
	}
	if u.Suspended() {
		return nil, nil, utils.E(utils.CodeAccountSuspended)
	}

	tok, err := s.emails.GetByHash(ctx, s.tokens.HashOTP(u.ID, otp))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, nil, utils.E(utils.CodeInvalidCredentials)
		}
		return nil, nil, mapRepoErr(err)
	}
	if tok.Purpose != PurposeOTPLogin || tok.ConsumedAt != nil || !tok.ExpiresAt.After(s.clock.Now()) || tok.UserID != u.ID {
		return nil, nil, utils.E(utils.CodeInvalidCredentials)
	}

	var pair *TokenPair
	now := s.clock.Now()
	err = s.tx.InTx(ctx, func(ctx context.Context) error {
		if err := s.emails.Consume(ctx, tok.ID, now); err != nil {
			return err
		}
		p, err := s.issueSession(ctx, u.ID, strings.TrimSpace(deviceName))
		if err != nil {
			return err
		}
		pair = p
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, nil, utils.E(utils.CodeInvalidCredentials)
		}
		return nil, nil, mapRepoErr(err)
	}
	if err := s.users.SetLastLogin(ctx, u.ID, now); err != nil {
		slog.ErrorContext(ctx, "set last_login_at", "user_id", u.ID, "error", err)
	}
	slog.InfoContext(ctx, "auth.otp_login", "user_id", u.ID)
	u.PasswordHash = ""
	return u, pair, nil
}

func (s *Service) Refresh(ctx context.Context, raw string) (*TokenPair, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, utils.E(utils.CodeValidation)
	}
	hash := s.tokens.Hash(raw)
	now := s.clock.Now()

	var pair *TokenPair
	err := s.tx.InTx(ctx, func(ctx context.Context) error {
		sess, err := s.sessions.GetByRefreshHashForUpdate(ctx, hash)
		if err != nil {
			return err
		}
		// Rotated/spent token (replaced_by set) or already flagged reuse.
		if sess.ReplacedBy != nil || sess.ReuseDetectedAt != nil {
			if markErr := s.sessions.MarkReuseAndRevokeFamily(ctx, sess.ID, sess.FamilyID, now); markErr != nil {
				return markErr
			}
			return utils.E(utils.CodeTokenReuse)
		}
		if sess.RevokedAt != nil || !sess.ExpiresAt.After(now) {
			return utils.E(utils.CodeUnauthenticated)
		}

		rawRT, rtHash := s.tokens.NewRefresh()
		newSess := &Session{
			ID:          s.ids.Session(),
			UserID:      sess.UserID,
			RefreshHash: rtHash,
			FamilyID:    sess.FamilyID,
			ExpiresAt:   now.Add(s.cfg.RefreshTTL),
			DeviceName:  sess.DeviceName,
			CreatedAt:   now,
			UpdatedAt:   now,
			LastSeenAt:  now,
		}
		if err := s.sessions.Insert(ctx, newSess); err != nil {
			return err
		}
		if err := s.sessions.RevokeAndReplace(ctx, sess.ID, newSess.ID, now); err != nil {
			return err
		}
		access, err := s.tokens.IssueAccess(ctx, sess.UserID, newSess.ID)
		if err != nil {
			return err
		}
		pair = &TokenPair{
			AccessToken:  access,
			RefreshToken: rawRT,
			ExpiresIn:    s.expiresIn(),
			TokenType:    "Bearer",
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, utils.E(utils.CodeUnauthenticated)
		}
		return nil, mapRepoErr(err)
	}
	slog.InfoContext(ctx, "auth.refresh")
	return pair, nil
}

func (s *Service) Logout(ctx context.Context, raw string) error {
	if strings.TrimSpace(raw) == "" {
		return utils.E(utils.CodeValidation)
	}
	hash := s.tokens.Hash(raw)
	sess, err := s.sessions.GetByRefreshHash(ctx, hash)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil
		}
		return mapRepoErr(err)
	}
	if err := s.sessions.Revoke(ctx, sess.ID, s.clock.Now()); err != nil && !errors.Is(err, ErrNotFound) {
		return mapRepoErr(err)
	}
	slog.InfoContext(ctx, "auth.logout", "session_id", sess.ID)
	return nil
}

func (s *Service) LogoutAll(ctx context.Context, userID string) (int, error) {
	if userID == "" {
		return 0, utils.E(utils.CodeUnauthenticated)
	}
	n, err := s.sessions.RevokeAllForUser(ctx, userID, s.clock.Now())
	if err != nil {
		return 0, mapRepoErr(err)
	}
	slog.InfoContext(ctx, "auth.logout_all", "user_id", userID, "revoked", n)
	return n, nil
}

func (s *Service) RequestPasswordReset(ctx context.Context, email string) error {
	if !validEmail(email) {
		return utils.E(utils.CodeValidation)
	}
	u, err := s.users.GetByEmailNormalized(ctx, normalizeEmail(email))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil
		}
		return mapRepoErr(err)
	}
	now := s.clock.Now()
	rawTok, tokHash := s.tokens.NewEmailToken(PrefixReset)
	et := &EmailToken{
		ID:        s.ids.Token(),
		UserID:    u.ID,
		Purpose:   PurposePasswordReset,
		TokenHash: tokHash,
		ExpiresAt: now.Add(s.cfg.EmailTokenTTL),
		CreatedAt: now,
	}
	if err := s.emails.Insert(ctx, et); err != nil {
		return mapRepoErr(err)
	}
	if err := s.mailer.SendPasswordReset(ctx, u.Email, rawTok); err != nil {
		slog.ErrorContext(ctx, "send password reset mail", "user_id", u.ID, "error", err)
	}
	slog.InfoContext(ctx, "auth.password_reset_requested", "user_id", u.ID)
	return nil
}

func (s *Service) ConfirmPasswordReset(ctx context.Context, rawToken, newPassword string) error {
	if strings.TrimSpace(rawToken) == "" {
		return utils.E(utils.CodeInvalidToken)
	}
	if len(newPassword) < minPasswordLen {
		return utils.E(utils.CodeWeakPassword)
	}
	tok, err := s.lookupEmailToken(ctx, rawToken, PrefixReset, PurposePasswordReset)
	if err != nil {
		return err
	}
	hash, err := s.hasher.Hash(ctx, newPassword)
	if err != nil {
		return utils.Generic(err)
	}
	now := s.clock.Now()
	err = s.tx.InTx(ctx, func(ctx context.Context) error {
		if err := s.emails.Consume(ctx, tok.ID, now); err != nil {
			return err
		}
		if err := s.users.UpdatePassword(ctx, tok.UserID, hash, now); err != nil {
			return err
		}
		_, err := s.sessions.RevokeAllForUser(ctx, tok.UserID, now)
		return err
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return utils.E(utils.CodeInvalidToken)
		}
		return mapRepoErr(err)
	}
	slog.InfoContext(ctx, "auth.password_reset_confirm", "user_id", tok.UserID)
	return nil
}

func (s *Service) Me(ctx context.Context, userID string) (*User, error) {
	if userID == "" {
		return nil, utils.E(utils.CodeUnauthenticated)
	}
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, utils.E(utils.CodeNotFound)
		}
		return nil, mapRepoErr(err)
	}
	u.PasswordHash = ""
	return u, nil
}

func (s *Service) ChangePassword(ctx context.Context, userID, current, newPassword string) error {
	if userID == "" {
		return utils.E(utils.CodeUnauthenticated)
	}
	if current == "" || newPassword == "" {
		return utils.E(utils.CodeValidation)
	}
	if len(newPassword) < minPasswordLen || newPassword == current {
		return utils.E(utils.CodeWeakPassword)
	}
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return utils.E(utils.CodeNotFound)
		}
		return mapRepoErr(err)
	}
	if u.PasswordHash == "" {
		s.hasher.DummyVerify(ctx)
		return utils.E(utils.CodeInvalidCredentials)
	}
	ok, verr := s.hasher.Verify(ctx, u.PasswordHash, current)
	if verr != nil {
		return utils.Generic(verr)
	}
	if !ok {
		return utils.E(utils.CodeInvalidCredentials)
	}
	hash, err := s.hasher.Hash(ctx, newPassword)
	if err != nil {
		return utils.Generic(err)
	}
	now := s.clock.Now()
	err = s.tx.InTx(ctx, func(ctx context.Context) error {
		if err := s.users.UpdatePassword(ctx, userID, hash, now); err != nil {
			return err
		}
		_, err := s.sessions.RevokeAllForUser(ctx, userID, now)
		return err
	})
	if err != nil {
		return mapRepoErr(err)
	}
	slog.InfoContext(ctx, "auth.password_change", "user_id", userID)
	return nil
}

// ValidateAccess parses a Bearer JWT, loads the session, and requires it
// not revoked and not expired.
func (s *Service) ValidateAccess(ctx context.Context, raw string) (string, string, error) {
	userID, sessionID, err := s.tokens.ParseAccess(ctx, raw)
	if err != nil {
		return "", "", utils.E(utils.CodeUnauthenticated).With(err)
	}
	sess, err := s.sessions.GetByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", "", utils.E(utils.CodeUnauthenticated)
		}
		return "", "", mapRepoErr(err)
	}
	now := s.clock.Now()
	if sess.RevokedAt != nil || !sess.ExpiresAt.After(now) || sess.UserID != userID {
		return "", "", utils.E(utils.CodeUnauthenticated)
	}
	return userID, sessionID, nil
}

func (s *Service) issueSession(ctx context.Context, userID, deviceName string) (*TokenPair, error) {
	now := s.clock.Now()
	rawRT, rtHash := s.tokens.NewRefresh()
	sess := &Session{
		ID:          s.ids.Session(),
		UserID:      userID,
		RefreshHash: rtHash,
		FamilyID:    s.ids.Family(),
		ExpiresAt:   now.Add(s.cfg.RefreshTTL),
		DeviceName:  deviceName,
		CreatedAt:   now,
		UpdatedAt:   now,
		LastSeenAt:  now,
	}
	if err := s.sessions.Insert(ctx, sess); err != nil {
		return nil, mapRepoErr(err)
	}
	access, err := s.tokens.IssueAccess(ctx, userID, sess.ID)
	if err != nil {
		return nil, err
	}
	return &TokenPair{
		AccessToken:  access,
		RefreshToken: rawRT,
		ExpiresIn:    s.expiresIn(),
		TokenType:    "Bearer",
	}, nil
}

func (s *Service) lookupEmailToken(ctx context.Context, raw, prefix, purpose string) (*EmailToken, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || (prefix != "" && !strings.HasPrefix(raw, prefix)) {
		return nil, utils.E(utils.CodeInvalidToken)
	}
	tok, err := s.emails.GetByHash(ctx, s.tokens.Hash(raw))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, utils.E(utils.CodeInvalidToken)
		}
		return nil, mapRepoErr(err)
	}
	if tok.Purpose != purpose || tok.ConsumedAt != nil || !tok.ExpiresAt.After(s.clock.Now()) {
		return nil, utils.E(utils.CodeInvalidToken)
	}
	return tok, nil
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
