package device

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"time"
)

// HMACSigner signs canonical snapshot bytes with HMAC-SHA256.
type HMACSigner struct {
	key   []byte
	keyID string
}

func NewHMACSigner(secret, keyID string) HMACSigner {
	if secret == "" {
		secret = DefaultSigningKey
	}
	if keyID == "" {
		keyID = KeyIDV1
	}
	return HMACSigner{key: []byte(secret), keyID: keyID}
}

func (s HMACSigner) KeyID() string { return s.keyID }

func (s HMACSigner) mac(canonical []byte) []byte {
	h := hmac.New(sha256.New, s.key)
	h.Write(canonical)
	return h.Sum(nil)
}

func (s HMACSigner) Sign(canonical []byte) string {
	return base64.StdEncoding.EncodeToString(s.mac(canonical))
}

func (s HMACSigner) Verify(canonical []byte, signature string) bool {
	got, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false
	}
	return hmac.Equal(got, s.mac(canonical))
}

type unsignedSnapshot struct {
	UserID             string       `json:"user_id"`
	OrganizationID     *string      `json:"organization_id"`
	Plan               string       `json:"plan"`
	SubscriptionStatus string       `json:"subscription_status"`
	Capabilities       Capabilities `json:"capabilities"`
	IssuedAt           string       `json:"issued_at"`
	RefreshAfter       string       `json:"refresh_after"`
	ExpiresAt          string       `json:"expires_at"`
	GraceUntil         *string      `json:"grace_until"`
	DeviceID           *string      `json:"device_id"`
	SchemaVersion      int          `json:"schema_version"`
}

func rfc3339(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func rfc3339Ptr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := rfc3339(*t)
	return &s
}

// CanonicalSnapshot is compact JSON of the snapshot excluding signature,
// with stable struct key order and RFC3339 UTC timestamps.
func CanonicalSnapshot(s Snapshot) ([]byte, error) {
	payload := unsignedSnapshot{
		UserID:             s.UserID,
		OrganizationID:     s.OrganizationID,
		Plan:               s.Plan,
		SubscriptionStatus: s.SubscriptionStatus,
		Capabilities:       s.Capabilities,
		IssuedAt:           rfc3339(s.IssuedAt),
		RefreshAfter:       rfc3339(s.RefreshAfter),
		ExpiresAt:          rfc3339(s.ExpiresAt),
		GraceUntil:         rfc3339Ptr(s.GraceUntil),
		DeviceID:           s.DeviceID,
		SchemaVersion:      s.SchemaVersion,
	}
	return json.Marshal(payload)
}

func (s HMACSigner) SignSnapshot(snap *Snapshot) error {
	canon, err := CanonicalSnapshot(*snap)
	if err != nil {
		return err
	}
	snap.Signature = s.Sign(canon)
	return nil
}

func (s HMACSigner) VerifySnapshot(snap Snapshot) bool {
	canon, err := CanonicalSnapshot(snap)
	if err != nil {
		return false
	}
	return s.Verify(canon, snap.Signature)
}

var _ Signer = HMACSigner{}
