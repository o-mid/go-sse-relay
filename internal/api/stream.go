package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/o-mid/go-sse-relay/internal/buffer"
	"github.com/o-mid/go-sse-relay/internal/channel"
	"github.com/o-mid/go-sse-relay/internal/hub"
	"github.com/o-mid/go-sse-relay/internal/streamtoken"
)

// StreamHandler serves SSE for a channel, with optional Last-Event-ID replay.
type StreamHandler struct {
	Hub           *hub.Hub
	Ring          *buffer.Ring
	SubscriberBuf int
	StreamSecret  string
}

func (h *StreamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ch := strings.TrimSpace(r.URL.Query().Get("channel"))
	if !channel.Valid(ch) {
		http.Error(w, "invalid channel", http.StatusBadRequest)
		return
	}
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		token = strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	}
	if err := streamtoken.Verify(h.StreamSecret, token, ch, time.Now().UTC()); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
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
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	lastID := strings.TrimSpace(r.Header.Get("Last-Event-ID"))
	if lastID == "" {
		lastID = strings.TrimSpace(r.URL.Query().Get("lastEventId"))
	}
	for _, ev := range h.Ring.Since(ch, lastID) {
		writeSSE(w, ev)
	}
	flusher.Flush()

	sub := h.Hub.Subscribe(ch, h.SubscriberBuf)
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
	_, _ = fmt.Fprintf(w, "id: %s\n", sanitizeOneLine(ev.ID))
	_, _ = fmt.Fprintf(w, "event: %s\n", sanitizeOneLine(ev.Type))
	_, _ = fmt.Fprintf(w, "data: %s\n\n", sanitizeOneLine(string(ev.Data)))
}

func sanitizeOneLine(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	return s
}
