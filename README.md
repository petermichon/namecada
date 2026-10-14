# Domain Browser

A name availability report. Enter a bare name (`acema`) and it checks that name
across several independent sources at once. A Go backend does the checking when
it is reachable; otherwise the app falls back to fetching the sources directly
from the browser. Dots aren't accepted — the sources are checked per label, so a
dotted input (`acema.com`) is rejected with a prompt to drop the dot rather than
silently reduced.

## How it works

Each source runs independently and streams its verdict into the report as soon
as it resolves, so one slow or failing source never blocks the others. Nothing
waits for the whole set.

The app prefers the backend: when its API is reachable it streams results from
`GET /check?name=<label>` over NDJSON, and only falls back to checking each
source directly from the browser when no backend is reachable. The backend is
where sources that need a proxy, a key, or no CORS are meant to live;
browser-direct is limited to sources that send `Access-Control-Allow-Origin` on
both the taken and free response. Anything that can't be reached either way is
deferred (see below).

### Sources

| Source    | Request                                                        | Verdict                   |
| --------- | -------------------------------------------------------------- | ------------------------- |
| Domains   | RDAP registry via IANA bootstrap                               | 200 registered / 404 free |
| npm       | `registry.npmjs.org/{name}`                                    | 200 taken / 404 free      |
| PyPI      | `pypi.org/pypi/{name}/json`                                    | 200 taken / 404 free      |
| crates.io | `crates.io/api/v1/crates/{name}`                               | 200 taken / 404 free      |
| NuGet     | `api.nuget.org/v3-flatcontainer/{name}/index.json`             | 200 taken / 404 free      |
| Hex       | `hex.pm/api/packages/{name}`                                   | 200 taken / 404 free      |
| Hex user  | `hex.pm/api/users/{name}`                                      | 200 taken / 404 free      |
| pub.dev   | `pub.dev/api/packages/{name}`                                  | 200 taken / 404 free      |
| Homebrew  | `formulae.brew.sh/api/formula/{name}.json`                     | 200 taken / 404 free      |
| Firefox   | `addons.mozilla.org/api/v5/addons/addon/{name}/`               | 200/401 taken / 404 free  |
| JSR       | `jsr.io/api/scopes/{name}`                                     | 200/400 taken / 404 free  |
| WordPress | `api.wordpress.org/plugins/info/1.2/?request[slug]={name}`     | 200/closed taken / 404 free |
| Open VSX  | `open-vsx.org/api/{name}` (namespace)                          | 200 taken / 404 free      |
| GitHub    | `api.github.com/users/{name}`                                  | 200 taken / 404 free      |
| GitLab    | `gitlab.com/api/v4/users?username=` + `/groups/{name}`         | user or group = taken     |
| Bitbucket | `api.bitbucket.org/2.0/workspaces/{name}`                      | 200/403 taken / 404 free  |
| Codeberg  | `codeberg.org/api/v1/users/{name}`                             | 200 taken / 404 body      |
| Wikipedia | `meta.wikimedia.org/w/api.php` globaluserinfo               | `missing` = free          |
| DEV       | `dev.to/api/users/by_username` + `/organizations/{name}`       | 200 taken / 404 free      |
| Farcaster | `fnames.farcaster.xyz/transfers/current?name={name}`           | 200 taken / 404 free      |
| Minehut   | `api.minehut.com/server/{name}?byName=true`                    | `server: null` = free     |
| Lichess   | `lichess.org/api/user/{name}`                                  | 200 taken / 404 free      |
| Chess.com | `api.chess.com/pub/player/{name}`                              | 200 taken / 404 free      |
| Hugging Face | `huggingface.co/api/users/{name}/overview`                 | 200 taken / 404 free      |
| Dailymotion | `api.dailymotion.com/user/{name}`                            | 200 taken / 404 free      |
| Mixcloud  | `api.mixcloud.com/{name}/`                                     | 200 taken / 404 free      |
| Open Collective | `opencollective.com/{name}.json`                         | 200 taken / 404 free      |
| Bluesky   | `bsky.social/xrpc/com.atproto.temp.checkHandleAvailability` (`{name}.bsky.social`) | resultAvailable free / resultUnavailable taken |
| Mastodon  | `mastodon.social/@{name}` (profile)                            | 200/410 taken / 404 free  |

Domains run once per TLD in `TLDS` (`.com`, `.org`, `.dev`, `.io`, `.ai`, `.app`), each a
separate provider. The TLD is resolved to its registry through the IANA RDAP
bootstrap (`data.iana.org/rdap/dns.json`, cached for 7 days).

