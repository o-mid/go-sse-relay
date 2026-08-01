package streamtoken_test

import (
	"testing"
	"time"

	"github.com/o-mid/go-sse-relay/internal/streamtoken"
)

func TestMintAndVerify(t *testing.T) {
	tok, err := streamtoken.Mint("s3cret", "demo_user", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if err := streamtoken.Verify("s3cret", tok.Raw, "demo_user", time.Now().UTC()); err != nil {
		t.Fatalf("verify: %v", err)
	}
	if err := streamtoken.Verify("s3cret", tok.Raw, "other", time.Now().UTC()); err == nil {
		t.Fatal("expected channel mismatch")
	}
	if err := streamtoken.Verify("wrong", tok.Raw, "demo_user", time.Now().UTC()); err == nil {
		t.Fatal("expected bad signature")
	}
}

func TestVerifyExpired(t *testing.T) {
	tok, err := streamtoken.Mint("s3cret", "demo_user", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	future := time.Now().UTC().Add(2 * time.Hour)
	if err := streamtoken.Verify("s3cret", tok.Raw, "demo_user", future); err == nil {
		t.Fatal("expected expired")
	}
}

func TestVerifySkippedWithoutSecret(t *testing.T) {
	if err := streamtoken.Verify("", "", "demo_user", time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
}
