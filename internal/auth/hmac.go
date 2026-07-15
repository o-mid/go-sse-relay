// Package auth verifies optional shared-secret signatures on ingest.
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"strings"
)

// ValidHMACSHA256 reports whether signature matches hex(hmac_sha256(secret, body)).
// Empty secret disables verification (local demos).
func ValidHMACSHA256(secret string, body []byte, signature string) bool {
	if secret == "" {
		return true
	}
	sig := strings.TrimSpace(signature)
	sig = strings.TrimPrefix(sig, "sha256=")
	if sig == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return subtle.ConstantTimeCompare([]byte(strings.ToLower(expected)), []byte(strings.ToLower(sig))) == 1
}
