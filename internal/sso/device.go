package sso

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/obs"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (s *Service) CreateDeviceCode(ctx context.Context, clientID, provider string) (*DeviceCodeResult, error) {
	if strings.TrimSpace(clientID) == "" {
		return nil, utils.E(utils.CodeValidation).WithDetails(map[string]any{"field": "client_id"})
	}
	if clientID != s.cfg.DeviceClientID {
		return nil, utils.E(utils.CodeInvalidClient)
	}
	p, err := s.provider(provider)
	if err != nil {
		return nil, err
	}
	if !p.DeviceCodeSupported() {
		return nil, utils.E(utils.CodeUnsupportedProvider)
	}
	now := s.clock.Now()
	id := s.ids.Device()
	userCode := s.ids.UserCode()
	d := &DeviceCode{
		ID:              id,
		DeviceCode:      id,
		UserCode:        userCode,
		ClientID:        clientID,
		Provider:        provider,
		Status:          StatusPending,
		IntervalSeconds: s.cfg.DeviceInterval,
		ExpiresAt:       now.Add(s.cfg.DeviceTTL),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.devices.Insert(ctx, d); err != nil {
		return nil, mapRepoErr(err)
	}
	complete := s.cfg.VerificationURI
	if strings.Contains(complete, "?") {
		complete += "&user_code=" + url.QueryEscape(userCode)
	} else {
		complete += "?user_code=" + url.QueryEscape(userCode)
	}
	return &DeviceCodeResult{
		DeviceCode:              d.DeviceCode,
		UserCode:                d.UserCode,
		VerificationURI:         s.cfg.VerificationURI,
		VerificationURIComplete: complete,
		ExpiresIn:               int(s.cfg.DeviceTTL.Seconds()),
		Interval:                d.IntervalSeconds,
	}, nil
}

func (s *Service) PollDeviceToken(ctx context.Context, deviceCode string) (*identity.User, *identity.TokenPair, error) {
	if strings.TrimSpace(deviceCode) == "" {
		return nil, nil, utils.E(utils.CodeValidation)
	}
	d, err := s.devices.GetByDeviceCode(ctx, deviceCode)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, nil, utils.E(utils.CodeInvalidGrant)
		}
		return nil, nil, mapRepoErr(err)
	}
	now := s.clock.Now()
	if d.Status == StatusConsumed {
		return nil, nil, utils.E(utils.CodeInvalidGrant)
	}
	if d.Status == StatusDenied {
		return nil, nil, utils.E(utils.CodeAccessDenied)
	}
	if d.Status == StatusExpired || !d.ExpiresAt.After(now) {
		if d.Status != StatusExpired {
			d.Status = StatusExpired
			d.UpdatedAt = now
			_ = s.devices.Update(ctx, d)
		}
		return nil, nil, utils.E(utils.CodeExpiredToken)
	}
	if d.Status == StatusPending {
		if d.LastPollAt != nil && now.Sub(*d.LastPollAt) < time.Duration(d.IntervalSeconds)*time.Second {
			return nil, nil, utils.E(utils.CodeSlowDown).WithDetails(map[string]any{"interval": d.IntervalSeconds})
		}
		t := now
		d.LastPollAt = &t
		d.UpdatedAt = now
		_ = s.devices.Update(ctx, d)
		return nil, nil, utils.E(utils.CodeAuthorizationPending).WithDetails(map[string]any{"interval": d.IntervalSeconds})
	}
	if d.Status != StatusAuthorized || d.UserID == "" {
		return nil, nil, utils.E(utils.CodeAuthorizationPending).WithDetails(map[string]any{"interval": d.IntervalSeconds})
	}
	user, pair, err := s.accounts.LoginAs(ctx, d.UserID, "")
	if err != nil {
		return nil, nil, err
	}
	cons := now
	d.Status = StatusConsumed
	d.ConsumedAt = &cons
	d.UpdatedAt = now
	if err := s.devices.Update(ctx, d); err != nil {
		return nil, nil, mapRepoErr(err)
	}
	l := obs.FromCtx(ctx)
	l.Info().Str("user_id", user.ID).Msg("auth.device_login")
	return user, pair, nil
}

func (s *Service) AuthorizeDevice(ctx context.Context, userCode, userID string) error {
	if strings.TrimSpace(userCode) == "" || strings.TrimSpace(userID) == "" {
		return utils.E(utils.CodeValidation)
	}
	if _, err := s.accounts.UserByID(ctx, userID); err != nil {
		return mapRepoErr(err)
	}
	d, err := s.devices.GetByUserCode(ctx, userCode)
	if err != nil {
		return mapRepoErr(err)
	}
	now := s.clock.Now()
	if d.Status != StatusPending || !d.ExpiresAt.After(now) {
		return utils.E(utils.CodeInvalidGrant)
	}
	t := now
	d.Status = StatusAuthorized
	d.UserID = userID
	d.AuthorizedAt = &t
	d.UpdatedAt = now
	return mapRepoErr(s.devices.Update(ctx, d))
}
