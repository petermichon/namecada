# Fixtures

This directory is the parity oracle for the engine rewrite. `sources.json` is the
complete source inventory, and each `<label>/<source>.json` is one provider's
captured result plus the raw HTTP request(s) and response(s) it made.

## Regenerate

From the repository root:

```sh
npm run server:export -- --label acema
```

`--label` is repeatable. With no label only `sources.json` is written.

## Format

- `sources.json` — every source, including the credential-gated ones that a
  zero-config scan omits. `keyed: true` marks a source that needs credentials.
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

- This is a **capture**, not live truth. Sources change and rate-limit; re-run
  the export instead of treating an old response as current.
- Keyed sources appear only in `sources.json`; they are not captured because
  that would require secrets.
- Secrets are redacted from request headers and query strings. Response cookies
  and noisy headers are not recorded.
- DNS providers make no HTTP request, so their fixture has an empty `requests`
  array. Domain fixtures record the RDAP request but not the shared IANA
  bootstrap fetch.
