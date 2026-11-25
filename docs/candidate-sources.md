# Candidate sources

Research note, 2026-11-25. Not part of the catalogue: nothing here has a
manifest row or a provider. Numbers come from today's snapshots of three
username-enumeration corpora plus live `curl` probes (desktop UA, no auth,
single sample per URL).

The first pass is implemented: the four namespace variants in §5, Linktree /
Vimeo / Itch.io from §6, and the A-grade rows in §4 now have manifest rows,
providers, fixtures and docs. Everything else below is backlog.

A second pass on 2026-11-25 mined WhatsMyName, Sherlock and Maigret directly:
713 uncatalogued domains were probed with the shipped client, 177 became
providers (see the corpus pass section in `docs/platforms.md`), and 20 more came
from a registry/namespace sweep (Scoop, CRAN, CTAN, PECL, opam, LuaRocks,
Modrinth, Factorio, Ansible Galaxy, Puppet Forge, Jenkins, Grafana, Quay,
packagecloud, ArtifactHub, Zenodo, Zora, Algorand NFD, Homebrew Cask, Sublime
Package Control). A third pass added Discourse instances and forum software, more package
registries (Stackage, ConanCenter, MacPorts, ELPA, Haxe, Fedora dist-git,
Bioconductor, Shopify apps, Zapier, Node-RED, Keybase), blog hosts (Ghost,
Hatena, Livedoor, Ameba, Cargo, Softr), TON DNS, two more PeerTube instances and
a second Lemmy instance, plus a Tumblr promotion. A fourth pass swept community forums: 53 Discourse instances verified through
`/u/{name}.json` (the auto-created `system` account is a reliable taken probe on
every instance), plus blog hosts (Ghost, Hatena, Livedoor, Ameba, Cargo, Softr,
Exblog, Mataroa, Teachable), TON DNS, PeerTube and Lemmy instances, and regional
platforms (Cnblogs, Douban, Ruby China, Kakuyomu, POSTYPE, Blockdit, Bionluk,
Instamojo, Pensador, Audius). The catalogue is now 537 rows: 505 checked, 3
keyed, 29 deferred.

## 1. What was measured

