# AGENTS.md

Guidance for agents working in this repo. See `README.md` for architecture and
`docs/platforms.md` for the source catalogue; this file covers commands and
conventions.

## Layout

- `web/` — Svelte 5 + Vite + Tailwind + shadcn-svelte app. A thin client of the
  server (streams `GET /check?name=<label>` over NDJSON).
- `fixtures/` — captured source inventory plus raw HTTP snapshots for engine
  parity. Regenerate with `npm run server:export -- --label <name>`; see
  `fixtures/README.md`.
- `rust/` — parity-first Rust workspace (`core` + `cli`) for the engine rewrite;
  see `rust/README.md`.
- `server/` — Go module (`server/go.mod`) holding the checks. Four packages:
  `internal/provider` (the `Provider` interface + registry), `internal/engine`
  (concurrent streaming), `internal/store` (in-memory TTL cache), `internal/sink`
  (SQLite). Commands under `server/cmd/domainbrowser`: `serve` and `check`.

## Commands

Install once (also wires git hooks): `npm install`.

Root scripts (run from the repo root):

| Command | Does |
| --- | --- |
| `npm run dev` | Vite dev server **and** the Go server, concurrently |
| `npm run dev:web` | Vite only |
| `npm run build` / `npm run preview` | Production build / preview |
| `npm test` | Vitest (web) |
| `npm run check` | svelte-check + tsc |
| `npm run lint` | prettier --check + eslint |
| `npm run format` | prettier --write + gofmt |
| `npm run verify` | `check` + `test` + `lint` + `server:lint` + `server:test` |

Server scripts (all shell out with `go -C server`):

| Command | Does |
| --- | --- |
| `npm run server:dev` | `go run ./cmd/domainbrowser serve` |
| `npm run server:watch` | air rebuild/restart on save (see `.air.toml`) |
| `npm run server:build` | build `server/bin/domainbrowser` |
| `npm run server:test` | `go test ./...` |
| `npm run server:lint` | `go vet ./...` |
| `npm run server:fmt` | `gofmt -w server` |
| `npm run server:tidy` | `go mod tidy` |
| `npm run server:export` | write `fixtures/sources.json` (add `-- --label <name>` to capture a label) |
| `npm run rust:test` | `cargo test` for the Rust workspace |
| `npm run rust:fmt:check` | Rust formatting check |
| `npm run rust:clippy` | Rust lints |
| `npm run rust:serve` | run the Rust NDJSON server on `127.0.0.1:8080` |

Run `npm run verify` (this is the canonical check; it also runs `go vet` and
`go test`) before finishing. The app expects the server on `:8080`; its CORS
origin defaults to the Vite origin `http://localhost:5173`, and any loopback
origin (`http://localhost:<port>`, `http://127.0.0.1:<port>`) is echoed back so
Vite can fall back to another port without restarting the server.

## Conventions

- **Commits** — Conventional Commits: `type(scope): summary`. Scopes are
  `web`, `server`, or `root` (e.g. `feat(server): add air live reload`). Hooks
  are not pushed; only commit when asked.
- **Formatting** — Go is formatted with `gofmt` (**tabs**), run via
  `npm run server:fmt`. The web workspace is formatted with **prettier**
  (configured with `useTabs: true`, single quotes, no trailing comma,
  printWidth 100 in `web/prettier.config.js`; the Tailwind plugin sorts
  classes). Husky's `pre-commit` hook runs lint-staged, which prettifies +
  eslint-fixes staged web files and gofmt-formats staged Go files.
- **Provider HTTP** — status-only registry sources reuse the `endpoint` helper
  (`server/internal/provider/registries.go`): a small struct that GETs
  `baseURL + path` and maps the HTTP status through a `codeMap`. It mirrors the
  old frontend `checkHttp` helper and covers every source that needs no
  response-body parsing. Sources that must inspect the body (e.g. GitHub,
  Bluesky, Wikipedia, Minehut) embed `base` (`provider.go`) and implement
  `Check` by hand, using `doGet`/`checkHTTP` from `http.go`.
- **Status** — a `Check` maps a response to `available` / `taken` / `unknown`.
  `unknown` is never cached, so transient failures self-heal. Match the
  documented semantics in the README (domain rows report registry state, handle
  rows mean "the name exists").
- **Keyed sources** — credentials live in `provider.Config`, read once in
  `cmd/domainbrowser` from `DOMAINBROWSER_<SOURCE>_<KIND>` env vars (the list is
  in the README). `provider.Default(cfg)` registers a keyed source only when its
  credentials are present; an absent credential **omits** the source entirely
  (never a row, request, or cache entry — not `unknown`). Secrets go in request
  headers only: never in a returned `Result.URL`/`Detail`, a log, or a test.
  Startup logs list enabled keyed sources by name only.

## Adding a source

1. Implement the `Provider` interface in `server/internal/provider`. Prefer
   `newEndpoint(id, platform, baseURL, path, pageURL, codes)` when the verdict is
   just an HTTP status; embed `base` and write a `Check` when it needs to parse
   the body. Add a doc comment noting the API path and the response codes.
2. Map the response to `available` / `taken` / `unknown`. Declare any documented
   minimum name length in the `minLengths` table in `provider.go` (or implement
   `MinLengthProvider`).
3. Register it in `provider.Default(cfg)` in `provider.go` (append to the
   `keyedSources` table when it needs a credential), and add the source row to
   the table in `README.md` (and `docs/platforms.md`).
4. Add tests under `server/internal/provider` covering available, taken, and
   unknown. Run `npm run verify` — the tree must stay green.
