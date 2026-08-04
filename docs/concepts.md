# Concepts

## Why this exists

A lot of backends learn about the real outcome of a payment, registration, or partner job **after** the first HTTP call returns. Clients still want the UI to update without hammering `GET` every second.

`go-sse-relay` sits in the middle:

- producers push events into a **channel**
- subscribers listen on that same channel over **SSE**
- a short **replay buffer** covers reconnects

It is intentionally small. One process, memory only, no database.

## Channel

A channel is just a string name, for example `demo_user` or `order:991`.

Rules:

- length 1–128
- letters, digits, `_`, `-`, `.`, `:`

Treat the channel name like a routing key. In a real deployment you usually derive it from an authenticated user or resource id in a layer **in front of** this relay.

## Event

Ingest body:

```json
{
  "type": "payment.updated",
  "data": { "status": "settled" }
}
```

- `type` becomes the SSE `event:` field
- `data` is sent as one JSON `data:` line
- the relay assigns a monotonic string `id` used for reconnect

## Fan-out

Every open stream on that channel gets the event. Slow clients with a full buffer drop events rather than blocking publishers.

## Replay

Each channel keeps the last N events (`REPLAY_CAPACITY`, default 64).

On connect, if the client sends `Last-Event-ID` (or `lastEventId`), the relay sends newer buffered events first, then stays open for live ones.

If the id is unknown, the whole buffer is replayed. That is a deliberate simple choice for v0.1.

## HMAC on ingest

When `HOOK_SECRET` is set, ingest requires a hex SHA-256 HMAC of the raw body in:

- `X-Signature`, or
- `X-Hub-Signature-256` (`sha256=<hex>` is accepted)

Streams are not authenticated in v0.1. Put the relay behind your own edge auth if channels are sensitive.

## Stream tokens

When `STREAM_SECRET` is set, subscribers must present a signed channel token. See [auth.md](auth.md).

## Multi-instance

`BUS=redis` or `BUS=nats` copies live events to other relay processes. Replay stays local. See [bus.md](bus.md).

## What this is not

- not a durable shared event log
- not a full webhook platform with retries and delivery receipts
