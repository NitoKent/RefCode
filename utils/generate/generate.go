package generate

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateReferralCode() string {
	b := make([]byte, 8)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
