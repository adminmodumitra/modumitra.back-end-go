package org

import (
	"crypto/sha256"
	"encoding/hex"
)

type SHATokens struct{}

func NewTokens() SHATokens { return SHATokens{} }

func (SHATokens) NewInvite() (string, string) {
	raw := PrefixInviteTok + randHex(32)
	return raw, HashOpaque(raw)
}

func (SHATokens) Hash(raw string) string { return HashOpaque(raw) }

func HashOpaque(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
