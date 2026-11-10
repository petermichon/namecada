# Rust core

The scan engine and server. Providers are ported; the checked-in fixtures are
the parity oracle.

## Status

- Working: manifest loading, provider/transport contract, bounded-concurrency
  engine, fixture replay, **51 status-only providers**, **28 custom providers**
  (body parsing, multi-request, 6 RDAP domains), and **8 DNS providers** —
  **87 keyless sources**.
- Production HTTP transport (`ureq`), the in-memory cache, and the NDJSON
  `serve` API are wired.
- The 4 keyed sources (Discogs, Twitch, Companies House, OpenCorporates) are
  deliberately out of scope for now and omitted entirely.

## Layout

- `core/` — `namecheck-core`: `Provider`/`Transport` traits,
  status/custom/DNS adapters, manifest, engine, cache, HTTP transport, replay.
- `cli/` — `namecheck`: `serve` and `check` commands.

## Commands

Requires `cargo` on `PATH` (rustup, then a new shell or
`source "$HOME/.cargo/env"`). `rust-toolchain.toml` pins stable and installs
`rustfmt`/`clippy`.

From the repository root:

```sh
npm run rust:test
npm run rust:fmt:check
npm run rust:clippy
npm run rust:build
```

Manual replay:

```sh
cd rust
cargo run -p namecheck -- check acema \
  --manifest ../fixtures/sources.json \
  --fixtures ../fixtures
```

Live check with real HTTP and DNS:

```sh
cargo run -p namecheck -- check acema \
  --live --manifest ../fixtures/sources.json
```

Serve the NDJSON API (`GET /check?name=<label>`), from the repository root:

```sh
npm run rust:serve
```

The Vite client defaults to `http://localhost:8080`. For a different port, set
`VITE_API_BASE`.

`--json` emits the same result shape the web client consumes. The replay CLI
covers the HTTP-backed providers; DNS providers are exercised by the core tests
because their fixtures have no HTTP transcript.

## Parity

`core/tests/parity.rs` loads each `fixtures/<label>/<source>.json`, replays the
recorded HTTP requests, and compares the Rust verdict and request count with the
captured fixture. A provider that makes the wrong request or leaves a recorded
request unconsumed fails the test.

## Next

1. Keep the keyed sources omitted unless they become a requirement.
2. Polish the Rust server and web integration.
3. Add the Tauri desktop shell only if it is still wanted.
