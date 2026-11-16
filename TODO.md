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

- 99designs — deferred: from a non-US egress `99designs.com/profiles/{name}`
  GeoIP-redirects to `en.99designs.fr`, which 404s every name — including
  `99designs` itself — so a status row would report everything free. Needs a
  locale-pinned host or a probe from a US egress; the earlier 200/404
  verification does not reproduce here.
- Review `docs/similar-sites.md` — 1,476 candidates, unreferenced and never
  verified. Triage by namespace first (a bare handle beats `{id}`, address and
  store routes), then probe the survivors for a 200/404 split; promote the clean
  ones into `docs/platforms.md` and decide then whether the raw file survives.

## Before mobile

The first three are wins on desktop too, so they come first regardless.

- [x] Cap engine concurrency — `DEFAULT_CONCURRENCY` in `core` holds a scan to
  16 checks in flight, down from one per provider (107 today); `with_concurrency`
  still overrides it for experiments.
- [x] Add a scan deadline — `DEFAULT_DEADLINE` closes a scan at 30 s; whatever
  has not resolved is reported `unknown` ("no verdict before the deadline"),
  which is never cached and self-heals.
- [x] Give `HttpTransport` explicit connect/read timeouts alongside the 10 s
  global bound, so one stalled host cannot hold a worker: 4 s for resolve and
  connect, 5 s for the headers and the body (`with_timeouts` overrides all
  three).
- [x] Persist the cache to `app_data_dir()` — the shell loads `cache.json` at
  startup and rewrites it after each scan; the one-hour ttl still applies and
  expired entries are dropped on save.
- [x] Add recent-checks history — the last six labels live in `localStorage`
  (`web/src/lib/recent.ts`), deduped case-insensitively: chips on the landing
  page and a header menu on the report, with a clear action. Kept out of the
  engine cache so a label outlives the one-hour verdict ttl.
- [x] `viewport-fit=cover` + `env(safe-area-inset-*)` — the body pads for the
  notch and gesture bar, and the sticky header offsets its `top` by the inset.
- [x] ≥44 px tap targets — rows measure 58 px and the filter chips 44 px at a
  360 px viewport; hover is decorative only, never the sole affordance.
- [x] Kill the WebView tap highlight and rubber-band overscroll —
  `-webkit-tap-highlight-color: transparent` and `overscroll-behavior-y: none`.
- [x] Re-check the results header at 360 px — 113 px tall with brand, search and
  theme, no horizontal overflow, and the status chips scroll inside their row.
- [ ] Confirm the existing `"windows": ["main"]` capability applies to the
  mobile webview; the desktop config is already in place.
- [ ] Generate Android mipmaps with `npx tauri icon`.
- [ ] Confirm `devUrl` (`localhost:5221`) survives `adb reverse` for
  `tauri android dev`.
- [ ] Decide the mobile source set — 143 rows, 107 of them checked, is a desktop
  answer.
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
