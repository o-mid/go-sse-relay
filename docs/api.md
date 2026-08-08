# HTTP API

Base URL in local defaults: `http://127.0.0.1:8080`

## `GET /healthz`

Liveness check.

```bash
curl -s http://127.0.0.1:8080/healthz
```

Example:

```json
{"status":"ok","bus":"memory","instance_id":"relay-1","stream_auth":false}
```

## `POST /v1/hooks/{channel}`

Publish one event to a channel.

```http
POST /v1/hooks/demo_user HTTP/1.1
Content-Type: application/json
X-Signature: <optional hex hmac>

{"type":"payment.updated","data":{"status":"settled","amount":"12.50"}}
```

Success (`202`):

```json
{"status":"accepted","id":"3","delivered":1}
```

`delivered` is the local subscriber count on this instance only.

| Status | When |
| --- | --- |
| 400 | Bad channel, bad JSON, missing `type` |
| 401 | Secret configured and signature wrong/missing |
| 405 | Method is not POST |

## `POST /v1/tokens`

Mint a signed stream token. See [auth.md](auth.md).

## `GET /v1/stream?channel={channel}`

Open an SSE stream. When `STREAM_SECRET` is set, pass `token=` or `Authorization: Bearer`.

```bash
curl -N 'http://127.0.0.1:8080/v1/stream?channel=demo_user&token=v1....'
```

Reconnect with `Last-Event-ID` or `lastEventId`.

Keepalive comments are sent about every 15 seconds.

## `GET /`

Embedded browser demo.

## Environment

| Variable | Default | Purpose |
| --- | --- | --- |
| `LISTEN_ADDR` | `127.0.0.1:8080` | Bind address |
| `INSTANCE_ID` | `relay-1` | Origin id for bus echo suppression |
| `HOOK_SECRET` | empty | Ingest HMAC |
| `STREAM_SECRET` | empty | Enable stream token checks |
| `TOKEN_MINT_SECRET` | falls back | Mint auth for `/v1/tokens` |
| `TOKEN_TTL_SECONDS` | `3600` | Default token TTL |
| `REPLAY_CAPACITY` | `64` | Per-channel replay size |
| `SUBSCRIBER_BUFFER` | `16` | Per-subscriber buffer |
| `BUS` | `memory` | `memory`, `redis`, or `nats` |
| `REDIS_URL` | empty | Required for `BUS=redis` |
| `NATS_URL` | `nats://127.0.0.1:4222` | Used for `BUS=nats` |