| Corpus                                                                    | Entries | What it gives                                                                                               |
| ------------------------------------------------------------------------- | ------- | ----------------------------------------------------------------------------------------------------------- |
| [WhatsMyName](https://github.com/webbreacher/whatsmyname) `wmn-data.json` | 691     | per-site URL template, expected/absent codes, present/absent strings, protection flags, known-taken handles |
| [Sherlock](https://github.com/sherlock-project/sherlock) `data.json`      | ~480    | URL template, `errorType` (`status_code` / `message`)                                                       |
| [Maigret](https://github.com/soxoj/maigret) `data.json`                   | 11,189  | URL template, `checkType`, tags                                                                             |

Deduped by registrable domain and diffed against the 211 rows in
`fixtures/sources.json`:

- **3,350** distinct domains are not in the catalogue. Most are dead, NSFW,
  regional or duplicates of an existing platform under another host.
- **154** appear in all three corpora — the high-confidence shortlist.
- **395** appear in two.

Quality signal inside the 154: 152 are detected by a page string rather than a
status code, 9 are flagged Cloudflare, 3 DDoS-Guard. Our `page.rs` adapter
already covers that shape, but a text marker is a maintenance liability, so
live behaviour beats corpus metadata for grading.

Live probes: 47 new candidates and all 36 deferred rows in the manifest
(`status.rs` is not the place to look them up — see §6).

## 2. Ways to find more sources, rated

| #   | Method                                                                                 | Yield                                            | Effort                                                 | Rating                                                                         |
| --- | -------------------------------------------------------------------------------------- | ------------------------------------------------ | ------------------------------------------------------ | ------------------------------------------------------------------------------ |
| 1   | Mine the three enumeration corpora and dedupe by registrable domain                    | 154 high-confidence candidates, ~1,400 plausible | 1 day, scripted                                        | **A** — cheapest, already scored, and it self-updates as the corpora grow      |
| 2   | Add **namespace variants** on platforms we already cover (orgs, groups, teams, scopes) | ~10 clean, high-reach rows from 3 days of work   | low                                                    | **A** — different namespace, same adapter, no new docs research                |
| 3   | Promote the 36 deferred manifest rows                                                  | ~5 actually checkable today                      | very low per row (metadata, docs, limits already done) | **A for the 5, F for the rest** — 14 of 29 probed returned 403                 |
| 4   | Package / artifact / model registries we do not cover                                  | ~15 realistic                                    | medium                                                 | **B+** — boring but stable, clean status codes, good long-tail value           |
| 5   | TLD/RDAP expansion                                                                     | 1 row per TLD, unlimited                         | low per TLD                                            | **B** — high yield, low marginal value; a name is a name in every TLD          |
| 6   | Category-gap sweep against Namechk / KnowEm / Instant Username                         | a few dozen                                      | medium                                                 | **C** — their lists are derivative of Sherlock; the gaps are mostly 403-walled |
| 7   | Regional / non-English platforms (CN, JP, KR, RU, BR, ID)                              | 50+, real reach outside our audience             | high (per-site research, unstable endpoints)           | **C** — do it only with a target audience in mind                              |
| 8   | Web3 name services (Basenames, Space ID, NFD, Tezos, Stacks)                           | 5–8                                              | medium                                                 | **C** — endpoints churn fast; 3 of 5 hosts I tried are already dead            |
| 9   | App-store / marketplace seller namespaces (Play, App Store, Amazon, Etsy)              | 5                                                | high                                                   | **D** — auth or JS-only in practice                                            |
| 10  | Dorking for `/user/{name}`, `/@{name}`, `/u/{name}` patterns                           | unbounded                                        | high                                                   | **D** — no way to separate dead sites from live ones at scale                  |

Rule of thumb confirmed by the probes: **a public, unauthenticated JSON or HTML
endpoint that answers 200/404 without Cloudflare is the only shape that
survives**; everything else rots or blocks. About half of the popular social
networks left in the corpora fail that test.

## 3. Rating rubric

- **Signal** — measured today with two names (one known-taken, one random):
  5 = clean split, 4 = split with a header/API detail, 3 = split only on body
  text, 2 = 403/429/needs JS, 1 = no signal.
- **Reach** — how often a person checking a brand would care (editorial).
- **Effort** — S = one `spec(...)` line; M = new adapter path or request
  detail; L = bespoke/auth.
- **Grade** — A: ship now; B: good, one wrinkle; C: niche or fragile; D: no.

## 4. Shortlist

### Developer platforms

| Source        | Endpoint                               | Probed                      | Signal | Effort   | Grade |
| ------------- | -------------------------------------- | --------------------------- | ------ | -------- | ----- |
| Docker Hub    | `hub.docker.com/u/{name}`              | 200 / 404                   | 4      | S        | **A** |
| GitHub orgs   | `api.github.com/orgs/{name}`           | 200 / 404                   | 5      | S        | **A** |
| GitLab groups | `gitlab.com/api/v4/groups/{name}`      | 200 / 404                   | 5      | S        | **A** |
| npm orgs      | `registry.npmjs.org/-/org/{name}/user` | 200 / 404                   | 5      | S        | **A** |
| PyPI orgs     | `pypi.org/org/{name}/`                 | 200 / 404                   | 4      | S        | **A** |
| Replit        | `replit.com/@{name}`                   | 200 / 404                   | 4      | S        | **A** |
| HackerRank    | `hackerrank.com/{name}`                | 200 / 404                   | 4      | S        | **B** |
| Codeforces    | `codeforces.com/profile/{name}`        | 200 / 400                   | 4      | S        | **B** |
| GeeksforGeeks | `geeksforgeeks.org/user/{name}/`       | 200 / 400                   | 4      | S        | **B** |
| AtCoder       | `atcoder.jp/users/{name}`              | 200 / 404                   | 4      | S        | **B** |
| devRant       | `devrant.com/users/{name}`             | 200 / 400                   | 4      | S        | **B** |
| Pastebin      | `pastebin.com/u/{name}`                | 200 / 404                   | 4      | S        | **B** |
| Bugcrowd      | `bugcrowd.com/{name}`                  | 200 / 404                   | 4      | S        | **B** |
| Intigriti     | `intigriti.com/{name}`                 | 200 / 404                   | 4      | S        | **B** |
| asciinema     | `asciinema.org/~{name}`                | 200 / 404                   | 4      | S        | **C** |
| Crowdin       | `crowdin.com/profile/{name}`           | 200 / 404                   | 4      | S        | **C** |
| WakaTime      | `wakatime.com/@{name}`                 | 200 / 404                   | 4      | S        | **C** |
| Monkeytype    | `monkeytype.com/profile/{name}`        | 200 / 404                   | 4      | S        | **C** |
| Scratch       | `scratch.mit.edu/users/{name}`         | 200 + marker / 200 + marker | 3      | S (page) | **B** |
| TryHackMe     | `tryhackme.com/p/{name}`               | 429 / 429                   | 2      | M        | **C** |
| Weblate       | `weblate.org/user/{name}/`             | 403 / 403                   | 1      | —        | **D** |
| LeetCode      | `leetcode.com/{name}/`                 | 400 / 400 (JS)              | 1      | L        | **D** |
| Hashnode      | `hashnode.com/@{name}`                 | 200 / 200                   | 2      | M (API)  | **C** |
| Keybase       | `keybase.io/{name}`                    | 200 / 200 (site is inert)   | 1      | —        | **D** |

### Social, reading, lifestyle

| Source        | Endpoint                         | Probed                      | Signal | Effort   | Grade |
| ------------- | -------------------------------- | --------------------------- | ------ | -------- | ----- |
| MyAnimeList   | `myanimelist.net/profile/{name}` | 200 / 404                   | 4      | S        | **A** |
| OpenStreetMap | `openstreetmap.org/user/{name}`  | 200 + marker / 200 + marker | 3      | S (page) | **A** |
| TradingView   | `tradingview.com/u/{name}/`      | 200 / 404                   | 4      | S        | **A** |
| Trello        | `trello.com/{name}`              | 200 / 404                   | 4      | S        | **B** |
| Speaker Deck  | `speakerdeck.com/{name}`         | 200 / 404                   | 4      | S        | **B** |
| omg.lol       | `{name}.omg.lol`                 | 200 / 404                   | 4      | S        | **C** |
| Vero          | `vero.co/{name}`                 | 200 / 404                   | 4      | S        | **C** |
| MySpace       | `myspace.com/{name}`             | 200 / 404                   | 4      | S        | **C** |
| Flightradar24 | `flightradar24.com/{name}`       | 200 / 404                   | 4      | S        | **C** |
| IFTTT         | `ifttt.com/{name}`               | 200 / 404                   | 4      | S        | **C** |
| Plurk         | `plurk.com/{name}`               | 200 / 200                   | 2      | M        | **C** |
| Letterboxd    | `letterboxd.com/{name}/`         | 403 / 403                   | 1      | —        | **D** |
| Untappd       | `untappd.com/user/{name}`        | 403 / 403                   | 1      | —        | **D** |
| VSCO          | `vsco.co/{name}`                 | 403 / 403                   | 1      | —        | **D** |
| SlideShare    | `slideshare.net/{name}`          | 406 / 406                   | 1      | —        | **D** |
| Wikidot       | `{name}.wikidot.com`             | 200 / 200                   | 2      | M        | **D** |

### Creative, media, money

| Source          | Endpoint                       | Probed    | Signal | Effort  | Grade           |
| --------------- | ------------------------------ | --------- | ------ | ------- | --------------- |
| ArtStation      | `artstation.com/{name}`        | 200 / 404 | 4      | S       | **A**           |
| Freesound       | `freesound.org/people/{name}/` | 200 / 404 | 4      | S       | **B**           |
| Ko-fi           | `ko-fi.com/{name}`             | 200 / 404 | 4      | S       | **B**           |
| Carrd           | `{name}.carrd.co`              | 200 / 404 | 4      | S       | **B**           |
| Tenor           | `tenor.com/u/{name}`           | 200 / 404 | 4      | S       | **B**           |
| SmugMug         | `{name}.smugmug.com`           | 200 / 404 | 4      | S       | **C**           |
| Carbonmade      | `{name}.carbonmade.com`        | 200 / 404 | 4      | S       | **C**           |
| Venmo           | `venmo.com/u/{name}`           | 200 / 404 | 4      | S       | **C** (US only) |
| Unsplash        | `unsplash.com/@{name}`         | 401 / 401 | 2      | M (key) | **C**           |
| Imgur           | `imgur.com/user/{name}`        | 429 / 429 | 2      | M       | **C**           |
| Buy Me a Coffee | `buymeacoffee.com/{name}`      | 404 / 404 | 1      | —       | **D**           |
| Coroflot        | `coroflot.com/{name}`          | 403 / 403 | 1      | —       | **D**           |
| FurAffinity     | `furaffinity.net/user/{name}`  | 403 / 403 | 1      | —       | **D**           |

### Gaming

| Source        | Endpoint                           | Probed    | Signal | Effort | Grade              |
| ------------- | ---------------------------------- | --------- | ------ | ------ | ------------------ |
| osu!          | `osu.ppy.sh/users/{name}`          | 200 / 404 | 4      | S      | **A**              |
| Geocaching    | `geocaching.com/profile/?u={name}` | 200 / 404 | 4      | S      | **B**              |
| Xbox gamertag | `xboxgamertag.com/search/{name}`   | 200 / 404 | 4      | S      | **C** (aggregator) |
| TETR.IO       | `tetr.io/@{name}`                  | 200 / 404 | 4      | S      | **C**              |
| Nightbot      | `nightbot.tv/dashboard/{name}`     | 200 / 404 | 4      | S      | **C**              |
| Speedrun.com  | `speedrun.com/user/{name}`         | 403 / 403 | 1      | —      | **D**              |
| Newgrounds    | `newgrounds.com/{name}`            | 403 / 403 | 1      | —      | **D**              |
| PSNProfiles   | `psnprofiles.com/{name}`           | 403 / 403 | 1      | —      | **D**              |
| Wowhead       | `wowhead.com/user/{name}`          | 403 / 403 | 1      | —      | **D**              |
| RuneScape     | `apps.runescape.com/...`           | DNS gone  | 1      | —      | **D**              |

## 5. Namespace variants — the cheapest wins

These are _new rows_, not new providers: each reuses an existing adapter and
adds a namespace the app currently misses.

| Variant             | Endpoint                               | Probed    | Grade |
| ------------------- | -------------------------------------- | --------- | ----- |
| GitHub organisation | `api.github.com/orgs/{name}`           | 200 / 404 | **A** |
| GitLab group        | `gitlab.com/api/v4/groups/{name}`      | 200 / 404 | **A** |
| Docker Hub owner    | `hub.docker.com/u/{name}`              | 200 / 404 | **A** |
| npm org             | `registry.npmjs.org/-/org/{name}/user` | 200 / 404 | **A** |
| PyPI organisation   | `pypi.org/org/{name}/`                 | 200 / 404 | **A** |
| Codeberg org        | `codeberg.org/org/{name}`              | untested  | **B** |

## 6. Promoting deferred rows

All 36 non-provider rows carry manifest metadata, docs rows and limits, so a
promotion is the cheapest possible source. Probed today:

| Row                                                                                                                            | Result               | Verdict                                   |
| ------------------------------------------------------------------------------------------------------------------------------ | -------------------- | ----------------------------------------- |
| Linktree                                                                                                                       | 200 / 406            | **A**, works with an HTML `Accept` header |
| Vimeo                                                                                                                          | 200 / 404            | **A**                                     |
| Figma Community                                                                                                                | 200 / 404            | **A**                                     |
| itch.io                                                                                                                        | 200 / 404            | **A**                                     |
| PyPI users                                                                                                                     | 200 / 200            | needs a body check                        |
| Reddit, Facebook, Pinterest                                                                                                    | 301 / 301            | redirect to login; needs OAuth            |
| OpenSea, Product Hunt, Tumblr, Genius, Medium, Patreon, Kickstarter, CodeSandbox, npm users, Fandom, Epic, Etsy, eBay, CodePen | 403 / 403            | Cloudflare/Akamai                         |
| LinkedIn                                                                                                                       | 999 / 999            | hard-blocked                              |
| Trakt, Last.fm, Lemon Squeezy, Unstoppable, Etsy-adjacent                                                                      | same code both names | no usable signal                          |

~14 of 29 are bot-walled. Promote the four that work; do not chase the rest
with header games.

## 7. Domain/TLD expansion

The RDAP adapter already checks 7 TLDs. Adding a TLD is one manifest row plus a
bootstrap entry, and the probe pattern generalises:

| TLD    | RDAP host                            | Probed                                       |
| ------ | ------------------------------------ | -------------------------------------------- |
| `.xyz` | `rdap.centralnic.com/xyz`            | 200 / 404                                    |
| `.io`  | `rdap.identitydigital.services/rdap` | 200 / 404 (already covered)                  |
| `.me`  | `rdap.nic.me`                        | host no longer resolves — needs re-discovery |
| `.now` | `rdap.centralnic.com/now`            | 404 / 404, response shape changed            |

Verdict: the mechanism is cheap but almost entirely a data problem — IANA's
bootstrap file lists ~1,200 TLDs, most with no RDAP. Pick the handful the
audience actually asks for (`.co`, `.xyz`, `.me`, `.gg`, `.tv`, `.uk`, `.de`,
`.fr`, `.nl`, `.eu`) rather than trying to cover the namespace.

## 8. What I would do next

1. **Ship the five namespace variants** (§5) — no new adapter, 5 rows, high reach.
2. **Promote Linktree, Vimeo, Figma, itch.io** (§6) — the docs already exist.
3. **Add the A/B rows from §4**: Docker Hub, MyAnimeList, ArtStation, osu!,
   Replit, HackerRank, TradingView, OpenStreetMap, Trello, Freesound, Geocaching.
4. Re-run `npm run probe` before each PR; these endpoints rot within months.
5. Leave §7 for a dedicated task; it is a data-volume decision, not a
   capability one.
