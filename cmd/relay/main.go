package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/o-mid/go-sse-relay/internal/api"
	"github.com/o-mid/go-sse-relay/internal/buffer"
	"github.com/o-mid/go-sse-relay/internal/bus"
	"github.com/o-mid/go-sse-relay/internal/config"
	"github.com/o-mid/go-sse-relay/internal/hub"
)

func main() {
	cfg := config.FromEnv()
	h := hub.New()
	ring := buffer.New(cfg.ReplayCapacity)

	eventBus, err := bus.Open(cfg.Bus, cfg.RedisURL, cfg.NATSURL)
	if err != nil {
		log.Fatalf("bus: %v", err)
	}
	defer eventBus.Close()

	busCtx, busCancel := context.WithCancel(context.Background())
	defer busCancel()
	if err := eventBus.Start(busCtx, func(ev hub.Event) {
		if ev.Origin != "" && ev.Origin == cfg.InstanceID {
			return
		}
		published, _ := h.Publish(ev)
		ring.Add(published)
	}); err != nil {
		log.Fatalf("bus start: %v", err)
	}

	srv := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           api.NewMux(cfg, h, ring, eventBus),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("listening on http://%s bus=%s instance=%s stream_auth=%v",
			cfg.ListenAddr, eventBus.Name(), cfg.InstanceID, cfg.StreamSecret != "")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	busCancel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
