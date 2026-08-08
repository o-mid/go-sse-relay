package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/o-mid/go-sse-relay/internal/auth"
	"github.com/o-mid/go-sse-relay/internal/buffer"
	"github.com/o-mid/go-sse-relay/internal/channel"
	"github.com/o-mid/go-sse-relay/internal/hub"
)

type hookEnvelope struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// HookHandler accepts partner-style POST events and fans them out.
type HookHandler struct {
	Hub    *hub.Hub
	Ring   *buffer.Ring
	Secret string
}

func (h *HookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ch := strings.TrimSpace(r.PathValue("channel"))
	if !channel.Valid(ch) {
		http.Error(w, "invalid channel", http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "read body", http.StatusBadRequest)
		return
	}
	sig := r.Header.Get("X-Signature")
	if sig == "" {
		sig = r.Header.Get("X-Hub-Signature-256")
	}
	if !auth.ValidHMACSHA256(h.Secret, body, sig) {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	var env hookEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(env.Type) == "" {
		http.Error(w, "type required", http.StatusBadRequest)
		return
	}
	data := env.Data
	if len(data) == 0 {
		data = json.RawMessage(`{}`)
	}

	published, delivered := h.Hub.Publish(hub.Event{
		Channel: ch,
		Type:    env.Type,
		Data:    append([]byte(nil), data...),
	})
	h.Ring.Add(published)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":    "accepted",
		"id":        published.ID,
		"delivered": delivered,
	})
}
