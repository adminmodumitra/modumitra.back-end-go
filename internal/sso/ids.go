package sso

import (
	"crypto/rand"
	"encoding/hex"
)

type IDGen interface {
	Identity() string
	State() string
	Device() string
	UserCode() string
}

type RandomIDs struct{}

func NewIDGen() RandomIDs { return RandomIDs{} }

func (RandomIDs) Identity() string { return PrefixIdentity + randHex(16) }
func (RandomIDs) State() string    { return PrefixState + randHex(16) }
func (RandomIDs) Device() string   { return PrefixDevice + randHex(16) }

func (RandomIDs) UserCode() string {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 8)
	raw := make([]byte, 8)
	if _, err := rand.Read(raw); err != nil {
		panic("crypto/rand: " + err.Error())
	}
	for i := 0; i < 8; i++ {
		b[i] = alphabet[int(raw[i])%len(alphabet)]
	}
	return string(b[0:4]) + "-" + string(b[4:8])
}

func randHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand: " + err.Error())
	}
	return hex.EncodeToString(b)
}
