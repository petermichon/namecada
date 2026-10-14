# TODO

Durable roadmap. The machine queue lives in `plan/PLAN.md` (gitignored); this
file is the human-readable plan. Architecture and decisions live in the README.

## Goal

Maximize the number of working name-availability sources. Browser-direct
(frontend-only) tops out around 64 sources; a backend unlocks ~49 more (no-CORS
APIs and keyed APIs).

## Roadmap

1. **Core** — done: scaffold, provider core (Lemmy, PeerTube), scan engine,
   in-memory store, JSON sink.
2. **Sinks** — pending: SQLite file sink (`check <label> --out x.sqlite`); HTTP
   API (`serve`).
3. **Web app** — pending: consume the API, keep browser-direct as a fallback.
4. **Sources** — port the remaining browser-direct providers; then use the
   backend to reach no-CORS and keyed sources.
5. **Optional front** — Datasette over the SQLite file for a browse/filter UI
   with no frontend code.
