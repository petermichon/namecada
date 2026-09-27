# domain-browser

A `.com` availability checker. Enter a domain label and it tells you whether the
`.com` is registered — using the authoritative Verisign RDAP registry, fetched
directly from the browser.

## How it works

The app makes a single request per lookup:

```
GET https://rdap.verisign.com/com/v1/domain/{name}.com
  200 -> registered
  404 -> not registered
  anything else -> unknown
```

Verisign sends `Access-Control-Allow-Origin: *` on both responses, so no backend
or proxy is needed. Verdicts are cached in `localStorage` for 24h.

**Limitations.** RDAP reports registry state, not registrability: a name may be
reserved/blocked (but unregistered), and premium pricing or a `pendingDelete`
state is not reflected. For "is it registered?", the registry is authoritative.

## Layout

- `frontend/` — Svelte 5 + Vite + Tailwind app (the whole product).
- `scripts/domscan.py` — offline bulk scanner for `.com` names via Verisign RDAP
  (offline research tooling; produces the datasets below). Not used by the app.
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
