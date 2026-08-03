package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/o-mid/go-sse-relay/internal/buffer"
	"github.com/o-mid/go-sse-relay/internal/bus"
	"github.com/o-mid/go-sse-relay/internal/config"
	"github.com/o-mid/go-sse-relay/internal/hub"
	"github.com/o-mid/go-sse-relay/internal/webui"
)

// NewMux wires HTTP routes for health, ingest, tokens, SSE, and the demo UI.
func NewMux(cfg config.Config, h *hub.Hub, ring *buffer.Ring, eventBus bus.Bus) http.Handler {
	if eventBus == nil {
		eventBus = bus.NewMemory()
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":      "ok",
			"bus":         eventBus.Name(),
			"instance_id": cfg.InstanceID,
			"stream_auth": cfg.StreamSecret != "",
		})
	})
	mux.Handle("POST /v1/hooks/{channel}", &HookHandler{
		Hub:        h,
		Ring:       ring,
		Bus:        eventBus,
		Secret:     cfg.HookSecret,
		InstanceID: cfg.InstanceID,
	})
	mux.Handle("POST /v1/tokens", &TokenHandler{
		MintSecret:   cfg.TokenMintSecret,
		StreamSecret: cfg.StreamSecret,
		DefaultTTL:   time.Duration(cfg.DefaultTokenTTL) * time.Second,
	})
	mux.Handle("GET /v1/stream", &StreamHandler{
		Hub:           h,
		Ring:          ring,
		SubscriberBuf: cfg.SubscriberBuf,
		StreamSecret:  cfg.StreamSecret,
	})
	mux.Handle("GET /", webui.Handler())
	return mux
}
