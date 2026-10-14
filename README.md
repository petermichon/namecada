# Domain Browser

A name availability report. Enter a bare name (`acema`) and it checks that name
across several independent sources at once, all through a Go server. Dots
aren't accepted — the sources are checked per label, so a dotted input
(`acema.com`) is rejected with a prompt to drop the dot rather than silently
reduced.

## How it works

Each source runs independently and streams its verdict into the report as soon
as it resolves, so one slow or failing source never blocks the others. Nothing
waits for the whole set.

The app is a thin client of the server: it streams results from
`GET /check?name=<label>` over NDJSON. Every source lives in the server, which
is where sources that need a proxy, a key, or no CORS belong. Anything that
can't be reached is deferred (see below).

### Sources

| Source    | Request                                                        | Verdict                   |
| --------- | -------------------------------------------------------------- | ------------------------- |
| Domains   | RDAP registry via IANA bootstrap                               | 200 registered / 404 free |
| ENS       | `api.ensideas.com/ens/resolve/{name}.eth`                     | non-null `address` taken / `null` free |
| SNS       | `sdk-proxy.sns.id/resolve/{name}.sol`                          | `"s":"ok"` taken / `"s":"error"` free |
| HNS       | `hnsscan.com/api/names/{name}/details`                         | `info.state` `CLOSED` taken / `AVAILABLE`/`UNKNOWN` free |
| Namecoin  | `explorer.namecoin.co/api/name/d%2F{name}`                     | 404 free / 200 `expired:true` free, `false` taken |
| Lens      | `api.lens.xyz/graphql` `account(username.localName:)` (POST)   | non-null `account` taken / null free |
| npm       | `registry.npmjs.org/{name}`                                    | 200 taken / 404 free      |
| PyPI      | `pypi.org/pypi/{name}/json`                                    | 200 taken / 404 free      |
| crates.io | `crates.io/api/v1/crates/{name}`                               | 200 taken / 404 free      |
| NuGet     | `api.nuget.org/v3-flatcontainer/{name}/index.json`             | 200 taken / 404 free      |
| RubyGems  | `rubygems.org/api/v1/gems/{name}.json`                         | 200 taken / 404 free      |
| Hex       | `hex.pm/api/packages/{name}`                                   | 200 taken / 404 free      |
| Hex user  | `hex.pm/api/users/{name}`                                      | 200 taken / 404 free      |
| pub.dev   | `pub.dev/api/packages/{name}`                                  | 200 taken / 404 free      |
| Homebrew  | `formulae.brew.sh/api/formula/{name}.json`                     | 200 taken / 404 free      |
| Hackage   | `hackage.haskell.org/package/{name}`                           | 200 taken / 404 free      |
| Clojars   | `clojars.org/api/artifacts/{name}`                             | 200 taken / 404 free      |
| Chocolatey | `community.chocolatey.org/packages/{name}`                    | 200 taken / 404 free      |
| AUR       | `aur.archlinux.org/packages/{name}`                            | 200 taken / 404 free      |
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
| Last.fm   | `last.fm/user/{name}` (profile)                                | 200 taken / 404 free      |
| SoundCloud | `soundcloud.com/{name}` (profile)                             | 200 taken / 404 free      |
| AniList   | `graphql.anilist.co` `User(name:)` (POST)                      | non-null `User` taken / null free |
| Open Collective | `opencollective.com/{name}.json`                         | 200 taken / 404 free      |
| Bluesky   | `bsky.social/xrpc/com.atproto.temp.checkHandleAvailability` (`{name}.bsky.social`) | resultAvailable free / resultUnavailable taken |
| Mastodon  | `mastodon.social/@{name}` (profile)                            | 200/410 taken / 404 free  |
| YouTube   | `youtube.com/@{name}` (profile)                                | 200 taken / 404 free      |
| Kaggle    | `kaggle.com/{name}` (profile)                                  | 200 taken / 404 free      |
| Dribbble  | `dribbble.com/{name}` (profile)                                | 200 taken / 404 free      |
| Behance   | `behance.net/{name}` (profile)                                 | 200 taken / 404 free      |
| Replicate | `replicate.com/{name}` (profile)                               | 200 taken / 404 free      |
| HackerNoon | `hackernoon.com/u/{name}` (profile)                           | 200 taken / 404 free      |
| Spotify   | `open.spotify.com/user/{name}` (profile)                       | 200 taken / 404 free      |
| Paragraph | `paragraph.com/@{name}` (profile)                              | 200 taken / 404 free      |
| Instagram | `instagram.com/{name}/` (profile)                              | `"username":"{name}"` taken / generic shell free |
| Hacker News | `news.ycombinator.com/user?id={name}`                        | `No such user.` free / `karma` taken |
| Discogs   | `api.discogs.com/users/{name}` (token)                         | 200 taken / 404 free      |
| X / Twitter | `x.com/{name}` (profile)                                     | 200 taken / 404 free      |
| Twitch    | `api.twitch.tv/helix/users?login={name}` (app token)           | non-empty `data` taken / empty free |
| Companies House | `api.company-information.service.gov.uk/search/companies?q={name}` (key) | exact company name taken / else free |
| OpenCorporates | `api.opencorporates.com/v0.4/companies/search?q={name}` (token) | exact company name taken / else free |
| Vercel    | `{name}.vercel.app` (subdomain)                                | 200 taken / 404 free      |
| Netlify   | `{name}.netlify.app` (subdomain)                               | 200 taken / 404 free      |
| Surge     | `{name}.surge.sh` (subdomain)                                  | 200 taken / 404 free      |
| Firebase Hosting | `{name}.web.app` (subdomain)                            | 200 taken / 404 free      |
| Neocities | `{name}.neocities.org` (subdomain)                             | 200 taken / 404 free      |
| Render    | `{name}.onrender.com` (subdomain)                              | 200 taken / 404 free      |
| Railway   | `{name}.up.railway.app` (subdomain)                            | 200 taken / 404 free      |
| Heroku    | `{name}.herokuapp.com` (subdomain)                             | 200 taken / 404 free      |
| Deno Deploy | `{name}.deno.dev` (subdomain)                                | 200 taken / 404 free      |
| Read the Docs | `{name}.readthedocs.io` (subdomain)                        | 200 taken / 404 free      |
| Gumroad   | `{name}.gumroad.com` (subdomain)                               | 200 taken / 404 free      |
| Substack  | `{name}.substack.com` (subdomain)                              | 200 taken / 404 free      |
| eu.org    | `{name}.eu.org` (DNS, no wildcard)                             | resolves taken / NXDOMAIN free |
| DuckDNS   | `{name}.duckdns.org` (DNS, no wildcard)                        | resolves taken / NXDOMAIN free |
| No-IP     | `{name}.no-ip.org` (DNS, no wildcard)                          | resolves taken / NXDOMAIN free |
| dynv6     | `{name}.dynv6.net` (DNS, no wildcard)                          | resolves taken / NXDOMAIN free |
| noted.lol | `{name}.noted.lol` (DNS, no wildcard)                          | resolves taken / NXDOMAIN free |
| is-a-good.dev | `{name}.is-a-good.dev` (DNS, no wildcard)                  | resolves taken / NXDOMAIN free |
| Cloudflare Pages | `{name}.pages.dev` (DNS, no wildcard)                    | resolves taken / NXDOMAIN free |
| Fly.io    | `{name}.fly.dev` (DNS, no wildcard)                            | resolves taken / NXDOMAIN free |

