package billing

import (
	"crypto/rand"
	"encoding/hex"
)

type RandomIDs struct{}

func NewIDGen() RandomIDs { return RandomIDs{} }

func (RandomIDs) Account() string      { return PrefixAccount + randHex(16) }
func (RandomIDs) Subscription() string { return PrefixSub + randHex(16) }
func (RandomIDs) Invoice() string      { return PrefixInvoice + randHex(16) }
func (RandomIDs) Event() string        { return PrefixEvent + randHex(16) }

func randHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand: " + err.Error())
	}
	return hex.EncodeToString(b)
}
