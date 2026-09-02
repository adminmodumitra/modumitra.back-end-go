package org

import (
	"errors"
	"net/mail"
	"strings"
	"unicode/utf8"

	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/outbox"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

type Service struct {
	orgs      OrgRepo
	members   MemberRepo
	invites   InvitationRepo
	users     identity.UserRepo
	tx        TxRunner
	mailer    Mailer
	clock     Clock
	ids       IDGen
	tokens    Tokens
	trialDays int
	outbox    outbox.Writer
}

func NewService(
	orgs OrgRepo,
	members MemberRepo,
	invites InvitationRepo,
	users identity.UserRepo,
	tx TxRunner,
	mailer Mailer,
	clock Clock,
	ids IDGen,
	tokens Tokens,
	cfg Settings,
	ob outbox.Writer,
) *Service {
	if clock == nil {
		clock = SystemClock{}
	}
	if ids == nil {
		ids = NewIDGen()
	}
	if tokens == nil {
		tokens = NewTokens()
	}
	if cfg.TrialDays <= 0 {
		cfg.TrialDays = 14
	}
	return &Service{
		orgs:      orgs,
		members:   members,
		invites:   invites,
		users:     users,
		tx:        tx,
		mailer:    mailer,
		clock:     clock,
		ids:       ids,
		tokens:    tokens,
		trialDays: cfg.TrialDays,
		outbox:    ob,
	}
}

func (s *Service) Outbox() outbox.Writer { return s.outbox }

func NewMemoryService(users identity.UserRepo, trialDays int) (*Service, *RecordingMailer, *Memory) {
	mem := NewMemory()
	mail := &RecordingMailer{}
	svc := NewService(mem.Orgs(), mem.Members(), mem.Invites(), users, mem, mail, SystemClock{}, NewIDGen(), NewTokens(), Settings{TrialDays: trialDays}, outbox.NewMemoryWriter())
	return svc, mail, mem
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
	return err == nil && addr.Address != ""
}

func validOrgName(s string) bool {
	n := utf8.RuneCountInString(strings.TrimSpace(s))
	return n >= 1 && n <= 120
}

func mapRepoErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrNotFound) || errors.Is(err, identity.ErrNotFound) {
		return utils.E(utils.CodeNotFound)
	}
	return utils.From(err)
}
