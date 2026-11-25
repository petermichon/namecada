# Architecture

How a check runs, where the code lives, and how to add a source.

## How it works

Each source runs independently and streams its verdict into the report as soon
as it resolves, so one slow or failing source never blocks the others. Nothing
waits for the whole set.

The app is a static client of one server process: `GET /v1/check?name=<label>`
streams NDJSON events — a `manifest`, one `result` per source, then `done` or
`pending`. The manifest lists every source (checked, keyed or deferred) with its
URL, so the page renders the full report before checks finish and fills rows in
place.

A scan is a **resumable job**, not a request. The job runs the whole source list
in the background and keeps every verdict it resolves; a request streams what
the job already has for a few seconds, then closes with
`{"type":"pending","scan":…,"cursor":…}` and the client asks for the next page.
A 30-second scan budget therefore no longer decides what is knowable: sources the
request did not reach stay `pending` and are collected on the following page.
`unknown` is reserved for a source that answered and could not be interpreted.

Development uses Vite with `/v1` proxied to `serve`. Vite is pinned to port
5221 with `--strictPort`, so the desktop shell always attaches to this project's
dev server rather than whatever else is on a default port.

`core` has no HTTP dependency. `serve` is a development/CLI harness only; the
packaged app is Tauri (`apps/desktop/src-tauri/`), calling the engine directly and feeding
the same events through the `CheckClient` seam, with no localhost server. The
web client picks its transport once at load: `invoke('check')` with an IPC
channel inside the desktop shell, the NDJSON endpoint everywhere else. The
desktop scans one label on one device, so it streams the whole scan over the one
channel; the HTTP server pages it, because a request there has a response budget
and a client that may disconnect. The
source registry is compiled into the binary (`Manifest::embedded`);
`--manifest` overrides it for development.

Desktop app: `npm run tauri:dev` (Vite dev server plus the native window) and
`npm run tauri:build` (bundles the single-file web build). The shell keeps its
verdict cache in `cache.json` under `app_data_dir()`: it loads that file at
startup and rewrites it after each scan, so a relaunch within the ttl reuses the
previous verdicts instead of rescanning.

## Rust core

`packages/core` is the engine; `apps/cli` is a development harness around it.
The root Cargo workspace holds both:

- `packages/core/src/provider.rs` — the `Provider` and `Transport` contracts and
  the result/status types.
- `packages/core/src/adapters/` — one module per source shape: `status.rs`
  (verdict is the status code), `api.rs` (JSON/XML/GraphQL endpoints), `page.rs`
  (server-rendered HTML), `domains.rs` (RDAP), `dns.rs` (wildcard subdomains).
  `shared.rs` holds the request helpers and `spec.rs` the id → check table.
- `packages/core/src/registry.rs` — the compiled registry: live providers and stats.
- `packages/core/src/engine.rs` — bounded-concurrency scan engine with its one-hour
  verdict cache; streams each result as it resolves. `DEFAULT_CONCURRENCY`
  caps a scan at 32 checks in flight and `DEFAULT_DEADLINE` closes one pass at
  30 s, reporting anything unresolved as `pending`. A non-definitive answer
  (transport error or `unknown`) gets one retry after 200 ms before it is
  reported, because most first-pass misses are a burst or a reset.
- `packages/core/src/jobs.rs` — resumable `ScanJob`s: one scan per label that
  outlives a response, plus the `Jobs` registry the transports keep.
- `packages/core/src/http.rs` — production HTTP transport: a 10 s global bound with
  4 s connect and 5 s read bounds so a stalled host frees its worker. It connects
  over IPv4 first and retries with every address family when nothing connected,
  because a resolver that returns AAAA first lets a broken IPv6 route eat the
  connect budget.
- `apps/cli/src/bin/serve.rs` — the NDJSON HTTP server; `capture.rs` records
  fixtures; `probe.rs` checks the sources that have no provider yet.

Two development binaries drive it:

- `npm run rust:serve` runs the `serve` binary, exposing
  `GET /v1/check?name=<label>`: a manifest, one result per source as it
  resolves, then a done event.
- `npm run capture` runs the `capture` binary, recording fixtures from live
  providers (`--source id` limits it to one source).

There is no one-shot CLI check; `curl 'localhost:8080/v1/check?name=acema'`
against `serve` is the same thing. Fixtures replay in the test suite.

The credential-gated sources (Discogs, Companies House, OpenCorporates)
are intentionally omitted. They produce no row and make no
request.

The live source is always the source of truth. The cache is a record of what a
source last reported, not an authoritative result: a row can be stale,
rate-limited, or wrong, and a domain row captures registry state rather than
registrability, so re-check the source rather than trusting the cache.

### Fixtures

`fixtures/sources.json` is the complete source inventory, and
`fixtures/<label>/<source>.json` holds each provider's captured result and the
raw HTTP request/response it made. The checked-in labels are `acema` (mostly
available) and `google` (mostly taken), so both verdict branches are exercised.

Fixtures are a **capture**, not live truth: responses change. `npm run capture`
records them from the Rust core (see `fixtures/README.md` for the flags and the
format), so refresh a fixture that has gone stale rather than treating an old
response as current.

## Caching

`serve` caches results per provider, keyed `provider:label` (e.g. `lemy:acema`),
with a 1-hour TTL. Only definitive verdicts are cached — `unknown` and `pending`
are never stored, so transient rate-limit or network errors self-heal and a
deadline cut is retried on the next page. A fresh entry is served without
touching the network.

