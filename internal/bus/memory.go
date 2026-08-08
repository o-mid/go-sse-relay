package bus

import (
	"context"

	"github.com/o-mid/go-sse-relay/internal/hub"
)

// Memory is a no-op bus for single-instance mode.
type Memory struct{}

// NewMemory returns a single-process bus.
func NewMemory() *Memory { return &Memory{} }

func (m *Memory) Name() string { return "memory" }

func (m *Memory) Publish(context.Context, hub.Event) error { return nil }

func (m *Memory) Start(context.Context, Handler) error { return nil }

func (m *Memory) Close() error { return nil }
