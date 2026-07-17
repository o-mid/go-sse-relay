package api_test

import (
	"bufio"
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/o-mid/go-sse-relay/internal/api"
	"github.com/o-mid/go-sse-relay/internal/buffer"
	"github.com/o-mid/go-sse-relay/internal/config"
	"github.com/o-mid/go-sse-relay/internal/hub"
)

func TestHealthz(t *testing.T) {
	mux := newTestMux("")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestHookRequiresValidSignatureWhenSecretSet(t *testing.T) {
	mux := newTestMux("topsecret")
	body := []byte(`{"type":"payment.updated","data":{"status":"settled"}}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/hooks/user_1", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d want 401", rec.Code)
	}

	mac := hmac.New(sha256.New, []byte("topsecret"))
	_, _ = mac.Write(body)
	sig := hex.EncodeToString(mac.Sum(nil))
	req = httptest.NewRequest(http.MethodPost, "/v1/hooks/user_1", bytes.NewReader(body))
	req.Header.Set("X-Signature", sig)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestStreamReceivesPublishedEvent(t *testing.T) {
	cfg := config.Config{SubscriberBuf: 4, ReplayCapacity: 8}
	h := hub.New()
	ring := buffer.New(cfg.ReplayCapacity)
	mux := api.NewMux(cfg, h, ring)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	streamReq := httptest.NewRequestWithContext(ctx, http.MethodGet, "/v1/stream?channel=user_9", nil)
	streamRec := httptest.NewRecorder()
	done := make(chan struct{})
	go func() {
		defer close(done)
		mux.ServeHTTP(streamRec, streamReq)
	}()

	deadline := time.Now().Add(2 * time.Second)
	for h.SubscriberCount("user_9") == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if h.SubscriberCount("user_9") == 0 {
		t.Fatal("subscriber never registered")
	}

	body := []byte(`{"type":"payment.updated","data":{"status":"settled"}}`)
	hookReq := httptest.NewRequest(http.MethodPost, "/v1/hooks/user_9", bytes.NewReader(body))
	hookRec := httptest.NewRecorder()
	mux.ServeHTTP(hookRec, hookReq)
	if hookRec.Code != http.StatusAccepted {
		t.Fatalf("hook status=%d", hookRec.Code)
	}

	deadline = time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if strings.Contains(streamRec.Body.String(), "event: payment.updated") &&
			strings.Contains(streamRec.Body.String(), `"status":"settled"`) {
			cancel()
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	cancel()
	t.Fatalf("stream body missing event: %q", streamRec.Body.String())
}

func TestReplayUsesLastEventID(t *testing.T) {
	cfg := config.Config{SubscriberBuf: 4, ReplayCapacity: 8}
	h := hub.New()
	ring := buffer.New(cfg.ReplayCapacity)
	mux := api.NewMux(cfg, h, ring)

	body := []byte(`{"type":"one","data":{"n":1}}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/hooks/replay", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	var first map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &first); err != nil {
		t.Fatal(err)
	}
	id1, _ := first["id"].(string)
	if id1 == "" {
		t.Fatal("missing id")
	}

	body = []byte(`{"type":"two","data":{"n":2}}`)
	req = httptest.NewRequest(http.MethodPost, "/v1/hooks/replay", bytes.NewReader(body))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	streamReq := httptest.NewRequestWithContext(ctx, http.MethodGet, "/v1/stream?channel=replay", nil)
	streamReq.Header.Set("Last-Event-ID", id1)
	streamRec := httptest.NewRecorder()
	mux.ServeHTTP(streamRec, streamReq)

	var events []string
	sc := bufio.NewScanner(strings.NewReader(streamRec.Body.String()))
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "event: ") {
			events = append(events, strings.TrimPrefix(line, "event: "))
		}
	}
	if len(events) < 1 || events[0] != "two" {
		t.Fatalf("events=%v body=%q", events, streamRec.Body.String())
	}
	for _, ev := range events {
		if ev == "one" {
			t.Fatalf("replay included event before Last-Event-ID: %v", events)
		}
	}
}

func newTestMux(secret string) http.Handler {
	cfg := config.Config{HookSecret: secret, SubscriberBuf: 4, ReplayCapacity: 8}
	return api.NewMux(cfg, hub.New(), buffer.New(cfg.ReplayCapacity))
}
