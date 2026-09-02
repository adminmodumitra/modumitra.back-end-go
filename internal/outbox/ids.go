package outbox

import (
	"crypto/rand"
	"encoding/hex"
)

const PrefixOutbox = "obx_"

func newPublicID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand: " + err.Error())
	}
	return PrefixOutbox + hex.EncodeToString(b)
}
