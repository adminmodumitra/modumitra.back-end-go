package identity

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    uint32 = 3
	argonMemory  uint32 = 64 * 1024
	argonThreads uint8  = 4
	argonKeyLen  uint32 = 32
	argonSaltLen        = 16
)

// Argon2idHasher stores PHC strings: $argon2id$v=19$m=...,t=...,p=...$salt$hash
type Argon2idHasher struct {
	dummyHash string
}

func NewArgon2idHasher() *Argon2idHasher {
	h := &Argon2idHasher{}
	dummy, err := h.Hash(context.Background(), "dummy-timing-password")
	if err == nil {
		h.dummyHash = dummy
	}
	return h
}

func (h *Argon2idHasher) Hash(_ context.Context, password string) (string, error) {
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argonMemory, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

func (h *Argon2idHasher) Verify(_ context.Context, hash, password string) (bool, error) {
	if hash == "" || password == "" {
		return false, nil
	}
	salt, want, time, memory, threads, keyLen, err := parsePHC(hash)
	if err != nil {
		return false, err
	}
	got := argon2.IDKey([]byte(password), salt, time, memory, threads, keyLen)
	if subtle.ConstantTimeCompare(got, want) == 1 {
		return true, nil
	}
	return false, nil
}

func (h *Argon2idHasher) DummyVerify(ctx context.Context) {
	if h == nil || h.dummyHash == "" {
		_, _ = h.Hash(ctx, "dummy-timing-password")
		return
	}
	_, _ = h.Verify(ctx, h.dummyHash, "dummy-timing-password")
}

func parsePHC(phc string) (salt, key []byte, time, memory uint32, threads uint8, keyLen uint32, err error) {
	// $argon2id$v=19$m=65536,t=3,p=4$salt$hash
	parts := strings.Split(phc, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return nil, nil, 0, 0, 0, 0, fmt.Errorf("unsupported password hash")
	}
	params := strings.Split(parts[3], ",")
	var m, t uint32
	var p uint8
	for _, kv := range params {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			continue
		}
		n, convErr := strconv.ParseUint(v, 10, 32)
		if convErr != nil {
			return nil, nil, 0, 0, 0, 0, convErr
		}
		switch k {
		case "m":
			m = uint32(n)
		case "t":
			t = uint32(n)
		case "p":
			p = uint8(n)
		}
	}
	salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, 0, 0, 0, 0, err
	}
	key, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, 0, 0, 0, 0, err
	}
	if m == 0 || t == 0 || p == 0 || len(key) == 0 {
		return nil, nil, 0, 0, 0, 0, fmt.Errorf("invalid password hash")
	}
	return salt, key, t, m, p, uint32(len(key)), nil
}

// PlainHasher is a fast test double. Never use in production.
type PlainHasher struct{}

func (PlainHasher) Hash(_ context.Context, password string) (string, error) {
	return "plain:" + password, nil
}

func (PlainHasher) Verify(_ context.Context, hash, password string) (bool, error) {
	return hash == "plain:"+password, nil
}

func (PlainHasher) DummyVerify(context.Context) {}
