# Fixtures

This directory is the parity oracle for the engine rewrite. `sources.json` is the
complete source inventory, and each `<label>/<source>.json` is one provider's
captured result plus the raw HTTP request(s) and response(s) it made.

The checked-in captures are `acema` (mostly available) and `google` (mostly
taken), so both verdict branches are exercised.

## Regenerate

The Rust-only tree does not have a fixture exporter yet, so these captures are
frozen. If a source changes and a capture must be refreshed, recover the
exporter from the Git history before the Go server was removed.

With the old Go exporter it was:

```sh
npm run server:export -- --label acema
```

`--label` was repeatable. With no label only `sources.json` was written.

## Format

- `sources.json` — every source, including credential-gated and deferred ones
  that a zero-config scan omits. `keyed: true` marks a source that needs
  credentials; a source with no provider renders as "not checked".
  `caseSensitive: true` keeps the user's raw label case for that source (the
  engine lowercases for everyone else). `minLength`/`maxLength` short-circuit
  labels the source cannot own; each limit's provenance is recorded per source
  in `docs/platforms.md`. The registry is compiled into the binary via
  `Manifest::embedded`; `--manifest` overrides it during development.
- `<label>/<source>.json` — `status`, `detail`, the provider's public URL, and
  a `requests` array. Each request carries its method, URL, headers, body, and
  the response status, selected headers, and body. A transport failure is
  recorded in `error` with no `response`.

Bodies are recorded as `bodyEncoding: "utf8"` or `"base64"`. The provider always
receives the full body; only the snapshot is capped at 256 KiB, in which case
`bodyTruncated` is true. Each fixture calls the provider's `Check` directly; the
engine's minimum/maximum-length short-circuits are described by
`minLength`/`maxLength` in `sources.json` and must be tested separately.

## Caveats

- This is a **capture**, not live truth. Sources change and rate-limit; refresh
  the capture from the archived exporter instead of treating an old response as
  current.
- Keyed sources appear only in `sources.json`; they are not captured because
  that would require secrets.
- Secrets are redacted from request headers and query strings. Response cookies
  and noisy headers are not recorded.
- DNS providers make no HTTP request, so their fixture has an empty `requests`
  array. Domain fixtures record the RDAP request but not the shared IANA
  bootstrap fetch.
