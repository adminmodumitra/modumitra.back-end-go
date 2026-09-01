package billing

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// Webhook HMAC failures are HTTP 401 invalid_signature (API_CONTRACTS_V1.md route 63).
// X-Billing-Signature: hex HMAC-SHA256 of the raw body.
// Stripe-Signature: t=<unix>,v1=<hex> over "{t}.{body}", or v1 hex of the raw body (fake adapter).

func SignBody(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func PayloadDigest(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func VerifyWebhookSignature(secret string, body []byte, stripeSig, billingSig string) error {
	if strings.TrimSpace(secret) == "" {
		return utils.E(utils.CodeInvalidSignature)
	}
	billingSig = strings.TrimSpace(billingSig)
	stripeSig = strings.TrimSpace(stripeSig)
	if billingSig != "" {
		if hexMACEqual(secret, body, billingSig) {
			return nil
		}
		return utils.E(utils.CodeInvalidSignature)
	}
	if stripeSig != "" {
		if verifyStripeSignature(secret, body, stripeSig) {
			return nil
		}
		return utils.E(utils.CodeInvalidSignature)
	}
	return utils.E(utils.CodeInvalidSignature)
}

func verifyStripeSignature(secret string, body []byte, header string) bool {
	t, v1 := parseStripeHeader(header)
	if v1 == "" {
		return hexMACEqual(secret, body, header)
	}
	if t != "" {
		signed := append([]byte(t+"."), body...)
		if hexMACEqual(secret, signed, v1) {
			return true
		}
	}
	return hexMACEqual(secret, body, v1)
}

func parseStripeHeader(header string) (t, v1 string) {
	for _, part := range strings.Split(header, ",") {
		k, val, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			continue
		}
		switch k {
		case "t":
			t = val
		case "v1":
			v1 = val
		}
	}
	return t, v1
}

func hexMACEqual(secret string, payload []byte, sigHex string) bool {
	want, err := hex.DecodeString(SignBody(secret, payload))
	if err != nil {
		return false
	}
	got, err := hex.DecodeString(strings.TrimSpace(sigHex))
	if err != nil || len(got) != len(want) {
		return false
	}
	return hmac.Equal(want, got)
}
