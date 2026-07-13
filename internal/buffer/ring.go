// Package buffer keeps a short per-channel event history for SSE reconnect.
package buffer

import (
	"sync"

	"github.com/o-mid/go-sse-relay/internal/hub"
)

// Ring stores the last N events per channel.
type Ring struct {
	mu       sync.RWMutex
	capacity int
	byChan   map[string]*channelBuf
}

type channelBuf struct {
	events []hub.Event
}

// New creates a ring with the given per-channel capacity.
func New(capacity int) *Ring {
	if capacity < 1 {
		capacity = 1
	}
	return &Ring{
		capacity: capacity,
		byChan:   make(map[string]*channelBuf),
	}
}

// Add appends an event and trims to capacity.
func (r *Ring) Add(ev hub.Event) {
	r.mu.Lock()
	defer r.mu.Unlock()
	buf := r.byChan[ev.Channel]
	if buf == nil {
		buf = &channelBuf{}
		r.byChan[ev.Channel] = buf
	}
	buf.events = append(buf.events, ev)
	if len(buf.events) > r.capacity {
		buf.events = append([]hub.Event(nil), buf.events[len(buf.events)-r.capacity:]...)
	}
}

// Since returns events after lastID (exclusive). If lastID is empty or unknown,
// it returns the full buffered history for the channel.
func (r *Ring) Since(channel, lastID string) []hub.Event {
	r.mu.RLock()
	defer r.mu.RUnlock()
	buf := r.byChan[channel]
	if buf == nil || len(buf.events) == 0 {
		return nil
	}
	if lastID == "" {
		out := make([]hub.Event, len(buf.events))
		copy(out, buf.events)
		return out
	}
	idx := -1
	for i, ev := range buf.events {
		if ev.ID == lastID {
			idx = i
			break
		}
	}
	if idx < 0 {
		out := make([]hub.Event, len(buf.events))
		copy(out, buf.events)
		return out
	}
	rest := buf.events[idx+1:]
	out := make([]hub.Event, len(rest))
	copy(out, rest)
	return out
}
