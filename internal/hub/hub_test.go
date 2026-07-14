package hub_test

import (
	"testing"
	"time"

	"github.com/o-mid/go-sse-relay/internal/hub"
)

func TestPublishFansOutToSubscribers(t *testing.T) {
	h := hub.New()
	a := h.Subscribe("user_1", 4)
	b := h.Subscribe("user_1", 4)
	defer h.Unsubscribe(a)
	defer h.Unsubscribe(b)

	published, n := h.Publish(hub.Event{
		Channel: "user_1",
		Type:    "payment.updated",
		Data:    []byte(`{"status":"settled"}`),
	})
	if n != 2 {
		t.Fatalf("delivered=%d want 2", n)
	}
	if published.ID == "" {
		t.Fatal("expected event id")
	}

	for _, sub := range []*hub.Subscriber{a, b} {
		select {
		case ev := <-sub.C:
			if ev.Channel != "user_1" {
				t.Fatalf("channel=%q", ev.Channel)
			}
			if ev.Type != "payment.updated" {
				t.Fatalf("type=%q", ev.Type)
			}
			if ev.ID != published.ID {
				t.Fatalf("id=%q want %q", ev.ID, published.ID)
			}
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for event")
		}
	}
}

func TestPublishIgnoresOtherChannels(t *testing.T) {
	h := hub.New()
	sub := h.Subscribe("user_a", 1)
	defer h.Unsubscribe(sub)

	_, n := h.Publish(hub.Event{Channel: "user_b", Type: "x", Data: []byte(`{}`)})
	if n != 0 {
		t.Fatalf("delivered=%d want 0", n)
	}
	if h.SubscriberCount("user_a") != 1 {
		t.Fatalf("subscriber count=%d", h.SubscriberCount("user_a"))
	}
}

func TestSlowSubscriberDoesNotBlock(t *testing.T) {
	h := hub.New()
	sub := h.Subscribe("ch", 1)
	defer h.Unsubscribe(sub)

	// Fill the buffer.
	_, _ = h.Publish(hub.Event{Channel: "ch", Type: "one", Data: []byte(`1`)})
	// Second publish should drop for this sub, not block.
	done := make(chan int, 1)
	go func() {
		_, n := h.Publish(hub.Event{Channel: "ch", Type: "two", Data: []byte(`2`)})
		done <- n
	}()
	select {
	case n := <-done:
		if n != 0 {
			t.Fatalf("delivered=%d want 0 for full buffer", n)
		}
	case <-time.After(time.Second):
		t.Fatal("publish blocked on slow subscriber")
	}
}
