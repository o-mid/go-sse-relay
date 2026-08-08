# Changelog

## v0.1.0 — 2026-08-08

First public release.

- Ingest events with `POST /v1/hooks/{channel}`
- Live fan-out over `GET /v1/stream`
- Optional HMAC on ingest
- Per-channel replay buffer + `Last-Event-ID`
- Channel name validation
- Embedded browser demo (single binary)
- Docker image, CI, Makefile
