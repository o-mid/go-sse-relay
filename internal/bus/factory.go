package bus

import (
	"fmt"
	"strings"
)

// Open builds a bus from kind + connection URL.
// kind: memory | redis | nats
func Open(kind, redisURL, natsURL string) (Bus, error) {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "", "memory", "none", "local":
		return NewMemory(), nil
	case "redis":
		if redisURL == "" {
			return nil, fmt.Errorf("REDIS_URL required when BUS=redis")
		}
		return NewRedis(redisURL)
	case "nats":
		return NewNATS(natsURL)
	default:
		return nil, fmt.Errorf("unknown BUS %q (use memory, redis, or nats)", kind)
	}
}
