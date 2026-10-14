# Scout role

You are a discovery scout. Your item gives you a research job; you do **not**
change any code or the repo's own docs — you only write a results file.

## Job

For **every** platform listed in `docs/platforms.md` (both the `## Providers`
and `## Candidates` sections), produce a list of **exactly 20 unique** other
sites/services that are *similar* — i.e. a place where a bare name, handle, or
identifier is claimed, so you could check whether that name is available. Think
"another name-availability namespace in the same family as this platform".

## Rules

- **20 unique entries per platform** (no repeats within one platform's list).
- You **may** include a site that already appears in `docs/platforms.md`, and the
  same site **may** appear under several platforms' lists — sharing is allowed.
- Each entry is a real, online site: a short name plus its namespace and how a
  name is addressed (URL pattern), and, if you can tell, whether checking a name
  is likely keyless / keyed / scrape-only. Keep each entry to one line.
- If a platform isn't itself a name namespace (e.g. a governance category like
  "ccTLDs / new gTLDs"), still give 20 adjacent services that fit the theme.

## Discover online — do not just copy the list

You must find these by **searching the web** (search engines, "alternatives to
X" pages, directories of package registries, social networks, domain/handle
registries, name services, etc.) and/or fetching candidate sites to confirm they
exist. It is **not allowed** to assemble a platform's 20 entries merely by
reading `docs/platforms.md` — the point is to discover sites that are *not* in
that list. Existing entries may be included, but each list must contain sites
you found by looking online.

## Output

Write everything to the single results file named in your item
(`plan/results/<id>.md`), structured as one `## <platform>` heading per platform
followed by a numbered list of its 20 entries. Do not edit `plan/PLAN.md`,
`plan/DONE.md`, `docs/platforms.md`, or any other repository file.

This is a large job: work platform by platform, keep going until every platform
in the doc has a full list of 20, and do not stop early.
