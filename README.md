# go-sse-relay

Small Go service that accepts partner-style webhook events and fans them out to live clients over **Server-Sent Events (SSE)**.

Useful when a backend receives async callbacks and mobile/web clients need live status updates without polling.

## Quick start

```bash
cp .env.example .env
make run
```

Open [http://127.0.0.1:8080/](http://127.0.0.1:8080/), click **Connect stream**, then **Publish sample event**.

Or from another terminal:

```bash
make demo
curl -N 'http://127.0.0.1:8080/v1/stream?channel=demo_user'
```

## HTTP API

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/healthz` | Liveness |
| `POST` | `/v1/hooks/{channel}` | Ingest JSON event |
| `GET` | `/v1/stream?channel=...` | SSE subscribe |
| `GET` | `/` | Browser demo |

### Ingest body

```json
{
  "type": "payment.updated",
  "data": { "status": "settled", "amount": "12.50" }
}
```

Optional HMAC: set `HOOK_SECRET` and send hex SHA-256 HMAC of the raw body in `X-Signature` (or `sha256=<hex>` / `X-Hub-Signature-256`).

### Reconnect

Clients may send `Last-Event-ID` (SSE standard) or `lastEventId` query param. The relay replays buffered events after that id (see `REPLAY_CAPACITY`).

## Capability map

| Capability | Path |
| --- | --- |
| Fan-out hub | `internal/hub` |
| Replay ring buffer | `internal/buffer` |
| HMAC verify | `internal/auth` |
| Hook + SSE HTTP | `internal/api` |
| Entrypoint | `cmd/relay` |
| Demo UI | `web/index.html` |
| Architecture notes | `docs/architecture.md` |

## Develop

```bash
make test
make vet
make build
```

## Docker

```bash
docker build -t go-sse-relay .
docker run --rm -p 8080:8080 go-sse-relay
```

## License

MIT
