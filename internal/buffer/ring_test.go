package buffer_test

import (
	"testing"

	"github.com/o-mid/go-sse-relay/internal/buffer"
	"github.com/o-mid/go-sse-relay/internal/hub"
)

func TestRingTrimsToCapacity(t *testing.T) {
	r := buffer.New(2)
	r.Add(hub.Event{ID: "1", Channel: "c"})
	r.Add(hub.Event{ID: "2", Channel: "c"})
	r.Add(hub.Event{ID: "3", Channel: "c"})
	got := r.Since("c", "")
	if len(got) != 2 {
		t.Fatalf("len=%d want 2", len(got))
	}
	if got[0].ID != "2" || got[1].ID != "3" {
		t.Fatalf("ids=%v", []string{got[0].ID, got[1].ID})
	}
}

func TestRingSinceExclusive(t *testing.T) {
	r := buffer.New(10)
	r.Add(hub.Event{ID: "1", Channel: "c"})
	r.Add(hub.Event{ID: "2", Channel: "c"})
	r.Add(hub.Event{ID: "3", Channel: "c"})
	got := r.Since("c", "1")
	if len(got) != 2 || got[0].ID != "2" || got[1].ID != "3" {
		t.Fatalf("got=%v", got)
	}
}

func TestRingUnknownLastIDReplaysAll(t *testing.T) {
	r := buffer.New(10)
	r.Add(hub.Event{ID: "1", Channel: "c"})
	got := r.Since("c", "missing")
	if len(got) != 1 || got[0].ID != "1" {
		t.Fatalf("got=%v", got)
	}
}