Domains run once per TLD (`.com`, `.org`, `.dev`, `.io`, `.ai`, `.app`), each a
separate provider. The TLD is resolved to its registry through the IANA RDAP
bootstrap (`data.iana.org/rdap/dns.json`). Domains are **paid** sources: a
registration costs money, and the amount varies (premium names cost more). ENS
(`.eth`) is likewise paid — registering costs ETH. SNS (`.sol`) is paid too —
registering costs SOL. HNS is paid as well — a new Handshake name is won in an
auction, and an expired one is reclaimed the same way. Namecoin (`.bit`) is paid
too — registering a `d/` name costs NMC. Heroku is also paid (its free
tier ended in November 2022) but is still surfaced. Every handle, package and
social source is **free** to register. The
report renders a Paid card above a Free card, and each provider carries a `paid`
flag the server streams on every `/check` result.

The free-subdomain sources (eu.org, DuckDNS, No-IP, dynv6, noted.lol,
is-a-good.dev, Cloudflare Pages, Fly.io) have no wildcard DNS, so they resolve `{name}.{host}` and treat a
successful answer as taken and an NXDOMAIN as free. Because that is a DNS
lookup, the label is sent to the resolver — the same privacy exposure as
looking the name up directly.

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

## Server

`server/` is the optional Go module that runs the checks server-side. It is a
small core assembled from four pluggable packages:

- `internal/provider` — the `Provider` interface (`ID`, `Platform`, `Name`,
  `URL`, `Check`) plus an ordered registry; each source is a provider.
- `internal/engine` — runs the registered providers concurrently and streams a
  result as each resolves.
- `internal/store` — a cache of verdicts keyed `provider:label` with a one-hour
  TTL.
