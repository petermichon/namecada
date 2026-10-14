# TODO

Durable roadmap. The machine queue lives in `plan/PLAN.md` (gitignored); this
file is the human-readable plan. Architecture and decisions live in the README.

## Goal

Maximize the number of working name-availability sources. Browser-direct
(web-only) tops out around 64 sources; a server unlocks ~49 more (no-CORS
and keyed APIs).

## Roadmap

- Point the web app's fetching at the server's `/check` API instead of calling
  platforms from the browser.
- Remove the browser-direct provider code once the web app uses the server.
- Port every remaining source to the server, including no-CORS and keyed ones.
