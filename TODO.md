# TODO

Durable roadmap. The machine queue lives in `plan/PLAN.md` (gitignored); this
file is the human-readable plan. Architecture and decisions live in the README;
the full candidate list, with per-source reachability notes, lives in
`docs/platforms.md`.

## Goal

Report whether a bare name is taken across as many independent sources as
possible. The app is server-only now, so CORS is no longer the ceiling — the
remaining limit is each source's API shape (keys, scrapes, non-label
namespaces), not the browser.

## Next

Ordered phases; `docs/platforms.md` lists the sources in each. Promote a source
to a `plan/PLAN.md` item when it reaches the front of the line.

1. No-CORS registries the server can now reach — the largest batch.
2. Hosting subdomains (`{name}.<host>`): server-side HTTP status sidesteps
   wildcard DNS.
3. Free subdomain namespaces, via DoH or HTTP status.
4. Hacker News — now viable: the server reads the profile page.
5. Keyed sources — gated on a key/config story.
6. Trademark and company registries — keyed, and marks are not names.

