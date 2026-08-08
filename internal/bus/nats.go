package bus

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/o-mid/go-sse-relay/internal/hub"
)

// NATS fans events over a NATS subject.
type NATS struct {
	nc   *nats.Conn
	sub  *nats.Subscription
	url  string
}

// NewNATS connects to a NATS server URL.
func NewNATS(url string) (*NATS, error) {
	if url == "" {
		url = nats.DefaultURL
	}
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}
	return &NATS{nc: nc, url: url}, nil
}

func (n *NATS) Name() string { return "nats" }

func (n *NATS) Publish(ctx context.Context, ev hub.Event) error {
	_ = ctx
	payload, err := Encode(ev)
	if err != nil {
		return err
	}
	return n.nc.Publish(Subject, payload)
}

func (n *NATS) Start(ctx context.Context, handler Handler) error {
	sub, err := n.nc.Subscribe(Subject, func(msg *nats.Msg) {
		ev, err := Decode(msg.Data)
		if err != nil {
			return
		}
		handler(ev)
	})
	if err != nil {
		return fmt.Errorf("nats subscribe: %w", err)
	}
	n.sub = sub
	go func() {
		<-ctx.Done()
		_ = sub.Unsubscribe()
	}()
	return nil
}

func (n *NATS) Close() error {
	if n.sub != nil {
		_ = n.sub.Unsubscribe()
	}
	if n.nc != nil {
		n.nc.Close()
	}
	return nil
}
