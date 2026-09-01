package device

import (
	"crypto/rand"
	"encoding/hex"
)

type RandomIDs struct{}

func NewIDGen() RandomIDs { return RandomIDs{} }

func (RandomIDs) Device() string      { return PrefixDevice + randHex(16) }
func (RandomIDs) Entitlement() string { return PrefixEntitlement + randHex(16) }

func randHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand: " + err.Error())
	}
	return hex.EncodeToString(b)
}
