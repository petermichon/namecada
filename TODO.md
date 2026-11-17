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

## Mobile

Android only; iOS needs macOS, so it is out of scope here. The toolchain is in
place (see `docs/android.md`); the app runs on the A015 and a full scan takes
about three seconds.

1. Dev: `npm run android:dev` (see `docs/android.md` for the SDK setup).
2. Release: `npx tauri android build --apk`, then `npm run android:sign`.
3. Distribute: tag `v<version>` and the workflow attaches signed APKs for
   Obtainium; the Fastlane listing and the F-Droid recipe are in
   `docs/distribution.md`.

Same IPC and engine as the desktop shell.

## Release

`v0.1.0` shipped from a tag: the workflow builds, signs and attaches
`namecada-<version>-<abi>.apk` plus SHA-256 sums, and Obtainium follows the
repository. One-time phone setup: add the repository URL with an `arm64-v8a`
APK filter, and add Obtainium's own repository so it updates itself. Next
release: bump `version` in
`apps/desktop/src-tauri/tauri.conf.json` (versionCode derives from it), commit,
tag `v<version>`, push. The workflow rejects a tag that disagrees with the
config.

## Site

`apps/site` is the page at `namecada.com`: plain HTML/CSS built by Vite, behind
nginx on the edge host. Parked:

- link "Add to Obtainium" from the download block — it is the install path this
  project actually uses, and the page does not offer it;
- stop "Download for Android" pointing at `releases/latest`: one desktop-only
  tag makes the button lie;
- `og:image` from `docs/banner.png`, and the report screenshot in the page.

## Maintenance

- `fixtures/README.md`: record what the 8.7 MB of captured bodies are, when they
  were taken, and that they exist only as a test oracle.
- Decide whether `gen/android` is committed or regenerated: the APK is built
  from files the repository does not contain. The pinned Tauri CLI covers the
  acute risk.
- Second Fastlane screenshot: only when filing IzzyOnDroid or F-Droid.

