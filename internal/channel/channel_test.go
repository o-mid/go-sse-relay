package channel_test

import (
	"strings"
	"testing"

	"github.com/o-mid/go-sse-relay/internal/channel"
)

func TestValid(t *testing.T) {
	cases := map[string]bool{
		"demo_user":   true,
		"user-1":      true,
		"tenant:42":   true,
		"a.b":         true,
		"":            false,
		"bad channel": false,
		"../x":        false,
		"a/b":         false,
		strings.Repeat("x", 129): false,
		strings.Repeat("x", 128): true,
	}
	for name, want := range cases {
		if got := channel.Valid(name); got != want {
			t.Fatalf("Valid(%q)=%v want %v", name, got, want)
		}
	}
}
