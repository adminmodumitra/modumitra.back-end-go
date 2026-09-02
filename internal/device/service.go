package device

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/Danish-13-Khan/backend-cabinet/internal/outbox"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

type Service struct {
	devices DeviceRepo
	ents    EntitlementRepo
	tx      TxRunner
	orgs    OrgLookup
	signer  Signer
	clock   Clock
	ids     IDGen
	outbox  outbox.Writer
}

func NewService(
	devices DeviceRepo,
	ents EntitlementRepo,
	tx TxRunner,
	orgs OrgLookup,
	signer Signer,
	clock Clock,
	ids IDGen,
	ob outbox.Writer,
) *Service {
	if clock == nil {
		clock = SystemClock{}
	}
	if ids == nil {
		ids = NewIDGen()
	}
	if signer == nil {
		s := NewHMACSigner(DefaultSigningKey, KeyIDV1)
		signer = s
	}
	return &Service{
		devices: devices,
		ents:    ents,
		tx:      tx,
		orgs:    orgs,
		signer:  signer,
		clock:   clock,
		ids:     ids,
		outbox:  ob,
	}
}

func NewMemoryService(orgs OrgLookup, signingKey string) (*Service, *Memory) {
	mem := NewMemory()
	svc := NewService(mem.Devices(), mem.Entitlements(), mem, orgs, NewHMACSigner(signingKey, KeyIDV1), SystemClock{}, NewIDGen(), outbox.NewMemoryWriter())
	return svc, mem
}

func validName(s string) bool {
	n := utf8.RuneCountInString(strings.TrimSpace(s))
	return n >= 1 && n <= 80
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

func (s *Service) ownedDevice(ctx context.Context, userID, deviceID string) (*Device, error) {
	if userID == "" {
		return nil, utils.E(utils.CodeUnauthenticated)
	}
	if deviceID == "" {
		return nil, utils.E(utils.CodeNotFound)
	}
	d, err := s.devices.GetByID(ctx, deviceID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if d.UserID != userID {
		return nil, utils.E(utils.CodeNotFound)
	}
	return d, nil
}
