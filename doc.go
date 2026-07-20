// Package ratelimit provides a concurrency-safe, in-memory fixed-window rate limiter.
//
// A Limiter tracks opaque string keys. Each key starts its own fixed window on the
// first accepted observation after the previous window expires. The implementation
// is sharded to reduce lock contention and hashes keys before storing them so the
// retained key size is bounded.
//
// The limiter is process-local. Applications that require a limit shared by
// multiple processes or machines need a coordinated external store instead.
package ratelimit
