// Package bus fans events across relay instances.
package bus

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/o-mid/go-sse-relay/internal/hub"
)

// Subject is the shared pub/sub topic for cross-instance events.
const Subject = "gossrelay.events"

// Envelope is the wire format on Redis/NATS.
type Envelope struct {
	ID      string `json:"id"`
	Channel string `json:"channel"`
	Type    string `json:"type"`
	Data    []byte `json:"data"`
	Origin  string `json:"origin"`
}

// Handler is called for remote events.
type Handler func(ev hub.Event)

// Bus publishes and receives relay events between processes.
type Bus interface {
	Publish(ctx context.Context, ev hub.Event) error
	Start(ctx context.Context, handler Handler) error
	Close() error
	Name() string
}

// Encode marshals an event for the wire.
func Encode(ev hub.Event) ([]byte, error) {
	return json.Marshal(Envelope{
		ID:      ev.ID,
		Channel: ev.Channel,
		Type:    ev.Type,
		Data:    ev.Data,
		Origin:  ev.Origin,
	})
}

// Decode unmarshals a wire payload.
func Decode(b []byte) (hub.Event, error) {
	var env Envelope
	if err := json.Unmarshal(b, &env); err != nil {
		return hub.Event{}, err
	}
	if env.Channel == "" || env.Type == "" {
		return hub.Event{}, fmt.Errorf("invalid bus envelope")
	}
	return hub.Event{
		ID:      env.ID,
		Channel: env.Channel,
		Type:    env.Type,
		Data:    env.Data,
		Origin:  env.Origin,
	}, nil
}
