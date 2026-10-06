# Domain Browser

A name availability report. Enter a name — bare (`acema`) or a full domain
(`acema.com`, `www.bbc.co.uk`) — and it checks that name across several
independent sources at once, fetched directly from the browser.

## How it works

The input is reduced to its **label**: `acema.com` → `acema`, `www.bbc.co.uk` →
`bbc`. Each source then runs independently and streams its verdict into the
report as soon as it resolves, so one slow or failing source never blocks the
others. Nothing waits for the whole set.

Every source is **browser-direct**: it must send `Access-Control-Allow-Origin`,
so there is no backend or proxy. Anything that can't be reached that way is
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
| Hacker News | `hacker-news.firebaseio.com/v0/user/{name}.json`             | null = free               |
| Minehut   | `api.minehut.com/server/{name}?byName=true`                    | `server: null` = free     |
| Lichess   | `lichess.org/api/user/{name}`                                  | 200 taken / 404 free      |
| Chess.com | `api.chess.com/pub/player/{name}`                              | 200 taken / 404 free      |
| Hugging Face | `huggingface.co/api/users/{name}/overview`                 | 200 taken / 404 free      |
| Dailymotion | `api.dailymotion.com/user/{name}`                            | 200 taken / 404 free      |
| Mixcloud  | `api.mixcloud.com/{name}/`                                     | 200 taken / 404 free      |
| Open Collective | `opencollective.com/{name}.json`                         | 200 taken / 404 free      |
| Bluesky   | `public.api.bsky.app` resolveHandle (`{name}.bsky.social`)     | 200 taken / 400 free      |
| Mastodon  | `mastodon.social/api/v1/accounts/lookup?acct={name}`           | 200 taken / 404 free      |

Domains run once per TLD in `TLDS` (`.com`, `.org`, `.dev`, `.io`, `.ai`, `.app`), each a
separate provider. The TLD is resolved to its registry through the IANA RDAP
bootstrap (`data.iana.org/rdap/dns.json`, cached for 7 days).

GitHub users and organizations share a single namespace, so one `/users` lookup
covers both.

### Semantics

- Domain rows report **registry state**, not registrability: a name may be
  reserved or premium and still show as free. Coverage follows the bootstrap.
- Dev/social rows mean "the name exists", not legal or brand availability.
- Some platforms enforce a minimum name length (Chess.com 3, Lichess 2); shorter
  names are shown as **unavailable** (not registrable) rather than free.

## Caching

Results are cached per provider in `localStorage`, keyed `provider:label`
(e.g. `domain.com:acema`), with a 1-hour TTL. Only definitive verdicts are
cached — `unknown` is never stored, so transient rate-limit or network errors
self-heal. A fresh entry is served without touching the network.

This matters most for GitHub and npm, which rate-limit unauthenticated requests
(GitHub: 60/hour per IP).

## Layout

- `frontend/` — Svelte 5 + Vite + Tailwind + shadcn-svelte (maia) app.
  - `src/lib/providers.ts` — `Provider` interface, the provider registry,
    `streamAll`/`checkAll`.
  - `src/lib/providerCache.ts` — per-provider response cache.
  - `src/lib/rdap.ts`, `src/lib/bootstrap.ts` — domain lookup.
- `scripts/domscan.py` — offline bulk scanner for `.com` names via Verisign RDAP
  (offline research tooling; not used by the app).
- `data/full.tsv` — legacy bulk snapshot from that scanner.

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

## Adding a source

Implement the `Provider` interface in `src/lib/providers.ts` (a `check` that maps
an HTTP response to `available` / `taken` / `unknown`) and add it to the
`providers` array. Confirm the source sends `Access-Control-Allow-Origin` on
**both** 200 and 404 first — some send it only on success, which makes the
verdict unreadable in the browser.

Deferred (not browser-reachable): RubyGems (no CORS on 404), Packagist (no CORS),
Reddit (403 from server IPs), Docker Hub (no CORS), HN Algolia (no availability
signal), Go proxy (modules are paths, not single labels), and trademark / social
/ company registries (need a key and a proxy).

## Bulk scanning (offline)

`scripts/domscan.py` brute-forces the 5-letter space over a chosen alphabet and
appends results to a TSV, resumable and rate-limited:

```sh
python3 scripts/domscan.py --letters aceimnorsuvwxz --workers 8 --rate 30 --out data/full.tsv
```

Respect Verisign's RDAP rate limits — lower `--rate` if you see sustained `429`s.
