package api

import (
	"net/http"

	"github.com/o-mid/go-sse-relay/internal/buffer"
	"github.com/o-mid/go-sse-relay/internal/config"
	"github.com/o-mid/go-sse-relay/internal/hub"
	"github.com/o-mid/go-sse-relay/internal/webui"
)

// NewMux wires HTTP routes for health, ingest, SSE, and the demo UI.
func NewMux(cfg config.Config, h *hub.Hub, ring *buffer.Ring) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.Handle("POST /v1/hooks/{channel}", &HookHandler{
		Hub:    h,
		Ring:   ring,
		Secret: cfg.HookSecret,
	})
	mux.Handle("GET /v1/stream", &StreamHandler{
		Hub:           h,
		Ring:          ring,
		SubscriberBuf: cfg.SubscriberBuf,
	})
	mux.Handle("GET /", webui.Handler())
	return mux
}
