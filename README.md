# Domain Browser

A domain availability checker. Enter a full domain including its TLD
(`acema.com`, `acema.io`, `acema.co.uk`) and it tells you whether it is
registered, using the authoritative RDAP registry, fetched directly from the
browser.

## How it works

An explicit TLD is required — no default is assumed. The input is reduced to its
**registrable domain** (eTLD+1, via the Public Suffix List) — so `abc.def.com`
checks `def.com` and `www.bbc.co.uk` checks `bbc.co.uk`, since registries only
resolve registrable domains. The TLD is then resolved to its registry via the
IANA RDAP bootstrap (`data.iana.org/rdap/dns.json`, cached for 7 days), and a
single request is made:

```
GET {registry-base}domain/{name}     e.g. https://rdap.verisign.com/com/v1/domain/acema.com
  200 -> registered
  404 -> not registered
  anything else -> unknown
```

Registries send `Access-Control-Allow-Origin: *`, so no backend or proxy is
needed. Verdicts are cached in `localStorage`, shown instantly, and revalidated in
the background once older than 1h (stale-while-revalidate). Registered domains
also show their registry registration date. `.com`/`.net` have a built-in
fallback so they work even if the bootstrap is unreachable.

**Limitations.** RDAP reports registry state, not registrability: a name may be
reserved/blocked (but unregistered), and premium pricing or a `pendingDelete`
state is not reflected. Coverage follows the bootstrap — TLDs not listed there
(e.g. `.io`, `.co`, `.de`) report "no RDAP endpoint". For "is it registered?",
the registry is authoritative.

## Layout

- `frontend/` — Svelte 5 + Vite + Tailwind app (the whole product).
- `scripts/domscan.py` — offline bulk scanner for `.com` names via Verisign RDAP
  (offline research tooling; produces the dataset below). Not used by the app.
- `data/full.tsv` — legacy bulk snapshot (`name<TAB>status`) from that scanner.

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

## Bulk scanning (offline)

`scripts/domscan.py` brute-forces the 5-letter space over a chosen alphabet and
appends results to a TSV, resumable and rate-limited:

```sh
python3 scripts/domscan.py --letters aceimnorsuvwxz --workers 8 --rate 30 --out data/full.tsv
```

Respect Verisign's RDAP rate limits — lower `--rate` if you see sustained `429`s.