GitHub users and organizations share a single namespace, so one `/users` lookup
covers both.

### Semantics

- Domain rows report **registry state**, not registrability: a name may be
  reserved or premium and still show as free. Coverage follows the bootstrap.
  The RFC 6761 special-use labels (`example`, `test`, `invalid`, `localhost`)
  are always shown as unavailable, since some registries reserve them without
  creating an RDAP record.
- Handle and package rows mean "the name exists", not legal or brand
  availability.
- Some platforms enforce a minimum name length (Chess.com 3, Lichess 2); shorter
  names are shown as **unavailable** (not registrable) rather than free.

## Backend

`backend/` is the optional Go module that runs the checks server-side. It is a
small core assembled from four pluggable packages:

- `internal/provider` — the `Provider` interface (`ID`, `Platform`, `Name`,
  `URL`, `Check`) plus an ordered registry; each source is a provider.
- `internal/engine` — runs the registered providers concurrently and streams a
  result as each resolves, mirroring the frontend `streamAll`.
- `internal/store` — a cache of verdicts keyed `provider:label` with a one-hour
  TTL, mirroring the frontend `providerCache`.
- `internal/sink` — durable destinations for a finished scan. The SQLite sink
  upserts one row per `provider:label` into a `checks` table (primary key `name`
  + `provider`, so a re-check updates in place) and exposes an `available` view
  over the free names.

Two commands drive it (`go run ./cmd/domainbrowser ...`): `serve` exposes the
engine over HTTP for the app, and `check <label> [--out <file.sqlite>]` prints
results as JSON or writes them to a SQLite file.

The live source is always the source of truth. SQLite — like the in-memory store
— is a **cache** of what a source last reported, not an authoritative record: a
row can be stale, rate-limited, or wrong, and a domain row captures registry
state rather than registrability, so re-check the source rather than trusting
the file. The backend is preferred over browser-direct because it lifts the
browser's CORS limit, keeps one shared cache in front of rate-limited APIs, and
is where keyed or no-CORS sources can be added; the browser path remains the
fallback so the app still works with no backend running.

## Caching

Results are cached per provider in `localStorage`, keyed `provider:label`
(e.g. `domain.com:acema`), with a 1-hour TTL. Only definitive verdicts are
cached — `unknown` is never stored, so transient rate-limit or network errors
self-heal. A fresh entry is served without touching the network.

This matters most for GitHub and npm, which rate-limit unauthenticated requests
(GitHub: 60/hour per IP).

## Layout

- `backend/` — Go core (provider/engine/store/sink), see [Backend](#backend).
- `frontend/` — Svelte 5 + Vite + Tailwind + shadcn-svelte (maia) app.
  - `src/lib/providers.ts` — `Provider` interface, the provider registry,
    `streamAll`/`checkAll`. Each provider declares a `platform`; the report
    groups the URLs into one card per platform.
  - `src/lib/api.ts` — backend client: `streamResults` prefers the `serve` API
    (`GET /check?name=<label>`, NDJSON) and falls back to `streamAll`
    browser-direct when it is unreachable. The base URL defaults to
    `http://localhost:8080` and is overridable at build time with
    `VITE_API_BASE`.
  - `src/lib/providerCache.ts` — per-provider response cache.
  - `src/lib/rdap.ts`, `src/lib/bootstrap.ts` — domain lookup.

## Development

```sh
npm install        # installs the frontend workspace
npm run dev        # vite dev server
npm run build      # single-file production build
npm test           # vitest
npm run check      # svelte-check + tsc
npm run lint       # prettier --check + eslint
npm run format     # prettier --write
```

With a backend running (`cd backend && go run ./cmd/domainbrowser serve`), the
app streams results from it; without one it checks the sources directly. The
backend's CORS origin defaults to the Vite dev origin (`http://localhost:5173`).

## Adding a source

Implement the `Provider` interface in `src/lib/providers.ts` (a `check` that maps
an HTTP response to `available` / `taken` / `unknown`) and add it to the
`providers` array. Confirm the source sends `Access-Control-Allow-Origin` on
**both** 200 and 404 first — some send it only on success, which makes the
verdict unreadable in the browser.

Deferred (not browser-reachable): RubyGems (no CORS on 404), Packagist (no CORS),
Reddit (403 from server IPs), Docker Hub (no CORS), Go proxy (modules are paths,
not single labels), Hacker News (its API only exposes users with public activity,
so it can never confirm a name is free; the profile page that can has no CORS),
and trademark / social / company registries (need a key and a proxy).
