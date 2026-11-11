# AGENTS.md

Guidance for agents working in this repo. See `README.md` for architecture and
`docs/platforms.md` for the source catalogue; this file covers commands and
conventions.

## Layout

- `web/` — Svelte 5 + Vite + Tailwind + shadcn-svelte app. A thin client of the
  server (streams `GET /v1/check?name=<label>` NDJSON events:
  `manifest` → `result`* → `done`). The UI depends on the `CheckClient` seam in
  `src/lib/api.ts`; `httpCheckClient` is the development adapter.
- `rust/` — the engine plus a development/CLI harness. `core/` holds the `Provider`/`Transport`
  contracts, the status/custom/DNS adapters, the engine, cache, HTTP transport,
  and fixture replay. `cli/` holds the `serve` and `check` commands.
- `fixtures/` — captured source inventory plus raw HTTP snapshots. These are a
  frozen parity oracle; the Rust-only tree has no exporter yet. See
  `fixtures/README.md`.
- `docs/platforms.md` — the full source catalogue and per-source notes.

## Commands

Install once (also wires git hooks): `npm install`.

Root scripts (run from the repo root):

| Command | Does |
| --- | --- |
| `npm run dev` | Vite dev server **and** the Rust development server, concurrently |
| `npm run dev:web` | Vite only |
| `npm run dev:android` | USB phone workflow: attach to a running dev server or pick a free port, then `adb reverse` |
| `npm run build` / `npm run preview` | Production build / preview |
| `npm test` | Vitest (web) |
| `npm run check` | svelte-check + tsc |
| `npm run lint` | prettier --check + eslint |
| `npm run format` | prettier --write + cargo fmt |
| `npm run verify` | web check/test/lint + Rust fmt/clippy/test |
| `npm run rust:test` | `cargo test` for the Rust workspace |
| `npm run rust:fmt:check` | Rust formatting check |
| `npm run rust:clippy` | Rust lints |
| `npm run rust:build` | release build of the Rust binary |
| `npm run rust:serve` | run the Rust NDJSON server on `127.0.0.1:8080` |

Rust scripts require `cargo` on `PATH` (rustup, then a new shell or
`source "$HOME/.cargo/env"`); `rust-toolchain.toml` pins the stable toolchain.

Run `npm run verify` (the canonical check) before finishing. The app calls
`/v1` on its own origin; Vite proxies that to the development server on `:8080`. `serve` is
a development/CLI harness; the packaged app is Tauri and calls `core` directly
through the `CheckClient` seam.

## Conventions

- **Commits** — Conventional Commits: `type(scope): summary`. Scopes are `web`,
  `rust`, or `root`. Hooks are not pushed; only commit when asked.
- **Formatting** — Rust uses `cargo fmt`, run via `npm run format` or
  `npm run rust:fmt`. The web workspace uses **prettier** (configured with
  `useTabs: true`, single quotes, no trailing comma, printWidth 100 in
  `web/prettier.config.js`; the Tailwind plugin sorts classes). Husky's
  `pre-commit` hook runs lint-staged, which prettifies + eslint-fixes staged web
  files and cargo-fmts Rust files.
- **Provider HTTP** — status-only sources live in `rust/core/src/status.rs`: a
  request template plus a `CodeMap`. Body-parsing, multi-request, POST/GraphQL,
  and RDAP sources live in `rust/core/src/custom.rs`. DNS sources live in
  `rust/core/src/dns.rs`. Use the shared `Transport` seam; never call the HTTP
  client from a provider directly.
- **Status** — a `Check` maps a response to `available` / `taken` / `unknown`.
  `unknown` is never cached, so transient failures self-heal. Match the
  documented semantics in the README (domain rows report registry state, handle
  rows mean "the name exists").
- **Keyed sources** — intentionally omitted for now. Do not add a row, request,
  or cache entry for a credential-gated source. If they return, keep secrets in
  request headers only and out of results, logs, and tests.

## Adding a source

1. Add the metadata to `fixtures/sources.json`: `id`, `platform`,
   `platformName`, `nameTemplate`, `urlTemplate`, `paid`, `minLength`,
   `maxLength`, `keyed`, `tags`. A deferred/manual source stops here: with no
   provider it is listed as "not checked" automatically (`checked` is derived,
   never stored).
2. Implement the check in `rust/core`:
   - status-only → `status.rs` spec table;
   - body/multi-request/POST/RDAP → `custom.rs` spec table;
   - DNS → `dns.rs` host table.
3. Map the response to `available` / `taken` / `unknown`. Return `unknown` on an
   ambiguous or unparseable body, never a transport error.
4. Add a unit test or fixture. Status/custom providers are compared in
   `rust/core/tests/parity.rs`; DNS uses a fake resolver.
5. Add the source row to `README.md` and `docs/platforms.md`, then run
   `npm run verify` — the tree must stay green.
