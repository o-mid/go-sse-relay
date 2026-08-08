# Architecture

```
                 POST /v1/hooks/{channel}
                          |
                          v
                 +------------------+
                 | HMAC (optional)  |
                 +--------+---------+
                          |
                          v
                 +------------------+     +----------------+
                 | assign id        |---->| ring buffer    |
                 | hub.Publish      |     | (local replay) |
                 +--------+---------+     +--------+-------+
                          |                        |
                          | live                   | on connect (+ token)
                          v                        v
                 +------------------+     GET /v1/stream
                 | local SSE subs   |<----------------
                 +--------+---------+
                          |
                          | BUS=redis|nats
                          v
                 +------------------+
                 | other instances  |
                 | hub.Publish      |
                 +------------------+
```

## Capability map

| Capability | Path |
| --- | --- |
| Channel name rules | `internal/channel` |
| Fan-out hub | `internal/hub` |
| Replay ring buffer | `internal/buffer` |
| Ingest HMAC | `internal/auth` |
| Stream tokens | `internal/streamtoken` |
| Bus (memory/redis/nats) | `internal/bus` |
| Hook HTTP | `internal/api/hooks.go` |
| Token mint HTTP | `internal/api/tokens.go` |
| SSE HTTP | `internal/api/stream.go` |
| Embedded demo UI | `internal/webui` |
| Process entrypoint | `cmd/relay` |

## Process notes

- Single binary with embedded demo UI.
- Hub and ring are in memory per process.
- Bus carries live events across instances; replay is still local.
- Nodes ignore bus messages with their own `INSTANCE_ID` origin.

## Non-goals (still)

- Durable shared event log
- Stream auth that is more than a channel-bound HMAC token
- Outbound webhook delivery retries
