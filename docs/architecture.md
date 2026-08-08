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
                 | hub.Publish      |     | (replay)       |
                 +--------+---------+     +--------+-------+
                          |                        |
                          | live                   | on connect
                          v                        v
                 +------------------+     GET /v1/stream
                 | subscribers SSE  |<----------------
                 +------------------+
```

## Capability map

| Capability | Path |
| --- | --- |
| Channel name rules | `internal/channel` |
| Fan-out hub | `internal/hub` |
| Replay ring buffer | `internal/buffer` |
| HMAC verify | `internal/auth` |
| Hook HTTP | `internal/api/hooks.go` |
| SSE HTTP | `internal/api/stream.go` |
| Embedded demo UI | `internal/webui` |
| Process entrypoint | `cmd/relay` |

## Process notes

- Single binary. The demo UI is embedded, so the working directory does not matter.
- Hub and ring live in memory. Restart clears subscribers and replay.
- Publish never blocks on a slow subscriber; that subscriber may miss events if its buffer is full.
- Good fit behind one instance or as a local/dev relay. Multi-node needs a shared bus (later work).

## Non-goals (v0.1)

- Multi-node clustering
- Persistent event store
- Stream authentication beyond knowing the channel name
- Outbound webhook delivery retries
