# HTTP API

Base URL in local defaults: `http://127.0.0.1:8080`

## `GET /healthz`

Liveness check.

```bash
curl -s http://127.0.0.1:8080/healthz
# {"status":"ok"}
```

## `POST /v1/hooks/{channel}`

Publish one event to a channel.

Request:

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

| Field | Meaning |
| --- | --- |
| `id` | Event id assigned by the relay |
| `delivered` | How many live subscribers received it right now (0 is fine) |

Common errors:

| Status | When |
| --- | --- |
| 400 | Bad channel, bad JSON, missing `type` |
| 401 | Secret configured and signature wrong/missing |
| 405 | Method is not POST |

## `GET /v1/stream?channel={channel}`

Open an SSE stream.

```bash
curl -N 'http://127.0.0.1:8080/v1/stream?channel=demo_user'
```

Optional reconnect:

```http
GET /v1/stream?channel=demo_user HTTP/1.1
Last-Event-ID: 3
```

or

```text
/v1/stream?channel=demo_user&lastEventId=3
```

Wire format example:

```text
id: 4
event: payment.updated
data: {"status":"settled","amount":"12.50"}

: keepalive
```

Keepalive comments are sent about every 15 seconds so proxies do not idle-close the socket.

## `GET /`

Embedded browser demo (connect + publish buttons).

## Environment

| Variable | Default | Purpose |
| --- | --- | --- |
| `LISTEN_ADDR` | `127.0.0.1:8080` | Bind address |
| `HOOK_SECRET` | empty | Shared secret for ingest HMAC |
| `REPLAY_CAPACITY` | `64` | Per-channel replay size |
| `SUBSCRIBER_BUFFER` | `16` | Per-subscriber channel buffer |
