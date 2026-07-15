package auth_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/o-mid/go-sse-relay/internal/auth"
)

func TestValidHMACSHA256(t *testing.T) {
	body := []byte(`{"ok":true}`)
	mac := hmac.New(sha256.New, []byte("s3cret"))
	_, _ = mac.Write(body)
	sig := hex.EncodeToString(mac.Sum(nil))

	if !auth.ValidHMACSHA256("s3cret", body, sig) {
		t.Fatal("expected valid signature")
	}
	if !auth.ValidHMACSHA256("s3cret", body, "sha256="+sig) {
		t.Fatal("expected valid prefixed signature")
	}
	if auth.ValidHMACSHA256("s3cret", body, "deadbeef") {
		t.Fatal("expected invalid signature")
	}
	if !auth.ValidHMACSHA256("", body, "") {
		t.Fatal("empty secret should skip verification")
	}
}
