package identity

import (
	"context"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

type accessClaims struct {
	SessionID string `json:"sid"`
	jwt.RegisteredClaims
}

type JWTTokens struct {
	secret []byte
	issuer string
	ttl    time.Duration
	clock  Clock
}

func NewJWTTokens(secret, issuer string, ttl time.Duration, clock Clock) *JWTTokens {
	if issuer == "" {
		issuer = Issuer
	}
	if clock == nil {
		clock = SystemClock{}
	}
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	return &JWTTokens{
		secret: []byte(secret),
		issuer: issuer,
		ttl:    ttl,
		clock:  clock,
	}
}

func (t *JWTTokens) IssueAccess(_ context.Context, userID, sessionID string) (string, error) {
	now := t.clock.Now()
	claims := accessClaims{
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    t.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(t.ttl)),
			ID:        randHex(16),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := tok.SignedString(t.secret)
	if err != nil {
		return "", utils.Generic(err)
	}
	return s, nil
}

func (t *JWTTokens) ParseAccess(_ context.Context, raw string) (string, string, error) {
	unauth := func(cause error) (string, string, error) {
		return "", "", utils.E(utils.CodeUnauthenticated).With(cause)
	}
	if raw == "" {
		return unauth(nil)
	}
	parsed, err := jwt.ParseWithClaims(raw, &accessClaims{}, func(tok *jwt.Token) (any, error) {
		if tok.Method != jwt.SigningMethodHS256 {
			return nil, utils.E(utils.CodeUnauthenticated)
		}
		return t.secret, nil
	}, jwt.WithIssuer(t.issuer), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return unauth(err)
	}
	claims, ok := parsed.Claims.(*accessClaims)
	if !ok || !parsed.Valid || claims.Subject == "" || claims.SessionID == "" {
		return unauth(nil)
	}
	return claims.Subject, claims.SessionID, nil
}

func (t *JWTTokens) NewRefresh() (string, string) {
	raw := PrefixRefresh + randHex(32)
	return raw, HashOpaque(raw)
}

func (t *JWTTokens) NewEmailToken(prefix string) (string, string) {
	raw := prefix + randHex(32)
	return raw, HashOpaque(raw)
}

func (t *JWTTokens) Hash(raw string) string { return HashOpaque(raw) }

func (t *JWTTokens) HashOTP(userID, otp string) string { return HashOTP(userID, otp) }

func HashOpaque(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// HashOTP binds a 6-digit code to a user so the unique token_hash constraint
// is not global across accounts.
func HashOTP(userID, otp string) string {
	sum := sha256.Sum256([]byte(PurposeOTPLogin + ":" + userID + ":" + otp))
	return hex.EncodeToString(sum[:])
}

// NewOTP returns a cryptographically random 6-digit code, left-padded.
func NewOTP() string {
	n, err := crand.Int(crand.Reader, big.NewInt(1_000_000))
	if err != nil {
		panic("crypto/rand: " + err.Error())
	}
	return fmt.Sprintf("%06d", n.Int64())
}
