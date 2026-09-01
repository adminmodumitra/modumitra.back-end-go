package org

import (
	"context"
	"log/slog"
	"strings"
	"sync"
)

type LogMailer struct {
	includeToken bool
}

func NewLogMailer(logLevel string) *LogMailer {
	return &LogMailer{includeToken: strings.EqualFold(strings.TrimSpace(logLevel), "debug")}
}

func (m *LogMailer) SendInvite(ctx context.Context, email, rawToken, orgName string) error {
	args := []any{"kind", "invite", "email", email, "org", orgName}
	if m != nil && m.includeToken {
		args = append(args, "token", rawToken)
	}
	slog.InfoContext(ctx, "mail queued", args...)
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
