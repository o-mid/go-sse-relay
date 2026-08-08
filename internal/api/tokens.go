package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/o-mid/go-sse-relay/internal/auth"
	"github.com/o-mid/go-sse-relay/internal/channel"
	"github.com/o-mid/go-sse-relay/internal/streamtoken"
)

type mintRequest struct {
	Channel    string `json:"channel"`
	TTLSeconds int    `json:"ttl_seconds"`
}

// TokenHandler mints signed stream tokens.
type TokenHandler struct {
	MintSecret   string
	StreamSecret string
	DefaultTTL   time.Duration
}

func (h *TokenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if h.StreamSecret == "" {
		http.Error(w, "stream auth disabled", http.StatusBadRequest)
		return
	}
	if h.MintSecret == "" {
		http.Error(w, "token mint secret not configured", http.StatusServiceUnavailable)
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<16))
	if err != nil {
		http.Error(w, "read body", http.StatusBadRequest)
		return
	}
	if !mintAuthorized(h.MintSecret, body, r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req mintRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	ch := strings.TrimSpace(req.Channel)
	if !channel.Valid(ch) {
		http.Error(w, "invalid channel", http.StatusBadRequest)
		return
	}
	ttl := h.DefaultTTL
	if req.TTLSeconds > 0 {
		ttl = time.Duration(req.TTLSeconds) * time.Second
	}
	tok, err := streamtoken.Mint(h.StreamSecret, ch, ttl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"token":      tok.Raw,
		"channel":    tok.Channel,
		"expires_at": tok.Exp.Format(time.RFC3339),
	})
}

func mintAuthorized(secret string, body []byte, r *http.Request) bool {
	bearer := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	if bearer != "" && bearer == secret {
		return true
	}
	sig := r.Header.Get("X-Signature")
	if sig == "" {
		sig = r.Header.Get("X-Hub-Signature-256")
	}
	return auth.ValidHMACSHA256(secret, body, sig)
}
