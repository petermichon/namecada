# Namecheck

A name availability report. Enter a bare name (`acema`) and it checks that name
across several independent sources at once, all through one Rust core. Dots
aren't accepted — the sources are checked per label, so a dotted input
(`acema.com`) is rejected with a prompt to drop the dot rather than silently
reduced.

## How it works

Each source runs independently and streams its verdict into the report as soon
as it resolves, so one slow or failing source never blocks the others. Nothing
waits for the whole set.

The app is a static client of one server process: `GET /v1/check?name=<label>`
streams NDJSON events — a `manifest`, one `result` per source, then `done`. The
manifest lists every source (checked, keyed or deferred) with its URL, so the
page renders the full report before checks finish and fills rows in place.

Development uses Vite with `/v1` proxied to `serve`. Vite is pinned to port
5221 with `--strictPort`, so the desktop shell always attaches to this project's
dev server rather than whatever else is on a default port.

`core` has no HTTP dependency. `serve` is a development/CLI harness only; the
packaged app is Tauri (`src-tauri/`), calling the engine directly and feeding
the same events through the `CheckClient` seam, with no localhost server. The
web client picks its transport once at load: `invoke('check')` with an IPC
channel inside the desktop shell, the NDJSON endpoint everywhere else. The
source registry is compiled into the binary (`Manifest::embedded`);
`--manifest` overrides it for development.

