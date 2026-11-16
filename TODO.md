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

- Review `docs/similar-sites.md` — 1,476 candidates, unreferenced and never
  verified. Triage by namespace first (a bare handle beats `{id}`, address and
  store routes), then probe the survivors for a 200/404 split; promote the clean
  ones into `docs/platforms.md` and decide then whether the raw file survives.

## Before mobile

Device and platform checks left before the Android build.

- [ ] Confirm the existing `"windows": ["main"]` capability applies to the
  mobile webview; the desktop config is already in place.
- [ ] Generate Android mipmaps with `npx tauri icon`, from a 1024 px source
  (the current `icons/icon.png` is 512).
- [ ] Confirm `devUrl` (`localhost:5221`) survives `adb reverse` for
  `tauri android dev`.
- [ ] Decide the mobile source set — 143 rows, 107 checked; measure a real scan
  on the phone first, since the 30 s deadline already bounds the worst case.
- [ ] Decide signing and commit policy: commit `src-tauri/gen/android`, keep the
  keystore out of the repo.
- [ ] Test on the phone's actual WebView (Tailwind 4 needs Chrome/WebView 111+).

## Mobile

Android only; iOS needs macOS, so it is out of scope here. Once the list above is
green:

1. Toolchain — see `docs/android.md` for the SDK, NDK, Rust target and
   `cargo-ndk`; `adb` and Java are already present and the JDK is the one open
   risk. Then `npx tauri android init`.
2. `npx tauri android dev` over USB, then `npx tauri android build` and
   sideload.

Same IPC and engine as the desktop shell.
