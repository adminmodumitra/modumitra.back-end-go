package identity

import (
	"context"
	"strings"
	"sync"

	"github.com/Danish-13-Khan/backend-cabinet/internal/obs"
)

// LogMailer logs that a mail was queued. The raw token is included only at debug.
type LogMailer struct {
	includeToken bool
}

func NewLogMailer(logLevel string) *LogMailer {
	return &LogMailer{includeToken: strings.EqualFold(strings.TrimSpace(logLevel), "debug")}
}

func (m *LogMailer) SendVerification(ctx context.Context, email, rawToken string) error {
	return m.log(ctx, "verification", email, rawToken)
}

func (m *LogMailer) SendPasswordReset(ctx context.Context, email, rawToken string) error {
	return m.log(ctx, "password_reset", email, rawToken)
}

func (m *LogMailer) SendOTP(ctx context.Context, email, otp string) error {
	l := obs.FromCtx(ctx)
	ev := l.Info().Str("kind", "otp").Str("email", email)
	if m != nil && m.includeToken {
		ev = ev.Str("otp", otp)
	}
	ev.Msg("mail queued")
	return nil
}

func (m *LogMailer) log(ctx context.Context, kind, email, rawToken string) error {
	l := obs.FromCtx(ctx)
	ev := l.Info().Str("kind", kind).Str("email", email)
	if m != nil && m.includeToken {
		ev = ev.Str("token", rawToken)
	}
	ev.Msg("mail queued")
	return nil
}

type MailCall struct {
	Email string
	Token string
}

// RecordingMailer is a test double that stores sent tokens.
type RecordingMailer struct {
	mu             sync.Mutex
	Verifications  []MailCall
	PasswordResets []MailCall
	OTPs           []MailCall
}

func (m *RecordingMailer) SendVerification(_ context.Context, email, rawToken string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Verifications = append(m.Verifications, MailCall{Email: email, Token: rawToken})
	return nil
}

func (m *RecordingMailer) SendPasswordReset(_ context.Context, email, rawToken string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PasswordResets = append(m.PasswordResets, MailCall{Email: email, Token: rawToken})
	return nil
}

func (m *RecordingMailer) SendOTP(_ context.Context, email, otp string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.OTPs = append(m.OTPs, MailCall{Email: email, Token: otp})
	return nil
}

func (m *RecordingMailer) LastVerification() (MailCall, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.Verifications) == 0 {
		return MailCall{}, false
	}
	return m.Verifications[len(m.Verifications)-1], true
}

func (m *RecordingMailer) LastPasswordReset() (MailCall, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.PasswordResets) == 0 {
		return MailCall{}, false
	}
	return m.PasswordResets[len(m.PasswordResets)-1], true
}

func (m *RecordingMailer) LastOTP() (MailCall, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.OTPs) == 0 {
		return MailCall{}, false
	}
	return m.OTPs[len(m.OTPs)-1], true
}