Desktop app: `npm run tauri:dev` (Vite dev server plus the native window) and
`npm run tauri:build` (bundles the single-file web build).

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
| Snap Store | `api.snapcraft.io/v2/snaps/info/{name}` (`Snap-Device-Series: 16`) | 200 taken / 404 free  |
| Chef Supermarket | `supermarket.chef.io/api/v1/cookbooks/{name}`               | 200 taken / 404 free      |
| CocoaPods | `cocoapods.org/pods/{name}`                                   | 200 taken / 404 free      |
| CPAN      | `fastapi.metacpan.org/v1/module/{name}` (case-sensitive)      | 200 taken / 404 free      |
| Lobsters  | `lobste.rs/~{name}`                                           | 200 taken / 404 free      |
| Blogger   | `{name}.blogspot.com`                                         | 200 taken / 404 free      |
| Flickr    | `flickr.com/people/{name}/`                                   | 200 taken / 404 free      |
| Goodreads | `goodreads.com/{name}` (301 to profile)                       | 301/200 taken / 404 free  |
| StackBlitz | `stackblitz.com/@{name}` (HTML Accept)                       | 200 taken / 404 free      |
| Slack     | `{name}.slack.com`                                            | 200/403 taken / 404 free  |
| Firefox   | `addons.mozilla.org/api/v5/addons/addon/{name}/`               | 200/401 taken / 404 free  |
| JSR       | `jsr.io/api/scopes/{name}`                                     | 200/400 taken / 404 free  |
| WordPress | `api.wordpress.org/plugins/info/1.2/?request[slug]={name}`     | 200/closed taken / 404 free |
| Open VSX  | `open-vsx.org/api/{name}` (namespace)                          | 200 taken / 404 free      |
| GitHub    | `api.github.com/users/{name}`                                  | 200 taken / 404 free      |
| GitLab    | `gitlab.com/api/v4/users?username=` + `/groups/{name}`         | user or group = taken     |
| Bitbucket | `api.bitbucket.org/2.0/workspaces/{name}`                      | 200/403 taken / 404 free  |
| Codeberg  | `codeberg.org/api/v1/users/{name}`                             | 200 taken / 404 body      |
| SourceForge | `sourceforge.net/projects/{name}/`                           | 200 taken / 404 free      |
| Wikipedia | `meta.wikimedia.org/w/api.php` globaluserinfo               | `missing` = free          |
| DEV       | `dev.to/api/users/by_username` + `/organizations/{name}`       | 200 taken / 404 free      |
| Farcaster | `fnames.farcaster.xyz/transfers/current?name={name}`           | 200 taken / 404 free      |
| Minehut   | `api.minehut.com/server/{name}?byName=true`                    | `server: null` = free     |
| Lichess   | `lichess.org/api/user/{name}`                                  | 200 taken / 404 free      |
| Chess.com | `api.chess.com/pub/player/{name}`                              | 200 taken / 404 free      |
| Hugging Face | `huggingface.co/api/users/{name}/overview`                 | 200 taken / 404 free      |
| Dailymotion | `api.dailymotion.com/user/{name}`                            | 200 taken / 404 free      |
| Mixcloud  | `api.mixcloud.com/{name}/`                                     | 200 taken / 404 free      |
| SoundCloud | `soundcloud.com/{name}` (profile)                             | 200 taken / 404 free      |
| Bandcamp  | `{name}.bandcamp.com` (redirect target)                        | `/signup` redirect free / 200 or elsewhere taken |
| AniList   | `graphql.anilist.co` `User(name:)` (POST)                      | non-null `User` taken / null free |
| Roblox    | `users.roblox.com/v1/usernames/users` (POST)                   | non-empty `data` taken / empty free |
| Minecraft | `api.mojang.com/users/profiles/minecraft/{name}`              | 200 taken / 404 free      |
| Open Collective | `opencollective.com/{name}.json`                         | 200 taken / 404 free      |
| Liberapay | `liberapay.com/{name}`                                        | 200 taken / 404 free      |
| Bluesky   | `bsky.social/xrpc/com.atproto.temp.checkHandleAvailability` (`{name}.bsky.social`) | resultAvailable free / resultUnavailable taken |
| Mastodon  | `mastodon.social/@{name}` (profile)                            | 200/410 taken / 404 free  |
| YouTube   | `youtube.com/@{name}` (profile)                                | 200 taken / 404 free      |
| Kaggle    | `kaggle.com/{name}` (profile)                                  | 200 taken / 404 free      |
| skills.sh | `skills.sh/{name}` (publisher profile)                         | 200 taken / 404 free      |
| Dribbble  | `dribbble.com/{name}` (profile)                                | 200 taken / 404 free; a non-public profile 404s an "isn't public yet" body and is taken |
| Behance   | `behance.net/{name}` (profile)                                 | 200 taken / 404 free      |
| Replicate | `replicate.com/{name}` (profile)                               | 200 taken / 404 free      |
| DeviantArt | `deviantart.com/{name}` (profile)                             | 200 taken / 404 free      |
| Snapchat  | `snapchat.com/add/{name}` (profile)                            | 200 taken / 404 free      |
| Strava    | `strava.com/athletes/{name}` (profile)                         | 200 taken / 404 free      |
| Swift Package Index | `swiftpackageindex.com/{name}` (publisher)            | 200 taken / 404 free; publishing nothing also 404s |
| TikTok | `tiktok.com/oembed?url=…/@{name}` | 200 `author_url` taken / 400 free |
| Threads | `threads.net/@{name}` | `(@name)` in title taken / `Threads • Log in` free |
| Twitch | `twitch.tv/{name}` | `{name} - Twitch` taken / `Twitch` free |
| Duolingo | `duolingo.com/profile/{name}` | og:title `on Duolingo` taken / `Learn a language` free |
| Telegram | `t.me/{name}` | `Telegram: View @` taken / `Contact @` free |
| HackerNoon | `hackernoon.com/u/{name}` (profile)                           | 200 taken / 404 free      |
| Spotify   | `open.spotify.com/user/{name}` (profile)                       | 200 taken / 404 free      |
| Paragraph | `paragraph.com/@{name}` (profile)                              | 200 taken / 404 free      |
| Steam     | `steamcommunity.com/id/{name}?xml=1`                           | `<steamID64>` taken / `<error>` free |
| Instagram | `instagram.com/{name}/` (profile)                              | `"username":"{name}"` taken / generic shell free |
| Hacker News | `hacker-news.firebaseio.com/v0/user/{name}.json`             | `null` free / object with `id` taken |
| Giphy     | `giphy.com/{name}` (profile)                                   | 200 canonical `/{name}` taken / 301 → `/explore/{name}` free |
| Minds     | `api/v1/channel/{name}` (profile `minds.com/{name}/`)           | `status:"success"` taken / `ChannelNotFoundException` free |
| X / Twitter | `x.com/{name}` (profile)                                     | 200 taken / 404 free      |
| Vercel    | `{name}.vercel.app` (subdomain)                                | 200/401 (SSO) taken / 404 free |
| Netlify   | `{name}.netlify.app` (subdomain)                               | 200 taken / 404 free      |
| Surge     | `{name}.surge.sh` (subdomain)                                  | 200 taken / 404 free      |
| Firebase Hosting | `{name}.web.app` (subdomain)                            | 200 taken / 404 free      |
| Neocities | `{name}.neocities.org` (subdomain)                             | 200 taken / 404 free      |
| Render    | `{name}.onrender.com` (subdomain)                              | 200 taken / 404 free      |
| Railway   | `{name}.up.railway.app` (subdomain)                            | 200 taken / 404 free      |
| Heroku    | `{name}.herokuapp.com` (subdomain)                             | 200 taken / 404 free      |
| Read the Docs | `{name}.readthedocs.io` (subdomain)                        | 200 taken / 404 free      |
| Gumroad   | `{name}.gumroad.com` (subdomain)                               | 200 taken / 404 free      |
| Substack  | `{name}.substack.com` (subdomain)                              | 200 taken / 404 free      |
| Shopify   | `{name}.myshopify.com` (subdomain)                             | 200 taken / 404 free      |
| Weebly    | `{name}.weebly.com` (subdomain)                                | 200 taken / 404 free      |
| eu.org    | `{name}.eu.org` (DNS, no wildcard)                             | resolves taken / NXDOMAIN free |
| DuckDNS   | `{name}.duckdns.org` (DNS, no wildcard)                        | resolves taken / NXDOMAIN free |
| No-IP     | `{name}.no-ip.org` (DNS, no wildcard)                          | resolves taken / NXDOMAIN free |
| dynv6     | `{name}.dynv6.net` (DNS, no wildcard)                          | resolves taken / NXDOMAIN free |
| noted.lol | `{name}.noted.lol` (DNS, no wildcard)                          | resolves taken / NXDOMAIN free |
| is-a-good.dev | `{name}.is-a-good.dev` (DNS, no wildcard)                  | resolves taken / NXDOMAIN free |
| Cloudflare Pages | `{name}.pages.dev` (DNS, no wildcard)                    | resolves taken / NXDOMAIN free |
| Fly.io    | `{name}.fly.dev` (DNS, no wildcard)                            | resolves taken / NXDOMAIN free |

