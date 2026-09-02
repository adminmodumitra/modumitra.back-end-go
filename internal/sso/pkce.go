package sso

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

// DeriveVerifier returns a PKCE S256 verifier bound to stateID so the raw
// verifier is never persisted. HMAC-SHA256(secret, stateID) → base64url.
func DeriveVerifier(secret, stateID string) string {
	if secret == "" {
		secret = "cabinet-studio-pkce-local-dev-only"
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(stateID))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func S256Challenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func HashVerifier(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return hex.EncodeToString(sum[:])
}
