# domain-browser

Browse and search every 5-letter `.com` domain made from the letters `aceimnorsuvwxz`
(14 letters → 14⁵ = 537,824 names), with availability status.

## Layout

- `data/full.tsv` — the scanned dataset: `name<TAB>status`, where status is
  `available`, `registered`, or `error:<code>`. Committed as a point-in-time snapshot.
- `scripts/domscan.py` — the scanner that produces `data/full.tsv`.
- `frontend/` — Svelte 5 + Vite single-page app.

## Data pipeline

```
Verisign RDAP  ->  scripts/domscan.py  ->  data/full.tsv
                                              |
                          npm run data  ->  src/lib/domains-names.txt + domains-avail.txt
                                              |
                                       frontend (vite build)
```

`npm run data` (run automatically before `build`) reads `data/full.tsv`, sorts it,
and emits the generated `domains-names.txt` / `domains-avail.txt` that the app imports.
Those two files are gitignored and rebuilt on every build.

## Availability source

Availability is determined authoritatively via **Verisign RDAP**, not WHOIS or DNS:

```
GET https://rdap.verisign.com/com/v1/domain/{name}.com
  200 -> registered
  404 -> available
  429 -> exponential backoff, then retry
  5xx / network error -> retry, then recorded as error:<code>
```

The committed snapshot was produced with:

```sh
python3 scripts/domscan.py --letters aceimnorsuvwxz --workers 8 --rate 30 --out data/full.tsv
```

The scanner is resumable: re-running appends only names not already present in `--out`.
Use `--limit N` for a trial run.

### Caveats

- The scan is a snapshot; domains expire and get registered continuously, so refresh
  it when accuracy matters.
- Respect Verisign's RDAP rate limits and terms of use. The default `--rate 30`
  (requests/second) is aggressive — lower it if you see sustained `429`s.
- `build-data.mjs` currently maps any status other than `available` to `registered`,
  which would fold `error:<code>` rows into "registered". The committed dataset has
  zero error rows, but check for them after a fresh scan.

## Development

```sh
npm install        # installs the frontend workspace
npm run dev        # vite dev server
npm run build      # rebuild data, then build the single-file app
npm test           # vitest
npm run check      # svelte-check + tsc
npm run lint       # prettier --check + eslint
npm run data:scan  # run the availability scanner (writes scan.tsv by default)
```
