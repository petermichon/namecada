# Rust packages

The scan engine and the development harness. Providers are ported; the
checked-in fixtures are the parity oracle.

## Status

- Working: manifest loading, the provider/transport contract, the
  bounded-concurrency engine, fixture replay, **107 keyless providers** — 63
  status-only, 36 custom (JSON/XML/GraphQL, server-rendered HTML, 6 RDAP
  domains) and 8 DNS wildcard subdomains.
- The production HTTP transport (`ureq`), the one-hour verdict cache and the
  NDJSON `serve` API are wired.
- The 3 keyed sources (Discogs, Companies House, OpenCorporates) are
  deliberately out of scope and omitted entirely: no row, request or cache
  entry.

## Layout

- `packages/core/` — `namecada-core`: the `Provider`/`Transport` traits,
  `adapters/` (status, custom, domains, DNS), the manifest, engine, cache, HTTP
  transport and fixture replay.
- `apps/cli/` — the `namecada` crate: the `serve`, `capture` and `probe`
  binaries.

Both are members of the root Cargo workspace. The desktop shell
(`apps/desktop/src-tauri`) is a separate workspace that depends on
`packages/core` by path.

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
npm run rust:serve
```

The registry is compiled in (`Manifest::embedded`); `--manifest
fixtures/sources.json` overrides it. Fixtures are replayed by the parity tests.
