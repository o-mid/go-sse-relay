# Architecture

## Flow

1. A producer `POST`s JSON to `/v1/hooks/{channel}` (optional HMAC).
2. The relay assigns an event id, stores a short replay window, and publishes to the in-memory hub.
3. Browsers or apps connected to `GET /v1/stream?channel=...` receive Server-Sent Events.
4. On reconnect, clients send `Last-Event-ID` and receive buffered events they missed (within capacity).

## Process map

| Capability | Path |
| --- | --- |
| Channel fan-out | `internal/hub` |
| Reconnect replay buffer | `internal/buffer` |
| Optional ingest HMAC | `internal/auth` |
| HTTP ingest | `internal/api/hooks.go` |
| SSE stream | `internal/api/stream.go` |
| Process wiring | `cmd/relay` |
| Browser demo | `web/index.html` |

## Non-goals (v0.1)

- Multi-node clustering / Redis bus
- Persistent event store
- Authentication for stream subscribers beyond channel name