This matters most for sources like GitHub and npm, which rate-limit
unauthenticated requests (GitHub: 60/hour per IP).

## Layout

- `packages/core/` — the engine: `Provider`/`Transport` contracts, adapters,
  registry, engine, cache, HTTP transport, events and fixture replay.
- `apps/cli/` — the `serve`, `capture` and `probe` binaries. It and
  `packages/core` make up the root Cargo workspace; see `packages/README.md`.
- `apps/desktop/src-tauri/` — the Tauri shell, its own Cargo workspace, driven
  by `apps/desktop/package.json`.
- `apps/web/` — Svelte 5 + Vite + Tailwind + shadcn-svelte (maia) app.
  - `src/lib/platforms.ts` — the result types (`Status`, `ProviderResult`),
    `toLabel`, and `suggestLabel`. The display name in `platformName` and the
    `paid` flag come from the manifest stream; the report renders one
    alphabetical list with a cost filter and marks paid rows.
  - `src/lib/api.ts` — the event contract and the `CheckClient` transport seam.
    `streamFromApi` (the `httpCheckClient`) consumes the `serve` API
    (`GET /v1/check?name=<label>`, NDJSON events); a native shell can inject its
    own client without touching the UI or lookup state.
  - `src/lib/recent.ts` — the recent labels, kept in `localStorage`
    (case-insensitively deduped, newest first) and separate from the engine's
    verdict cache so a label outlives the one-hour ttl. The report header shows
    them in a menu with a clear action.
- `apps/site/` — the static download page for `namecada.com`: plain HTML/CSS
  bundled by Vite into `dist/`, an nginx image and the compose file for the
  host stack. It has no backend; the product is the desktop and Android app.
  It is its own package (`apps/site/package.json`, its own lockfile, not a root
  workspace), so the image build installs Vite and nothing from `apps/web` or
  Tauri.
- `fixtures/` — captured source inventory and raw HTTP snapshots; see
  [Fixtures](#fixtures).

## Development

Rust commands need `cargo` on `PATH` (install with rustup, then open a new
shell or `source "$HOME/.cargo/env"`). The workspace pins stable Rust in
`rust-toolchain.toml`.

```sh
npm install        # installs the web workspace + git hooks
npm run dev        # Vite + the development server
npm run dev:host # proxied mode: fixed port for the local dev proxy and phone
npm run build      # single-file app build
npm test           # vitest
npm run check      # svelte-check + tsc
npm run lint       # prettier --check + eslint
npm run format     # prettier --write + cargo fmt
npm run verify     # web check/test/lint + Rust fmt/clippy/test
npm run rust:test  # Rust tests
npm run rust:clippy  # Rust lints
npm run rust:build  # release build
npm run rust:serve  # development NDJSON server on 127.0.0.1:8080
```

A husky `pre-commit` hook (installed by `npm install` via the `prepare`
script) runs lint-staged, which formats and lints staged web files
(prettier + eslint) and formats staged Rust files with `cargo fmt`.

`npm run dev` starts the Vite app and the development server. The app calls `/v1` on
its own origin and Vite proxies that to `serve` on `:8080`, so no CORS setup is
needed and Vite can fall back to another port. Use `npm run dev:web` when the
server is already running separately.

`npm run dev:host` is the proxied mode. It starts the same two processes but binds
the web app to `127.0.0.1:5220` so a machine-level dev proxy can expose it on a
local hostname; a phone reaches the same URL through one `adb reverse`. The Rust
`serve` on `:8080` stays Vite's proxy target as before.

The site is a separate, smaller project: `npm --prefix apps/site install` once,
then `npm run site:dev` (hot reload on `:5222`) or `npm run site:dev:host`
(loopback, so a dev proxy can route a hostname to it). `npm run site:build`
writes `apps/site/dist`, which is the only thing the nginx image copies; the
HTML and CSS stay framework-free. With the machine-level dev proxy used here
(devhost, `~/.config/dev`), the two surfaces are
`http://app.namecada.localhost:7357` (app, upstream `5220`) and
`http://namecada.localhost:7357` (site, upstream `5222`): the bare name is the
page a visitor would see, as in production — one hostname and one `adb reverse`
per surface, so the phone opens the same URLs.

## Adding a source

1. Add the source metadata to `fixtures/sources.json`: `id`, `platform`,
   `platformName`, `nameTemplate`, `urlTemplate`, `paid`, `minLength`,
   `maxLength`, and `keyed`. A source with no provider is listed as
   "not checked" automatically — `checked` is derived from the provider
   registry, never stored. The registry is compiled into the binary, so rebuild
   after editing.
2. Implement the check in `packages/core`:
   - a status-only source goes in `packages/core/src/adapters/status.rs`;
   - a JSON/XML/GraphQL source goes in `adapters/api.rs`;
   - a server-rendered HTML source goes in `adapters/page.rs`;
   - an RDAP or DNS source goes in `adapters/domains.rs` / `adapters/dns.rs`.
3. Register it in the corresponding spec table.
4. Add a unit test or capture fixture. Status-only and custom sources are
   exercised by `packages/core/tests/parity.rs`; DNS providers use a fake resolver.
5. Run `npm run verify`.

Checks run in the Rust core, so CORS is not a constraint. Keyed sources are
deliberately omitted for now.

Deferred: Packagist, Docker Hub, Go proxy (modules are paths, not single
labels), and the remaining trademark / social / company registries that need a
key.
