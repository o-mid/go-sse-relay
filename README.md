# go-sse-relay

Push an event in. Stream it out live.

`go-sse-relay` is a small Go service for the common “partner callback finished, client UI should update” problem. Producers `POST` JSON into a channel. Browsers or apps listen with Server-Sent Events. A short in-memory buffer helps after reconnects.

Repo: https://github.com/o-mid/go-sse-relay

## Quick start

```bash
git clone git@github.com:o-mid/go-sse-relay.git
cd go-sse-relay
cp .env.example .env
make run
```

Open http://127.0.0.1:8080/

1. Click **Connect stream**
2. Click **Publish sample event**
3. Watch the event show up in the log

From another terminal:

```bash
# listen
curl -N 'http://127.0.0.1:8080/v1/stream?channel=demo_user'

# publish
curl -s -X POST 'http://127.0.0.1:8080/v1/hooks/demo_user' \
  -H 'Content-Type: application/json' \
  -d '{"type":"payment.updated","data":{"status":"settled","amount":"12.50"}}'
```

Or: `make demo` (needs the server already running).

## Install / run without cloning

```bash
go run github.com/o-mid/go-sse-relay/cmd/relay@v0.1.0
```

```bash
go install github.com/o-mid/go-sse-relay/cmd/relay@v0.1.0
LISTEN_ADDR=127.0.0.1:8080 relay
```

## Docker

```bash
docker build -t go-sse-relay .
docker run --rm -p 8080:8080 go-sse-relay
```

## API at a glance

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/healthz` | Liveness |
| `POST` | `/v1/hooks/{channel}` | Ingest one event |
| `GET` | `/v1/stream?channel=...` | SSE subscribe |
| `GET` | `/` | Demo UI |

Ingest body:

```json
{
  "type": "payment.updated",
  "data": { "status": "settled", "amount": "12.50" }
}
```

Optional HMAC: set `HOOK_SECRET`, send hex SHA-256 of the raw body in `X-Signature`.

Reconnect: send `Last-Event-ID` (or `lastEventId`) to replay buffered events.

Full detail: [docs/api.md](docs/api.md)

## Docs

| Doc | Contents |
| --- | --- |
| [docs/concepts.md](docs/concepts.md) | Channels, fan-out, replay, limits |
| [docs/api.md](docs/api.md) | HTTP reference and env vars |
| [docs/architecture.md](docs/architecture.md) | Flow diagram and capability map |
| [CHANGELOG.md](CHANGELOG.md) | Release notes |

## Capability map

| Capability | Path |
| --- | --- |
| Channel rules | `internal/channel` |
| Fan-out hub | `internal/hub` |
| Replay buffer | `internal/buffer` |
| HMAC verify | `internal/auth` |
| Hook + SSE HTTP | `internal/api` |
| Embedded demo | `internal/webui` |
| Entrypoint | `cmd/relay` |

## Develop

```bash
make test
make vet
make build
```

## When to use it / when not to

Use it when you want a thin live pipe: ingest → fan-out → SSE, maybe for demos, local tools, or a single-node sidecar.

Do not use it alone as your multi-region event backbone. There is no shared bus and no durable log in v0.1.

## License

MIT
