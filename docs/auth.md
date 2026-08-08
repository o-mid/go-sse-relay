# Stream auth

When `STREAM_SECRET` is empty (default), anyone who knows a channel name can subscribe. That is fine for local demos.

When `STREAM_SECRET` is set, `GET /v1/stream` requires a signed token for that channel.

## Token format

```text
v1.<channel>.<exp_unix>.<hmac_sha256_hex>
```

The HMAC covers `v1|<channel>|<exp_unix>` using `STREAM_SECRET`.

## Mint a token

`POST /v1/tokens`

Auth (one of):

- `Authorization: Bearer <TOKEN_MINT_SECRET>`
- `X-Signature: <hex hmac of body with TOKEN_MINT_SECRET>`

`TOKEN_MINT_SECRET` falls back to `HOOK_SECRET`, then `STREAM_SECRET`.

```bash
curl -s -X POST http://127.0.0.1:8080/v1/tokens \
  -H 'Authorization: Bearer mint-secret' \
  -H 'Content-Type: application/json' \
  -d '{"channel":"demo_user","ttl_seconds":3600}'
```

Response:

```json
{
  "token": "v1.demo_user.1893456000.ab...",
  "channel": "demo_user",
  "expires_at": "2030-01-01T00:00:00Z"
}
```

## Subscribe with a token

Query param:

```bash
curl -N "http://127.0.0.1:8080/v1/stream?channel=demo_user&token=v1...."
```

Or header (for clients that can set headers; browsers' `EventSource` cannot, so the demo uses the query param):

```http
Authorization: Bearer v1....
```

## Notes

- Tokens are channel-bound. A token for `a` cannot open `b`.
- Put minting behind your own auth in production. This relay only checks the mint secret.
- Ingest HMAC (`HOOK_SECRET`) is separate from stream tokens.
