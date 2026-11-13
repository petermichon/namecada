# TODO

Durable roadmap. The machine queue lives in `plan/PLAN.md` (gitignored); this
file is the human-readable plan. Architecture and decisions live in the README;
the full candidate list, with per-source reachability notes, lives in
`docs/platforms.md`.

## Goal

Report whether a bare name is taken across as many independent sources as
possible. Checks run in the Rust server, so CORS is no longer the ceiling — the
remaining limit is each source's API shape (keys, scrapes, non-label
namespaces), not the browser.

## Next

Ordered by preference; `docs/platforms.md` lists the sources in each. Promote a
source to a `plan/PLAN.md` item when it reaches the front of the line.

1. Official keyless API — the default. Verify each on the taken/free response.
2. Body-parse a server-rendered page — fallback when there's no API: match a
   stable marker (e.g. `og:title`/`"username"`), never localized "not found"
   text; return `unknown` on ambiguity.
3. Keyed API — **currently omitted**: no row, request, or cache entry. Add one
   only when neither (1) nor (2) can answer and the source is high-value.
4. Headless-browser scraping — avoid; slow, brittle, usually bot-blocked. Never
   scrape behind a login.

## Sources

- Add 99designs — `99designs.com/profiles/{name}` answers 200 when taken and 404
  when free, status-only with no JS, so it is a status row. Match the status
  code, not the page title: the `.fr` locale redirect keeps the status, but the
  title is generic for numeric-id profiles.
- Review `docs/similar-sites.md` — 1,476 candidates, unreferenced and never
  verified. Triage by namespace first (a bare handle beats `{id}`, address and
  store routes), then probe the survivors for a 200/404 split; promote the clean
  ones into `docs/platforms.md` and decide then whether the raw file survives.

## Mobile

Wanted eventually, in this order:

1. Responsive pass on the existing web UI at ~360 px: filter chips/header
   wrapping, compact result rows, ≥44 px tap targets, no hover-only affordances.
2. Then the Android shell: SDK + NDK, `rustup target add aarch64-linux-android`,
   `cargo-ndk`, `npx tauri android init`, `npx tauri android dev` over USB.
   Same IPC and engine; iOS needs macOS, so it is out of scope here.
