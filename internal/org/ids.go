package org

import (
	"crypto/rand"
	"encoding/hex"
)

type RandomIDs struct{}

func NewIDGen() RandomIDs { return RandomIDs{} }

func (RandomIDs) Org() string        { return PrefixOrg + randHex(16) }
func (RandomIDs) Member() string     { return PrefixMember + randHex(16) }
func (RandomIDs) Invitation() string { return PrefixInvitation + randHex(16) }

func randHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand: " + err.Error())
	}
	return hex.EncodeToString(b)
}
