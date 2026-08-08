# Multi-instance bus

Single process mode uses an in-memory hub only (`BUS=memory`).

To run more than one relay instance, set a shared bus so an ingest on node A reaches subscribers on node B.

## Modes

| `BUS` | Needs | Notes |
| --- | --- | --- |
| `memory` | nothing | Default. One process. |
| `redis` | `REDIS_URL` | Redis Pub/Sub on subject `gossrelay.events` |
| `nats` | `NATS_URL` | NATS pub/sub on the same subject name |

## Local brokers

```bash
docker compose up -d redis
# or
docker compose up -d nats
```

## Two instances with Redis

Terminal A:

```bash
INSTANCE_ID=relay-a BUS=redis REDIS_URL=redis://127.0.0.1:6379/0 LISTEN_ADDR=127.0.0.1:8080 make run
```

Terminal B:

```bash
INSTANCE_ID=relay-b BUS=redis REDIS_URL=redis://127.0.0.1:6379/0 LISTEN_ADDR=127.0.0.1:8081 make run
```

Subscribe on B, publish on A:

```bash
curl -N 'http://127.0.0.1:8081/v1/stream?channel=demo_user'
curl -s -X POST 'http://127.0.0.1:8080/v1/hooks/demo_user' \
  -H 'Content-Type: application/json' \
  -d '{"type":"payment.updated","data":{"status":"settled"}}'
```

## How echo is avoided

Each event carries `origin` = `INSTANCE_ID`. A node ignores bus messages that it published itself. Local subscribers already got the event from the local hub.

## Limits

- Live fan-out crosses instances. Replay buffers stay **per process**. Sticky sessions help reconnect; shared durable replay is not in this version.
- Redis/NATS must be reachable from every instance.
- There is no ACL on the bus subject — keep the broker private.