- `internal/sink` — durable destinations for a finished scan. The SQLite sink
  upserts one row per `provider:label` into a `checks` table (primary key `name`
  + `provider`, so a re-check updates in place) and exposes an `available` view
  over the free names.

Two commands drive it (`go run ./cmd/domainbrowser ...`): `serve` exposes the
engine over HTTP for the app, streaming one NDJSON object per source as it
resolves, and `check <label> [--out <file.sqlite>]` runs a one-shot scan and
prints the whole result set as a single JSON array — or, with `--out`, writes it
to a SQLite file. Unlike `serve`, `check` buffers every source before emitting.

### Keyed sources

Some sources need an API key or token. Supply those through
`DOMAINBROWSER_<SOURCE>_<KIND>` environment variables; they are read once when
the process starts:

| Source | Environment variables |
| --- | --- |
| Discogs | `DOMAINBROWSER_DISCOGS_TOKEN` |
| Twitch | `DOMAINBROWSER_TWITCH_CLIENT_ID`, `DOMAINBROWSER_TWITCH_CLIENT_SECRET` |
| Companies House | `DOMAINBROWSER_COMPANIESHOUSE_KEY` |
| OpenCorporates | `DOMAINBROWSER_OPENCORPORATES_TOKEN` |

A source whose credentials are unset is **omitted** from the scan entirely — it
produces no row, makes no request, and never enters the cache. It is not
reported as `unknown`, which keeps that status meaning "checked, inconclusive".
Startup logs list the enabled keyed sources by name only; credential values are
never logged and never appear in a result.

The live source is always the source of truth. SQLite — like the in-memory store
— is a **cache** of what a source last reported, not an authoritative record: a
row can be stale, rate-limited, or wrong, and a domain row captures registry
state rather than registrability, so re-check the source rather than trusting
the file. The server is the sole source for the app: it lifts the browser's CORS
limit, keeps one shared cache in front of rate-limited APIs, and is where keyed
or no-CORS sources can be added.

## Caching

Results are cached per provider in the server's in-memory store, keyed
`provider:label` (e.g. `lemy:acema`), with a 1-hour TTL. Only definitive
verdicts are cached — `unknown` is never stored, so transient rate-limit or
network errors self-heal. A fresh entry is served without touching the network.

This matters most for sources like GitHub and npm, which rate-limit
unauthenticated requests (GitHub: 60/hour per IP).

## Layout

- `server/` — Go core (provider/engine/store/sink), see [Server](#server).
- `web/` — Svelte 5 + Vite + Tailwind + shadcn-svelte (maia) app.
  - `src/lib/platforms.ts` — the result types (`Status`, `ProviderResult`) and
    `toLabel`. The display name in `platformName` and the `paid` flag come from
    the server; the report renders the URLs as a Paid card above a Free card,
    each with its own heading and available count.
  - `src/lib/api.ts` — the server client: `streamFromApi` consumes the `serve`
    API (`GET /check?name=<label>`, NDJSON) and is the app's sole source. The
    base URL defaults to `http://localhost:8080` and is overridable at build
    time with `VITE_API_BASE`.

## Development

```sh
npm install        # installs the web workspace + git hooks
npm run dev        # vite dev server
npm run build      # single-file production build
npm test           # vitest
npm run check      # svelte-check + tsc
npm run lint       # prettier --check + eslint
npm run format     # prettier --write + gofmt
npm run verify     # check + test + lint (web and server)
npm run server:watch  # air: rebuild + restart the Go server on save
```

A husky `pre-commit` hook (installed by `npm install` via the `prepare`
script) runs lint-staged, which formats and lints staged web files
(prettier + eslint) and formats staged Go files (gofmt).

The app needs the server running (`cd server && go run ./cmd/domainbrowser
serve`); it streams every result from it. The server's CORS origin defaults to
the Vite dev origin (`http://localhost:5173`).

For the same edit-refresh loop as Vite's HMR, run `npm run server:watch`
instead. It drives [air](https://github.com/air-verse/air) (install with
`go install github.com/air-verse/air@latest`) using the root `.air.toml`, which
watches `server/` and, on a Go change, rebuilds and restarts `serve` on
`:8080`.

## Adding a source

Implement the `Provider` interface in the server
(`server/internal/provider`) — a `Check` that maps an HTTP response to
`available` / `taken` / `unknown` — and register it in `provider.Default(cfg)`.
A source that needs a credential goes in the `keyedSources` table with an
`enabled` predicate instead, so `Default` omits it when the credential is
unset (see [Keyed sources](#keyed-sources)). Sources live server-side, so CORS
is no longer a constraint.

Deferred: Packagist, Docker Hub, Go proxy (modules are paths, not
single labels), and the remaining trademark / social / company registries that
need a key.
