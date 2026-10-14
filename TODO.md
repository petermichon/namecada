# TODO

Durable roadmap. The machine queue lives in `plan/PLAN.md` (gitignored); this
file is the human-readable plan. Architecture and decisions live in the README.

## Goal

Maximize the number of working name-availability sources. Browser-direct
(frontend-only) tops out around 64 sources; a server unlocks ~49 more (no-CORS
and keyed APIs).

## Roadmap

- Move the web app onto the server API and drop the browser-direct sources.
- Port every remaining source to the server, including no-CORS and keyed ones.
- Add a Datasette front over the SQLite file for browsing without a web app.
