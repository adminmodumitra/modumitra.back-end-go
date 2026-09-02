package org

import (
	"context"
	"strings"
	"sync"

	"github.com/Danish-13-Khan/backend-cabinet/internal/obs"
)

type LogMailer struct {
	includeToken bool
}

func NewLogMailer(logLevel string) *LogMailer {
	return &LogMailer{includeToken: strings.EqualFold(strings.TrimSpace(logLevel), "debug")}
}

func (m *LogMailer) SendInvite(ctx context.Context, email, rawToken, orgName string) error {
	l := obs.FromCtx(ctx)
	ev := l.Info().Str("kind", "invite").Str("email", email).Str("org", orgName)
	if m != nil && m.includeToken {
		ev = ev.Str("token", rawToken)
	}
	ev.Msg("mail queued")
	return nil
}

type InviteCall struct {
	Email string
	Token string
	Org   string
}

type RecordingMailer struct {
	mu      sync.Mutex
	Invites []InviteCall
}

func (m *RecordingMailer) SendInvite(_ context.Context, email, rawToken, orgName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Invites = append(m.Invites, InviteCall{Email: email, Token: rawToken, Org: orgName})
	return nil
}

func (m *RecordingMailer) LastInvite() (InviteCall, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.Invites) == 0 {
		return InviteCall{}, false
	}
	return m.Invites[len(m.Invites)-1], true
}
