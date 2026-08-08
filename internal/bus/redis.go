package bus

import (
	"context"
	"fmt"

	"github.com/o-mid/go-sse-relay/internal/hub"
	"github.com/redis/go-redis/v9"
)

// Redis fans events over Redis Pub/Sub.
type Redis struct {
	client *redis.Client
	pubsub *redis.PubSub
}

// NewRedis connects using a redis URL (redis://...).
func NewRedis(redisURL string) (*Redis, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("redis url: %w", err)
	}
	client := redis.NewClient(opt)
	if err := client.Ping(context.Background()).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return &Redis{client: client}, nil
}

func (r *Redis) Name() string { return "redis" }

func (r *Redis) Publish(ctx context.Context, ev hub.Event) error {
	payload, err := Encode(ev)
	if err != nil {
		return err
	}
	return r.client.Publish(ctx, Subject, payload).Err()
}

func (r *Redis) Start(ctx context.Context, handler Handler) error {
	r.pubsub = r.client.Subscribe(ctx, Subject)
	if _, err := r.pubsub.Receive(ctx); err != nil {
		return fmt.Errorf("redis subscribe: %w", err)
	}
	ch := r.pubsub.Channel()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				ev, err := Decode([]byte(msg.Payload))
				if err != nil {
					continue
				}
				handler(ev)
			}
		}
	}()
	return nil
}

func (r *Redis) Close() error {
	var err error
	if r.pubsub != nil {
		err = r.pubsub.Close()
	}
	if r.client != nil {
		if e := r.client.Close(); e != nil && err == nil {
			err = e
		}
	}
	return err
}
