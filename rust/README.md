# Rust core

Parity-first Rust implementation of the scan engine. The Go server remains the
oracle while providers move here.

## Status

- Working: manifest loading, provider/transport contract, bounded-concurrency
  engine, fixture replay, **51 status-only providers**, **28 custom providers**
  (body parsing, multi-request, 6 RDAP domains), and **8 DNS providers** —
  **87 keyless sources**.
- The check runs against the committed fixtures through a replay transport; the
  production HTTP transport is not wired yet.
- Remaining: the 4 keyed sources (Discogs, Twitch, Companies House,
  OpenCorporates).

## Layout

- `core/` — `domain-browser-core`: `Provider` trait, status/custom/DNS adapters,
  manifest model, engine, fixture replay.
- `cli/` — `domain-browser`: replay CLI used to inspect the HTTP-backed port.

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

`--json` emits the same result shape as the Go API. The replay CLI covers the
HTTP-backed providers; DNS providers are exercised by the core tests because
their fixtures have no HTTP transcript.

## Parity

`core/tests/parity.rs` loads each `fixtures/acema/<source>.json`, replays the
recorded HTTP requests, and compares the Rust verdict and request count with the
Go capture. A provider that makes the wrong request or leaves a recorded request
unconsumed fails the test.

## Next

1. Add the production HTTP transport and a `serve` API.
2. Port the 4 keyed sources behind local configuration.
3. Add the Tauri desktop shell and switch the UI transport.
