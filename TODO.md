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

## Before mobile

The first three are wins on desktop too, so they come first regardless.

- [ ] Cap engine concurrency — `Engine::new` opens one connection per provider,
  ~130 at once. Settle on 16 in flight, as a constant in `core` so `serve`, the
  desktop shell and Android agree; keep `with_concurrency` for experiments.
- [ ] Add a scan deadline — at 16 workers and a 10 s request timeout a bad scan
  still runs ~90 s. Whatever has not resolved by the deadline becomes `unknown`,
  which is uncached and self-heals.
- [ ] Give `HttpTransport` explicit connect/read timeouts alongside the 10 s
  global bound, so one stalled host cannot hold a worker.
- [ ] Persist the cache to `app_data_dir()` — it is process-only today, so every
  Android relaunch is a cold scan.
- [ ] Add recent-checks history; it falls out of the persisted cache, and it is
  what makes a phone app worth opening.
- [ ] `viewport-fit=cover` + `env(safe-area-inset-*)`, or the notch and gesture
  bar clip the sticky header.
- [ ] ≥44 px tap targets and no hover-only affordances in `ResultRow`.
- [ ] Kill the WebView tap highlight and rubber-band overscroll.
- [ ] Re-check the results header at 360 px: brand, search, theme, two filter
  rows.
- [ ] Capability targets `"windows": ["main"]`; confirm it applies to the mobile
  webview.
- [ ] Generate Android mipmaps with `npx tauri icon`.
- [ ] Confirm `devUrl` (`localhost:5221`) survives `adb reverse` for
  `tauri android dev`.
- [ ] Decide the mobile source set — 130 sources is a desktop answer.
- [ ] Decide signing and commit policy: commit `src-tauri/gen/android`, keep the
  keystore out of the repo.
- [ ] Test on the phone's actual WebView (Tailwind 4 needs Chrome/WebView 111+).

## Mobile

Android only; iOS needs macOS, so it is out of scope here. Once the list above is
green:

1. Toolchain: SDK + NDK, `rustup target add aarch64-linux-android`, `cargo-ndk`,
   then `npx tauri android init`.
2. `npx tauri android dev` over USB, then `npx tauri android build` and
   sideload.

Same IPC and engine as the desktop shell.
