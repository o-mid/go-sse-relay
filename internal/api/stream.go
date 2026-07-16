package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/o-mid/go-sse-relay/internal/buffer"
	"github.com/o-mid/go-sse-relay/internal/hub"
)

// StreamHandler serves SSE for a channel, with optional Last-Event-ID replay.
type StreamHandler struct {
	Hub           *hub.Hub
	Ring          *buffer.Ring
	SubscriberBuf int
}

func (h *StreamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	channel := strings.TrimSpace(r.URL.Query().Get("channel"))
	if channel == "" {
		http.Error(w, "channel required", http.StatusBadRequest)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	lastID := strings.TrimSpace(r.Header.Get("Last-Event-ID"))
	if lastID == "" {
		lastID = strings.TrimSpace(r.URL.Query().Get("lastEventId"))
	}
	for _, ev := range h.Ring.Since(channel, lastID) {
		writeSSE(w, ev)
	}
	flusher.Flush()

	sub := h.Hub.Subscribe(channel, h.SubscriberBuf)
	defer h.Hub.Unsubscribe(sub)

	ctx := r.Context()
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-sub.C:
			if !ok {
				return
			}
			writeSSE(w, ev)
			flusher.Flush()
		case <-ticker.C:
			_, _ = fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}

func writeSSE(w http.ResponseWriter, ev hub.Event) {
	_, _ = fmt.Fprintf(w, "id: %s\n", ev.ID)
	_, _ = fmt.Fprintf(w, "event: %s\n", ev.Type)
	_, _ = fmt.Fprintf(w, "data: %s\n\n", string(ev.Data))
}
