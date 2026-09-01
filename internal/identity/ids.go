package identity

import (
	"crypto/rand"
	"encoding/hex"
)

type RandomIDs struct{}

func NewIDGen() RandomIDs { return RandomIDs{} }

func (RandomIDs) User() string    { return PrefixUser + randHex(16) }
func (RandomIDs) Session() string { return PrefixSession + randHex(16) }
func (RandomIDs) Token() string   { return PrefixToken + randHex(16) }
func (RandomIDs) Family() string  { return PrefixFamily + randHex(16) }

func randHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand failure means the host is unusable.
		panic("crypto/rand: " + err.Error())
	}
	return hex.EncodeToString(b)
}
