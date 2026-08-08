// Package hub fans published events out to per-channel subscribers.
package hub

import (
	"sync"
	"sync/atomic"
)

// Event is one fan-out message.
type Event struct {
	ID      string
	Channel string
	Type    string
	Data    []byte
	// Origin is the publishing instance id (used to ignore bus echoes).
	Origin string
}

// Subscriber receives events for a single channel.
type Subscriber struct {
	Channel string
	C       chan Event
}

// Hub keeps channel subscribers and publishes events to them.
type Hub struct {
	mu   sync.RWMutex
	subs map[string]map[*Subscriber]struct{}
	seq  atomic.Uint64
}

// New creates an empty hub.
func New() *Hub {
	return &Hub{
		subs: make(map[string]map[*Subscriber]struct{}),
	}
}

// Subscribe registers a buffered subscriber on channel.
func (h *Hub) Subscribe(channel string, buffer int) *Subscriber {
	if buffer < 1 {
		buffer = 1
	}
	sub := &Subscriber{
		Channel: channel,
		C:       make(chan Event, buffer),
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.subs[channel] == nil {
		h.subs[channel] = make(map[*Subscriber]struct{})
	}
	h.subs[channel][sub] = struct{}{}
	return sub
}

// Unsubscribe removes a subscriber and closes its channel.
func (h *Hub) Unsubscribe(sub *Subscriber) {
	if sub == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	set := h.subs[sub.Channel]
	if set == nil {
		return
	}
	if _, ok := set[sub]; !ok {
		return
	}
	delete(set, sub)
	if len(set) == 0 {
		delete(h.subs, sub.Channel)
	}
	close(sub.C)
}

// Publish delivers an event to all subscribers on its channel.
// Slow subscribers that fill their buffer drop the event (non-blocking send).
// The returned event includes the assigned ID.
func (h *Hub) Publish(ev Event) (Event, int) {
	if ev.ID == "" {
		ev.ID = h.nextID()
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	set := h.subs[ev.Channel]
	delivered := 0
	for sub := range set {
		select {
		case sub.C <- ev:
			delivered++
		default:
			// Drop when the subscriber is not keeping up.
		}
	}
	return ev, delivered
}

// SubscriberCount returns live subscribers for a channel.
func (h *Hub) SubscriberCount(channel string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.subs[channel])
}

func (h *Hub) nextID() string {
	n := h.seq.Add(1)
	// Decimal string keeps Last-Event-ID simple for clients.
	buf := make([]byte, 0, 20)
	return string(appendUint(buf, n))
}

func appendUint(dst []byte, n uint64) []byte {
	if n == 0 {
		return append(dst, '0')
	}
	var tmp [20]byte
	i := len(tmp)
	for n > 0 {
		i--
		tmp[i] = byte('0' + n%10)
		n /= 10
	}
	return append(dst, tmp[i:]...)
}
