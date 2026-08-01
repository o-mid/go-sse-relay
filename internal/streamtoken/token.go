// Package streamtoken mints and verifies signed channel access tokens.
package streamtoken

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Token is a signed capability for one channel until Exp.
type Token struct {
	Channel string
	Exp     time.Time
	Raw     string
}

// Mint returns v1.<channel>.<exp_unix>.<hex_hmac>.
func Mint(secret, channel string, ttl time.Duration) (Token, error) {
	if secret == "" {
		return Token{}, fmt.Errorf("secret required")
	}
	if channel == "" {
		return Token{}, fmt.Errorf("channel required")
	}
	if ttl <= 0 {
		ttl = time.Hour
	}
	exp := time.Now().UTC().Add(ttl).Unix()
	sig := sign(secret, channel, exp)
	raw := fmt.Sprintf("v1.%s.%d.%s", channel, exp, sig)
	return Token{
		Channel: channel,
		Exp:     time.Unix(exp, 0).UTC(),
		Raw:     raw,
	}, nil
}

// Verify checks signature, expiry, and channel match.
// wantChannel must match the token channel.
func Verify(secret, raw, wantChannel string, now time.Time) error {
	if secret == "" {
		return nil
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("token required")
	}
	parts := strings.Split(raw, ".")
	if len(parts) != 4 || parts[0] != "v1" {
		return fmt.Errorf("invalid token format")
	}
	channel := parts[1]
	expUnix, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid token expiry")
	}
	sig := parts[3]
	expected := sign(secret, channel, expUnix)
	if subtle.ConstantTimeCompare([]byte(strings.ToLower(expected)), []byte(strings.ToLower(sig))) != 1 {
		return fmt.Errorf("invalid token signature")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if now.Unix() > expUnix {
		return fmt.Errorf("token expired")
	}
	if wantChannel != "" && channel != wantChannel {
		return fmt.Errorf("token channel mismatch")
	}
	return nil
}

func sign(secret, channel string, exp int64) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = fmt.Fprintf(mac, "v1|%s|%d", channel, exp)
	return hex.EncodeToString(mac.Sum(nil))
}