Credential-gated sources (Discogs, Companies House, OpenCorporates) are not
currently included.

Domains run once per TLD (`.com`, `.net`, `.org`, `.dev`, `.io`, `.ai`, `.app`), each a
separate provider. Each TLD uses its registry's RDAP endpoint; the live IANA
bootstrap lookup is not wired into the core yet. Domains are **paid**
sources: a registration costs money, and the amount varies (premium names cost
more). ENS
(`.eth`) is likewise paid — registering costs ETH. SNS (`.sol`) is paid too —
registering costs SOL. HNS is paid as well — a new Handshake name is won in an
auction, and an expired one is reclaimed the same way. Namecoin (`.bit`) is paid
too — registering a `d/` name costs NMC. Heroku is also paid (its free
tier ended in November 2022) but is still surfaced. Every handle, package and
social source is **free** to register. Each source carries a `paid` flag; the
report marks paid rows and offers a cost filter, but otherwise renders one
alphabetical list.

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

## Rust core

`rust/` is the engine plus a development/CLI harness. It is a Rust workspace
assembled from a small core plus two binaries:

- `core/src/provider.rs` — the `Provider` and `Transport` contracts and the
  result/status types.
- `core/src/adapters/` — one module per source shape: `status.rs` (verdict is
  the status code), `api.rs` (JSON/XML/GraphQL endpoints), `page.rs`
  (server-rendered HTML), `domains.rs` (RDAP), `dns.rs` (wildcard subdomains).
  `shared.rs` holds the request helpers and `spec.rs` the id → check table.
