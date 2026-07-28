# ratelimit

[![CI](https://github.com/azamir911/ratelimit/actions/workflows/ci.yml/badge.svg)](https://github.com/azamir911/ratelimit/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/azamir911/ratelimit.svg)](https://pkg.go.dev/github.com/azamir911/ratelimit)
[![Release](https://img.shields.io/github/v/release/azamir911/ratelimit)](https://github.com/azamir911/ratelimit/releases/latest)
[![License](https://img.shields.io/github/license/azamir911/ratelimit)](LICENSE)

A dependency-free, concurrency-safe fixed-window rate limiter for Go applications.

The core package is designed for direct embedding in services. It uses sharded maps to reduce lock contention, hashes keys to a fixed-size representation, bounds retained keys, cleans expired state, and exposes explicit lifecycle and capacity errors.

## Install

The current public release is `v0.1.0`:

```bash
go get github.com/azamir911/ratelimit@v0.1.0
```

Versions below `v1.0.0` may introduce breaking API changes between minor releases. Pin a version in production and review the changelog before upgrading.

## Quick start

```go
package main

import (
    "log"
    "time"

    "github.com/azamir911/ratelimit"
)

func main() {
    limiter, err := ratelimit.New(ratelimit.Config{
        Limit:   100,
        Window:  time.Minute,
        MaxKeys: 100_000,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer limiter.Close()

    decision, err := limiter.Allow("tenant:42")
    if err != nil {
        log.Fatal(err)
    }

    if !decision.Allowed {
        log.Printf("retry after %s", decision.RetryAfter)
    }
}
```

## API

### Configuration

| Field | Meaning | Default |
|---|---|---|
| `Limit` | Maximum accumulated cost allowed per key and window | Required |
| `Window` | Per-key fixed-window duration | Required |
| `MaxKeys` | Maximum keys retained in memory | `100000` |
| `Shards` | Number of independently locked maps | `64` |
| `CleanupInterval` | Background expired-key cleanup frequency | `Window` |
| `DisableBackgroundCleanup` | Disable the cleanup goroutine | `false` |

`Limit` follows conventional semantics: with a limit of 10, requests 1 through 10 are allowed and request 11 is blocked.

### Decisions

`Allow` and `AllowN` return a `Decision` containing:

- `Allowed`
- `Limit`
- `Count`
- `Remaining`
- `ResetAt`
- `RetryAfter`

Blocked attempts are included in `Count`. The window reset time does not move when additional requests arrive.

### Lifecycle

`Close` is idempotent and stops background cleanup. Calls already in progress may complete; calls started after closure return `ratelimit.ErrClosed`.

### Capacity

Keys are stored as SHA-256 hashes, bounding retained key size and avoiding storage of raw identifiers. `MaxKeys` bounds cardinality. When capacity is full, new keys return `ratelimit.ErrCapacity`. The request path does not perform a full-map cleanup, so overload remains `O(1)`. Expired keys are reclaimed by background cleanup or an explicit `Cleanup()` call.

This protects the process from unbounded key growth, but callers must still choose a capacity and cleanup interval appropriate for their workload.

## Standalone HTTP service

The repository includes `ratelimitd`, a standard-library HTTP service built on the same package.

```bash
go run ./cmd/ratelimitd \
  -listen :8080 \
  -limit 100 \
  -window 1m \
  -max-keys 100000
```

Check a key:

```bash
curl -sS http://localhost:8080/v1/check \
  -H 'Content-Type: application/json' \
  -d '{"key":"tenant:42","cost":1}'
```

Example response:

```json
{
  "allowed": true,
  "limit": 100,
  "count": 1,
  "remaining": 99,
  "reset_at": "2026-07-20T16:00:00Z",
  "retry_after_ms": 0
}
```

Health endpoint:

```text
GET /healthz
```

## Algorithm and guarantees

This implementation uses a per-key fixed window that starts when a key is first observed after expiry.

- Average lookup and update complexity: `O(1)`
- Cleanup complexity: `O(number of retained keys)`
- Concurrency: sharded locks, with one lock acquired per request
- Key storage: SHA-256 digest, not the raw key
- Overflow: counters use saturating arithmetic
- Time: real timestamps retain Go's monotonic component inside the process

## Important limitations

This limiter is intentionally process-local. Separate processes maintain separate counters. It is not suitable for a globally coordinated limit across replicas without an external shared system.

Fixed windows can permit a burst around a window boundary. Applications requiring smoother traffic should use a token-bucket, leaky-bucket, or sliding-window algorithm.

Hashing bounds key size but does not authenticate keys. Build keys from trusted identity and routing information rather than directly accepting arbitrary unvalidated client input.

## Development

```bash
gofmt -w .
go vet ./...
go test ./...
go test -race ./...
go test -run=^$ -fuzz=FuzzAllow -fuzztime=3s .
go test -run=^$ -bench=. -benchmem .
```

## License

MIT
