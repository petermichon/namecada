# Rust core

Parity-first Rust implementation of the scan engine. The Go server remains the
oracle while providers move here.

## Status

- Working: manifest loading, provider/transport contract, bounded-concurrency
  engine, fixture replay, and **51 status-only providers**.
- The check runs entirely against the committed fixtures; there is no production
  HTTP transport yet.
- Custom providers (RDAP, DNS, body parsing, multi-request, keyed) still live in
  Go and are the next porting batches.

## Layout

- `core/` — `domain-browser-core`: `Provider` trait, `StatusProvider` adapter,
  manifest model, engine, fixture replay.
- `cli/` — `domain-browser`: replay CLI used to inspect the port.

## Commands

From the repository root:

```sh
npm run rust:test
npm run rust:fmt:check
npm run rust:clippy
```

Manual replay:

```sh
cd rust
cargo run -p domain-browser -- check acema \
  --manifest ../fixtures/sources.json \
  --fixtures ../fixtures
```

`--json` emits the same result shape as the Go API.

## Parity

`core/tests/parity.rs` loads each `fixtures/acema/<source>.json`, replays the
recorded HTTP requests, and compares the Rust verdict and request count with the
Go capture. A provider that makes the wrong request or leaves a recorded request
unconsumed fails the test.

## Next

1. Add the production HTTP transport and the CLI socket-facing `serve` later.
2. Port body-parsing providers (Minehut, Wikipedia, DEV, Bluesky, …).
3. Port multi-request providers (GitLab, DEV, PeerTube, Open Collective).
4. Port RDAP + DNS (DoH for web, native DNS for Tauri).
5. Add keyed providers behind local configuration.