- `core/src/registry.rs` — the compiled registry: live providers and stats.
- `core/src/engine.rs` — bounded-concurrency scan engine with its one-hour
  verdict cache; streams each result as it resolves.
- `core/src/http.rs` — production HTTP transport.
- `cli/src/bin/serve.rs` — the NDJSON HTTP server; `capture.rs` records fixtures.

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

Fixtures are a **capture**, not live truth: responses change. They were produced
by the pre-Rust exporter; the Rust-only tree has no exporter yet, so treat them
as a frozen oracle and regenerate from git history if a source changes.

## Caching

`serve` caches results per provider, keyed `provider:label` (e.g. `lemy:acema`),
with a 1-hour TTL. Only definitive verdicts are cached — `unknown` is never
stored, so transient rate-limit or network errors self-heal. A fresh entry is
served without touching the network.

This matters most for sources like GitHub and npm, which rate-limit
unauthenticated requests (GitHub: 60/hour per IP).

## Layout

- `rust/` — Rust engine and the development/CLI harness; see `rust/README.md`.
- `web/` — Svelte 5 + Vite + Tailwind + shadcn-svelte (maia) app.
  - `src/lib/platforms.ts` — the result types (`Status`, `ProviderResult`),
    `toLabel`, and `suggestLabel`. The display name in `platformName` and the
    `paid` flag come from the manifest stream; the report renders one
    alphabetical list with a cost filter and marks paid rows.
  - `src/lib/api.ts` — the event contract and the `CheckClient` transport seam.
    `streamFromApi` (the `httpCheckClient`) consumes the `serve` API
    (`GET /v1/check?name=<label>`, NDJSON events); a native shell can inject its
    own client without touching the UI or lookup state.
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

`npm run dev:host` is the proxied mode. It starts the same two processes but binds the web app to
`127.0.0.1:5220`, the upstream the machine-level dev proxy (`~/git/devhost`) maps to
`http://namecheck.localhost:7357`. The phone reaches the same URL through one
`adb reverse tcp:7357`; the Rust `serve` on `:8080` stays Vite's proxy target as before.

## Adding a source

1. Add the source metadata to `fixtures/sources.json`: `id`, `platform`,
   `platformName`, `nameTemplate`, `urlTemplate`, `paid`, `minLength`,
   `maxLength`, and `keyed`. A source with no provider is listed as
   "not checked" automatically — `checked` is derived from the provider
   registry, never stored. The registry is compiled into the binary, so rebuild
   after editing.
2. Implement the check in `rust/core`:
   - a status-only source goes in `rust/core/src/adapters/status.rs`;
   - a JSON/XML/GraphQL source goes in `adapters/api.rs`;
   - a server-rendered HTML source goes in `adapters/page.rs`;
   - an RDAP or DNS source goes in `adapters/domains.rs` / `adapters/dns.rs`.
3. Register it in the corresponding spec table.
4. Add a unit test or capture fixture. Status-only and custom sources are
   exercised by `rust/core/tests/parity.rs`; DNS providers use a fake resolver.
5. Run `npm run verify`.

Checks run in the Rust core, so CORS is not a constraint. Keyed sources are
deliberately omitted for now.

Deferred: Packagist, Docker Hub, Go proxy (modules are paths, not single
labels), and the remaining trademark / social / company registries that need a
key.
