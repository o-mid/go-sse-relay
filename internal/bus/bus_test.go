package bus_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/o-mid/go-sse-relay/internal/bus"
	"github.com/o-mid/go-sse-relay/internal/hub"
)

func TestEncodeDecode(t *testing.T) {
	in := hub.Event{ID: "9", Channel: "c", Type: "t", Data: []byte(`{"a":1}`), Origin: "node-a"}
	b, err := bus.Encode(in)
	if err != nil {
		t.Fatal(err)
	}
	out, err := bus.Decode(b)
	if err != nil {
		t.Fatal(err)
	}
	if out.ID != in.ID || out.Channel != in.Channel || out.Type != in.Type || out.Origin != in.Origin {
		t.Fatalf("out=%+v", out)
	}
	if string(out.Data) != string(in.Data) {
		t.Fatalf("data=%s", out.Data)
	}
}

func TestMemoryNoop(t *testing.T) {
	m := bus.NewMemory()
	if err := m.Publish(context.Background(), hub.Event{Channel: "c", Type: "t"}); err != nil {
		t.Fatal(err)
	}
	if err := m.Start(context.Background(), func(hub.Event) {}); err != nil {
		t.Fatal(err)
	}
	_ = m.Close()
}

func TestRedisFanout(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	a, err := bus.NewRedis("redis://" + mr.Addr())
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	b, err := bus.NewRedis("redis://" + mr.Addr())
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()

	got := make(chan hub.Event, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := b.Start(ctx, func(ev hub.Event) { got <- ev }); err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond)

	if err := a.Publish(ctx, hub.Event{
		ID: "1", Channel: "user_1", Type: "payment.updated",
		Data: []byte(`{"ok":true}`), Origin: "a",
	}); err != nil {
		t.Fatal(err)
	}

	select {
	case ev := <-got:
		if ev.Channel != "user_1" || ev.Origin != "a" {
			t.Fatalf("ev=%+v", ev)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for redis fan-out")
	}
}
