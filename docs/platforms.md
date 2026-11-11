# Platform backlog

Candidate sources for checking whether a name is taken. Flat list — one entry per
line.

**Rule:** nothing gets added until its signal is verified **server-side** on both
the taken and free response (see the README's "Adding a source"). CORS is no
longer a constraint — checks run in the Rust core — so the reachability notes
below, written in the browser-direct era, must be re-checked before implementing.

## Providers

- [x] Domains (RDAP TLDs) — ✅ — `.com .net .org .dev .io .ai .app` (`.io` via fallback) + any TLD in the IANA bootstrap
- [x] ENS (`.eth`) — ✅ — `api.ensideas.com/ens/resolve/{name}.eth` (200 always; non-null `address` = taken, `null` = free; third-party resolver — a follow-up could query a public Ethereum RPC against the ENS registry directly)
- [x] Handshake (HNS) — ✅ — `hnsscan.com/api/names/{name}/details` (200 for a known name; `info.state` `CLOSED` = taken, `AVAILABLE`/`UNKNOWN` = free — never use the historical `info.registered` flag alone; reserved names 502; third-party explorer with an undocumented API)
- [x] Namecoin (`.bit`) — ✅ — `explorer.namecoin.co/api/name/d%2F{name}` (the `d/` namespace backs `.bit`; 404 = never registered/free, 200 with `expired:false` = taken, `expired:true` = lapsed/free; third-party explorer with an undocumented API)
- [x] Solana Name Service (`.sol`) — ✅ — `sdk-proxy.sns.id/resolve/{name}.sol` (200 always; `{"s":"ok","result":"<pubkey>"}` = taken, `{"s":"error","result":"Domain not found"}` = free; third-party SDK proxy with an undocumented wire shape — anything else degrades to unknown)
- [x] npm — ✅ — `registry.npmjs.org/{name}`
- [x] PyPI — ✅ — `pypi.org/pypi/{name}/json`
- [x] crates.io — ✅ — `crates.io/api/v1/crates/{name}`
- [x] NuGet — ✅ — `api.nuget.org/v3-flatcontainer/{name}/index.json`
- [x] RubyGems — ✅ — `rubygems.org/api/v1/gems/{name}.json`
- [x] Hex (Elixir) — ✅ — `hex.pm/api/packages/{name}` (packages) and `hex.pm/api/users/{name}` (users)
- [x] pub.dev (Dart) — ✅ — `pub.dev/api/packages/{name}`
- [x] JSR (JavaScript) — ✅ — `jsr.io/api/scopes/{name}`
- [x] Homebrew — ✅ — `formulae.brew.sh/api/formula/{name}.json`
- [x] Chocolatey — ✅ — `community.chocolatey.org/packages/{name}` (status-only HTML; package ids are case-insensitive)
- [x] AUR (Arch) — ✅ — `aur.archlinux.org/packages/{name}` (status-only HTML; the RPC endpoint always answers 200)
- [x] Snap Store — ✅ — `api.snapcraft.io/v2/snaps/info/{name}` with `Snap-Device-Series: 16` (200 taken / 404 free; without the header every request fails, so this is the one custom check that sets its own headers; snap names are 1–40 lowercase letters, digits and hyphens, hence `maxLength` 40; page `snapcraft.io/{name}`)
- [x] Chef Supermarket — ✅ — `supermarket.chef.io/api/v1/cookbooks/{name}` (200 taken / 404 free; the old "no CORS" note predated server-side checks)
- [x] CocoaPods — ✅ — `cocoapods.org/pods/{name}` (200 taken / 404 free; the website route resolves pod names case-insensitively — `alamofire` renders Alamofire — so lowercased labels are safe here, unlike the exact-case trunk API)
- [x] SourceForge — ✅ — `sourceforge.net/projects/{name}/` (200 taken / 404 free; the 404 body is JSON, the 200 is HTML, so status alone is the signal)
- [x] Hackage — ✅ — `hackage.haskell.org/package/{name}` (bare name resolves; status-only HTML)
- [x] Clojars — ✅ — `clojars.org/api/artifacts/{name}` (single segment resolves)
- [x] WordPress plugins/themes — ✅ — `api.wordpress.org/plugins/info/1.2`
- [x] GitHub — ✅ — `api.github.com/users/{name}`
- [x] Bluesky — ✅ — `bsky.social/xrpc/com.atproto.temp.checkHandleAvailability` (`{name}.bsky.social`); resolveHandle hides reserved/deactivated/over-length (max 18) handles
- [x] Mastodon (`mastodon.social`) — ✅ — `/@{name}` profile (200/410 taken, 404 free; the lookup API hides suspended accounts as 404)
- [x] GitLab — ✅ — `gitlab.com/api/v4/users?username={name}` (array body), then `/api/v4/groups/{name}` as a fallback
- [x] Bitbucket — ✅ — `api.bitbucket.org/2.0/workspaces/{name}`
- [x] Codeberg — ✅ — `codeberg.org/api/v1/users/{name}`
- [x] Farcaster — ✅ — `fnames.farcaster.xyz/transfers/current?name={name}`
- [x] DEV — ✅ — `dev.to/api/users/by_username?url={name}`, then `/api/organizations/{name}` as a fallback
- [x] Wikipedia — ✅ — `meta.wikimedia.org/w/api.php` globaluserinfo (`missing` = free)
- [x] Substack — ✅ — `{name}.substack.com` (wildcard DNS; HTTP 404, HTML Accept; user-chosen publication names)
- [x] X / Twitter — ✅ — `x.com/{name}` (keyless profile; 200 taken / 404 free; twitter.com 301s to x.com)
- [x] YouTube — ✅ — `youtube.com/@{name}` (profile redirects and resolves 200 taken / 404 free; keyless, no Data API key needed)
- [x] Twitch — ✅ — `api.twitch.tv/helix/users?login={name}` (Client-Id + app access token via client-credentials grant; non-empty `data` taken / empty free)
- [x] Instagram — ✅ — `instagram.com/{name}/` (body signal: a claimed profile embeds `"username":"{name}"` — or a `<title>` starting with the handle — while a free name returns the generic `<title>Instagram</title>` shell; every name is a 200, so status alone has no signal)
- [x] Hacker News — ✅ — `news.ycombinator.com/user?id={name}` (body signal: `No such user.` = free, `karma` = taken; the firebase JSON API only exposes users with public activity, so `null` can't confirm free)
- [x] HackerNoon — ✅ — `hackernoon.com/u/{name}` (keyless profile: 200 taken / 404 free)
- [x] Open Collective — ✅ — `opencollective.com/{name}.json` (200 taken / 404 free). The CDN caches responses for 1h and can cache a transient 5xx per-`Origin`; the provider retries once with a cache-busting param to bypass a poisoned entry
- [x] Liberapay — ✅ — `liberapay.com/{name}` (200 taken / 404 free; a mixed-case slug 302s to its canonical case, which the redirect follow resolves)
- [x] Minehut — ✅ — `api.minehut.com/server/{name}?byName=true` (`server: null` = free)
- [x] Lichess — ✅ — `lichess.org/api/user/{name}`
- [x] Chess.com — ✅ — `api.chess.com/pub/player/{name}`
- [x] Hugging Face — ✅ — `huggingface.co/api/users/{name}/overview`
- [x] Dribbble — ✅ — `dribbble.com/{name}` (keyless profile: 200 taken / 404 free, but the route needs an HTML Accept — `application/json` 404s every name; a claimed profile that is not public also 404s, and its "isn't public yet" body marks it taken; the API v2 is publishing-only with no arbitrary-user endpoint, so the profile route is the signal)
- [x] Behance — ✅ — `behance.net/{name}` (keyless profile: 200 taken / 404 free; the Adobe API needs a key but the profile route does not)
- [x] Replicate — ✅ — `replicate.com/{name}` (keyless profile: 200 taken / 404 free; the API has no arbitrary-username lookup, so the profile page is the signal)
- [x] Companies House UK — ✅ — `api.company-information.service.gov.uk/search/companies?q={name}` (key; Basic auth; exact company-name hit taken / else free)
- [x] OpenCorporates — ✅ — `api.opencorporates.com/v0.4/companies/search?q={name}` (token; `X-API-TOKEN` header; exact company-name hit taken / else free)
- [x] Firefox Add-ons — ✅ — `addons.mozilla.org/api/v5`
- [x] SoundCloud — ✅ — `soundcloud.com/{name}` (keyless profile: 200 taken / 404 free)
- [x] Bandcamp — ✅ — `{name}.bandcamp.com` (a free name 303s to `bandcamp.com/signup?new_domain=`, a claimed one to its own content, so the redirect target is the verdict and must not be followed; the fan path `bandcamp.com/{name}` is not the signal — it 404s for artist-only accounts)
- [x] Gumroad — ✅ — `{name}.gumroad.com` (wildcard DNS; HTTP 404, no CORS; HTML Accept; user-chosen storefront names)
- [x] Vercel — ✅ — `{name}.vercel.app` (wildcard DNS; HTTP 404, no CORS)
- [x] Netlify — ✅ — `{name}.netlify.app` (wildcard DNS; HTTP 404, no CORS)
- [x] Render — ✅ — `{name}.onrender.com` (wildcard DNS; HTTP 404, no CORS; HTML Accept; spun-down free tier answers 503 → unknown)
- [x] Railway — ✅ — `{name}.up.railway.app` (wildcard DNS; HTTP 404, no CORS; HTML Accept; user-chosen names, collisions get a suffix so the exact label still 404s)
- [x] Heroku — ✅ — `{name}.herokuapp.com` (wildcard DNS; HTTP 404, no CORS; user-chosen names; paid-only since Nov 2022, so marked paid)
- [x] Surge — ✅ — `{name}.surge.sh` (wildcard DNS; HTTP 404, no CORS)
- [x] Firebase Hosting — ✅ — `{name}.web.app` (wildcard DNS; HTTP 404, no CORS)
- [x] Read the Docs — ✅ — `{name}.readthedocs.io` (wildcard DNS; HTTP 404, no CORS; a claimed slug 302s to the project docs, followed to 200; user-chosen project slugs)
- [x] Neocities — ✅ — `{name}.neocities.org` (wildcard DNS; HTTP 404, no CORS)
- [x] Shopify storefront — ✅ — `{name}.myshopify.com` (wildcard DNS; HTTP 404, HTML Accept; a claimed store 301s, followed to 200; user-chosen merchant names; distinct from the Shopify App Store)
- [x] Weebly — ✅ — `{name}.weebly.com` (wildcard DNS; HTTP 404, HTML Accept; user-chosen site names)
- [x] Open VSX — ✅ — `open-vsx.org/api/{namespace}` (namespace)
- [x] Dailymotion — ✅ — `api.dailymotion.com/user/{name}`
- [x] Lens Protocol — ✅ — `api.lens.xyz/graphql` (keyless GraphQL: POST an `account(username.localName:)` query; non-null `data.account` taken / null `data.account` free; page `hey.xyz/u/{name}`; undocumented keyless read access)
- [x] Paragraph — ✅ — `paragraph.com/@{name}` (keyless profile: 200 taken / real 404 free; `paragraph.xyz` 301s here)
- [x] Linktree — ✅ — `linktr.ee/{name}` (keyless profile: 200 taken / 404 free)
- [x] OpenSea — ✅ — `opensea.io/{name}` (keyless profile: 200 taken / 404 free; usernames are free to set with a wallet; Cloudflare-protected, so 403/429 degrades to unknown/uncached)
- [x] Spotify — ✅ — `open.spotify.com/user/{name}` (keyless profile: 200 taken / 404 free; the artist catalog still needs the keyed Web API)
- [x] Mixcloud — ✅ — `api.mixcloud.com/{name}/`
- [x] Discogs — ✅ — `api.discogs.com/users/{name}` (token; the keyless route 403s; 200 taken / 404 free)
- [x] AniList — ✅ — `graphql.anilist.co` (keyless GraphQL: POST a `User(name:)` query; non-null `data.User` taken / null `data.User` + "Not Found." error free)
- [x] Roblox — ✅ — `users.roblox.com/v1/usernames/users` (keyless: POST `{"usernames":["{name}"],"excludeBannedUsers":false}`; non-empty `data` array taken / empty free; page `www.roblox.com/users?username={name}`)
- [x] Minecraft (Mojang) — ✅ — `api.mojang.com/users/profiles/minecraft/{name}` (200 taken / 404 free; case-insensitive, 3–16 characters; 429 rate limits degrade to unknown, uncached; page `namemc.com/profile/{name}`)
- [x] Kaggle — ✅ — `kaggle.com/{name}` (keyless profile: 200 taken / 404 free)
- [x] skills.sh — ✅ — `skills.sh/{name}` (keyless publisher profile: 200 taken / 404 free; a name appears only once it publishes skills, and the app routes such as `/docs` and `/packs` answer 200, which reads as taken. This probes the GitHub-owner namespace only; publishers who serve skills from their own website live under the separate `/site/{host}` namespace, which is keyed by a hostname — dots, so not a bare label — and is not checked here)
- [x] eu.org — ✅ — `{name}.eu.org` (DNS, no wildcard: resolves = taken, NXDOMAIN = free; the label is sent to the resolver)
- [x] duckdns.org — ✅ — `{name}.duckdns.org` (DNS, no wildcard: resolves = taken, NXDOMAIN = free; the label is sent to the resolver)
- [x] no-ip.org — ✅ — `{name}.no-ip.org` (DNS, no wildcard: resolves = taken, NXDOMAIN = free; the label is sent to the resolver)
- [x] dynv6.net — ✅ — `{name}.dynv6.net` (DNS, no wildcard: resolves = taken, NXDOMAIN = free; the label is sent to the resolver)
- [x] noted.lol — ✅ — `{name}.noted.lol` (DNS, no wildcard: resolves = taken, NXDOMAIN = free; the label is sent to the resolver)
- [x] is-a-good.dev — ✅ — `{name}.is-a-good.dev` (DNS, no wildcard: resolves = taken, NXDOMAIN = free; the label is sent to the resolver)
- [x] pages.dev — ✅ — `{name}.pages.dev` (Cloudflare Pages; DNS, no wildcard: resolves = taken, NXDOMAIN = free; the label is sent to the resolver)
- [x] fly.dev — ✅ — `{name}.fly.dev` (Fly.io; DNS, no wildcard: resolves = taken, NXDOMAIN = free; the label is sent to the resolver)
- [x] Lemmy — ✅ — `lemmy.world/api/v3/user?username={name}` (200 = taken, 404 = free)
- [x] PeerTube — ✅ — `framatube.org/api/v1/accounts/{name}` and `/api/v1/video-channels/{name}` (accounts and channels share one actor namespace; either 200 = taken, both 404 = free)
- [x] Giphy — ✅ — `giphy.com/{name}` (keyless channel: 200 on the canonical `/{name}` taken / 301 to `/explore/{name}` free; status alone has no signal, so the redirect target or the `{Name} GIFs on GIPHY` title is the signal)
- [x] Minds — ✅ — `minds.com/api/v1/channel/{name}` (keyless JSON: `{"status":"success","channel":…}` taken / `{"status":"error","type":"ChannelNotFoundException"}` free; the HTML profile route at `minds.com/{name}/` is flaky — it sometimes serves the SPA shell for a claimed channel — so the API is the signal; page `minds.com/{name}/`)

## Candidates

- [ ] ccTLDs / new gTLDs — 🔌 — ruled out: not a distinct source — the IANA RDAP bootstrap only covers some TLDs, and any TLD it lists is already checked by the Domains provider
- [ ] Unstoppable Domains — 🔑 — deferred: the resolver API is keyed (401 unauthenticated)
- [ ] Packagist — 🕸 — composite `vendor/name`, no single-label namespace
- [ ] Maven Central — 🕸 — composite `group:artifact`, no single-label namespace
- [ ] Deno (`deno.land/x`) — 🕸 — `/x/{name}` 302-redirects and even taken names resolve 404, no signal
- [ ] Docker Hub — 🕸 — composite `user/repo`, no single-label namespace
- [ ] Flathub — 🔌 — ruled out: the app-id is reverse-DNS, not a single label, so a `{name}` probe never names a claimable app
- [ ] CPAN — 🕸 — deferred: module names are case-sensitive, and the input pipeline lowercases the label (`platforms.ts:14-18`, `serve.rs`), so `metacpan.org`/`fastapi.metacpan.org/v1/module/{name}` would false-available on a mixed-case module; needs a raw-case path through the client and engine first
- [ ] LuaRocks — 🕸 — composite `owner/name`, no single-label namespace
- [ ] Terraform Registry — 🕸 — scoped `namespace/type` path, not a single label
- [ ] VS Code Marketplace — 🕸 — no simple API (use Open VSX)
- [ ] JetBrains Marketplace — 🔌 — ruled out: `/api/searchPlugins?search=` is a fuzzy substring search, not an exact-name signal, and plugin ids are numeric
- [ ] Figma Community — 🕸 — no public API
- [ ] Shopify App Store — 🕸
- [ ] Nostr (NIP-05) — 🔌 — ruled out: names are per-domain (`https://{domain}/.well-known/nostr.json?name={name}`), not a global namespace, so no single-label source exists to probe
- [ ] Fandom — 🕸 — re-checked server-side (2026-11): `{name}.fandom.com` is wildcard DNS and every host answers 403 (Cloudflare) for real and fake names alike, so there is no signal
- [ ] Bilibili — 🕸 — API returns 412 (banned)
- [ ] Discord — 🕸 — usernames have no public API
- [ ] Telegram — 🕸 — keyless pass: `t.me/{name}` returns 200 for every name (JS/SPA shell), so no status signal; needs a key or a scrape — deferred, not ported
- [ ] Hashnode — 🕸 — the GraphQL host 301s (reads now paid) and `hashnode.com/{name}` is an always-200 SPA, so no signal
- [ ] Product Hunt — 🔑 — deferred: the GraphQL API is keyed
- [ ] LinkedIn — 🕸 — no public vanity-availability API; profile pages are auth-walled (999/redirect for both real and fake), so it cannot answer "is this name free"
- [ ] TikTok — 🕸 — keyless pass: every name returns 200 (JS/SPA shell), so no status signal; needs a key or a scrape — deferred, not ported
- [ ] Threads — 🕸 — keyless pass: 301 → (followed) 200 for both real and fake names, so no status signal; needs a key or a scrape — deferred, not ported
- [ ] Facebook — 🕸 — keyless pass: 200 (present) / 302 (missing) is a login redirect, not a name signal; needs a key or a scrape — deferred, not ported
- [ ] Reddit — 🕸 — keyless pass: 301 → (followed) 200 for both real and fake names, so no status signal (supersedes the old "403 from server IPs"); needs a key or a scrape — deferred, not ported
- [ ] Snapchat — 🕸 — keyless pass: the assumed URL shape 404s for both real and fake names, so no status signal; needs a key or a scrape — deferred, not ported
- [ ] Pinterest — 🕸 — keyless pass: every name returns 200 (JS/SPA shell), so no status signal; needs a key or a scrape — deferred, not ported
- [ ] Medium — 🕸 — keyless pass: 403 (bot-block) for both real and fake names, so no status signal; needs a key or a scrape — deferred, not ported
- [ ] XDA Forums — 🕸 — Cloudflare-blocked (403)
- [ ] USPTO (trademarks) — 🕸 — class-based, fuzzy-similarity trademark search; not a single claimable label, so it cannot answer name availability
- [ ] EUIPO (trademarks) — 🕸 — class-based trademark search (eSearch plus, OAuth); not a single claimable label, so it cannot answer name availability
- [ ] WIPO Global Brand Database — 🕸
- [ ] SEC EDGAR — 🕸 — keyless (descriptive User-Agent) but not a claimable namespace: it indexes issuer/filing names, which you cannot register
- [ ] Crunchbase — 🕸
- [ ] US state business registries — 🕸
- [ ] Chrome Web Store — 🔌 — ruled out: every name returns 302, so status alone has no signal
- [ ] Edge Add-ons — 🔌 — ruled out: every name returns 200, so status alone has no signal
- [ ] Apple App Store — 🕸
- [ ] Google Play — 🕸
- [x] Steam — ✅ — `steamcommunity.com/id/{name}?xml=1` (body signal: a taken custom id reports a `<steamID64>` element, a free one an `<error>` element; both are 200, so status alone has no signal)
- [ ] Epic Games — 🕸
- [ ] Patreon — 🕸 — bot-blocked: 403 for both real and fake names, so no signal
- [ ] Kickstarter — 🕸 — bot-blocked: 403 for both real and fake names, so no signal
- [ ] Indiegogo — 🕸
- [ ] Etsy — 🕸
- [ ] Amazon — 🕸
- [ ] eBay — 🕸
- [ ] IMDb — 🕸
- [ ] Azure Static Web Apps — 🔌 — `{name}.azurestaticapps.net` — the host is
  Azure-generated `{two-words}-{hash}.{region}.azurestaticapps.net`; Azure picks
  the phrase, so the name is **not user-choosable** — a single-label probe never
  corresponds to a claimable name.
- [ ] AWS Amplify — 🔌 — `{name}.amplifyapp.com` — host is
  `{branch}.{appid}.amplifyapp.com`, where `appid` is an AWS-generated id, not a
  chosen name; also non-wildcard (random names NXDOMAIN).
- [ ] Deno Deploy (Classic) — 🔌 — `{name}.deno.dev` was the preview URL of a
  Deploy Classic project, not an independently registrable name; Classic was
  shut down on 2026-07-20, `dash.deno.com` redirects to the new console, and the
  current platform uses wildcard `{app}.{org}.deno.net` hosts — composite, so
  there is no single claimable label to probe.
- [ ] Replit — 🔌 — `{name}.repl.co` — deprecated old domain; random NXDOMAIN and
  the apex 301s to `replit.com/deployments`.
- [ ] Glitch — 🔌 — `{name}.glitch.me` — Glitch is shut down; every name,
  including random ones, returns **410 Gone**.
- [ ] GitHub Pages — 🔌 — `{name}.github.io` — the subdomain *is* the GitHub
  account name, already covered by the GitHub provider; redundant.
- [ ] GitLab Pages — 🔌 — `{name}.gitlab.io` — the subdomain *is* the GitLab
  account name, already covered by the GitLab provider; redundant.
- [ ] Swift Package Index — 🕸 — 403
- [ ] Gradle Plugin Portal — 🕸 — no usable API
- [ ] Ansible Galaxy — 🕸 — scoped `namespace/name`, not a single label
- [ ] Puppet Forge — 🕸 — requires owner-name slug
- [ ] Elm packages — 🕸 — `author/project`, not a single label
- [ ] Quay (containers) — 🕸 — no public user endpoint
- [ ] Meteor — 🕸 — no API
- [ ] Nimble (Nim) — 🕸 — 502
- [ ] Crystal Shards — 🕸 — no endpoint
- [ ] npm users — 🕸 — `www.npmjs.com/~{name}` is a Cloudflare challenge (403 for
  both real and fake); `registry.npmjs.org/-/user/org.couchdb.user:{name}` is 401
  for both (auth required, no CORS); `/-/v1/search?text=maintainer:{name}` has CORS
  and does differ, but reflects published packages only — a real account with zero
  packages returns `total: 0`, so it can't prove the account exists.
- [ ] PyPI users — 🕸 — `pypi.org/user/{name}/` is HTML behind a bot challenge
  (200, byte-identical "Client Challenge" page for real and fake users) and sends
  no CORS; Warehouse exposes no JSON user endpoint.
- [ ] Stack Exchange — 🕸 — display names aren't unique (no availability)
- [ ] Go modules — 🕸 — modules are paths, not single labels
- [ ] WhatsApp — 🕸 — phone-based, no username
- [ ] Gmail — 🔌 — ruled out: the old keyless probe `mail.google.com/mail/gxlu?email={name}@gmail.com` now returns 204 with no `COMPASS` cookie for real and fake addresses alike (2026-11), and username availability only exists inside the bot-protected signup flow, which this repo won't scrape; dots and `+suffix` are normalized, so a literal local-part check would be wrong even if it worked
- [ ] Proton Mail — 🕸 — deferred: no public availability API. `account.proton.me/api/core/v4/users/available?Name={name}` needs an `x-pm-appversion` header plus an anonymous session (`POST /api/auth/v4/sessions`); with one, every name — including long random ones — returns 409 "Username already used", so taken/free is unverified. The endpoint is undocumented and version-gated, and the namespace is `local@domain` (proton.me/protonmail.com/pm.me/protonmail.ch), not a bare label
- [ ] Quora — 🕸
- [ ] Lobste.rs — 🕸
- [ ] Beehiiv — 🕸
- [ ] Ghost — 🕸
- [ ] WordPress.com — 🕸 — wildcard DNS
- [ ] Blogger — 🕸
- [ ] Tumblr — 🔑
- [ ] Vimeo — 🕸 — bot-blocked: 403 for both real and fake names, so no signal
- [ ] Odysee / LBRY — 🕸 — no usable API
- [ ] VK — 🔑
- [ ] Weibo — 🕸
- [ ] Zhihu — 🕸
- [ ] Douyin — 🕸
- [ ] Line — 🔑
- [ ] Kakao — 🔑
- [ ] Matrix — 🕸 — inconclusive: availability is a per-homeserver localpart, and the register/available endpoint only answers on open-registration servers (see plan/results/P2.md)
- [ ] Revolt — 🕸 — the profile route 301s, so no status signal
- [ ] Guilded — 🕸
- [ ] Slack — 🕸
- [ ] DeviantArt — 🕸 — profile pages are an always-200 SPA (any name returns 200), so no signal
- [ ] OpenNIC — 🔌 — ruled out: the DoH endpoints are dead
- [ ] Ko-fi — 🕸 — redirects, no signal
- [ ] Buy Me a Coffee — 🕸 — re-checked server-side (2026-11): `buymeacoffee.com/{name}` 404s for real and fake names alike, so there is no signal
- [ ] Lemon Squeezy — 🕸 — bot-blocked: 403 for both real and fake names, so no signal
- [ ] thanks.dev — 🕸 — Cloudflare-blocked (403)
- [ ] PayPal.me — 🕸 — 301 redirect / always 200, no signal
- [ ] Itch.io — 🕸
- [ ] Spotify (artist) — 🔑
- [ ] Apple Music — 🔑
- [ ] Audiomack — 🕸 — redirects
- [ ] Flickr — 🕸 — endpoints inconclusive
- [ ] Last.fm — 🔑 — the keyless profile route now bot-challenges (`last.fm/user/{name}` returns Cloudflare "Client Challenge" with 200 for every name), so it would need the keyed API (`ws.audioscrobbler.com/2.0/?method=user.getinfo&user={name}&api_key={key}`); deferred
- [ ] Letterboxd — 🕸
- [ ] Goodreads — 🕸
- [ ] MyAnimeList — 🕸 — Jikan API unavailable
- [ ] Trakt — 🔑
- [ ] Genius — 🔑
- [ ] Strava — 🔑
- [ ] Duolingo — 🕸
- [ ] CodeSandbox — 🕸 — Cloudflare-blocked (403)
- [ ] StackBlitz — 🕸
- [ ] CodePen — 🕸
- [ ] Observable — 🕸
- [ ] Wellfound (AngelList) — 🕸
- [ ] HelloWork — 🕸 — job board; no public username namespace/API
- [ ] Fiverr — 🕸
- [ ] Upwork — 🕸
- [ ] Polywork — 🕸
- [ ] freedns.afraid.org — 🔌 — ruled out: FreeDNS hostnames live under shared
  domains (`{name}.{shared-domain}`), not `{name}.freedns.afraid.org`, so a
  single-label probe never names a claimable host.
- [ ] js.org — 🔌 — ruled out: wildcard DNS — every `{name}.js.org` resolves and
  the host answers 200/302 for both real and fake names, so neither DNS nor HTTP
  status isolates a claim.
- [ ] is-a.dev — 🔌 — ruled out: wildcard DNS — every `{name}.is-a.dev` resolves
  and the host answers 200/302 for both real and fake names, so neither DNS nor
  HTTP status isolates a claim.
- [ ] ngrok-free.app — 🔌 — ruled out: ephemeral / no stable namespace — tunnels
  get an agent-assigned host, not a user-chosen name, so there is nothing
  claimable to probe.
- [ ] localhost.run — 🔌 — ruled out: ephemeral — each tunnel gets a random
  generated subdomain per session, so the name never resolves.
- [ ] trycloudflare.com — 🔌 — ruled out: ephemeral — Quick Tunnels get a random
  generated host per session; the name never resolves to a claimable namespace.
- [ ] Notion — 🕸 — `{name}.notion.site` — keyless pass: 200 for both a real and a
  fake page, so no status signal; deferred, not ported
- [ ] Imgur — 🕸 — `imgur.com/user/{name}` — keyless pass: an always-200 SPA shell
  for both real and fake users, so no status signal; deferred, not ported
- [ ] Wix — 🕸 — `{name}.wixsite.com` — keyless pass: 404 for both real and fake
  sites, so no status signal
- [ ] Squarespace — 🕸 — `{name}.squarespace.com` — keyless pass: 401 for a claimed
  site / 404 for a free name — ambiguous, so no reliable signal
- [ ] ChatGPT profiles — 🕸 — `chatgpt.com/u/{name}` — keyless pass: signed-out
  requests 404 for both real and fake usernames (profiles require sign-in, and a
  username can be taken with no visible profile); the `chatgpt.com/@{name}` form
  does not exist, and `*.chatgpt.com` has no wildcard DNS — deferred, not ported
- [ ] ChatGPT Sites — 🔌 — ruled out: public Sites live at
  `{site}.{username-or-workspace}.chatgpt.site`, not `{name}.chatgpt.site` — a
  single label never names a claimable namespace (`uchz.chatgpt.site` and
  `openai.chatgpt.site` both 404), personal default namespaces are random ids,
  and a taken private Site answers 401
- [ ] AI chat share links (Claude, Qwen, DeepSeek, Mistral) — 🔌 — ruled out:
  public URLs are opaque ids, not names — `claude.ai/share/<uuid>` and
  `claude.ai/public/artifacts/<uuid>`, `chat.deepseek.com/share/<token>`, and
  equivalent Qwen/Vibe links; none has a public `/u/{name}` or `@<name>` route,
  and Mistral's help says share links never include the username
- [ ] Perplexity — 🕸 — `perplexity.ai/@username` is the only reported public
  profile route, but Cloudflare 403s every keyless request (real and fake alike,
  including `/@`, `/u/` and `/user/`), so there is no signal; the actual share
  URLs are `search/<slug>-<id>` and `page/<slug>-<id>`, where the id — not the
  slug — is the unique key; deferred, not ported
- [ ] Bloomberg — 🔌 — ruled out: `bloomberg.com/profile/company/{id}` is keyed
  by a Bloomberg company id/ticker (`1654159D:US`), not a claimable username/org
  label; the page is also bot-blocked (403 "Are you a robot?" for both real and
  fake ids)
- [ ] NewsBreak — 🔌 — ruled out: publisher pages are
  `newsbreak.com/m/{slug}-{id}`, where the numeric id is the key —
  `…/m/zzqx9-560145` returns HackerNoon's publisher page — so the slug is
  cosmetic and there is no username/org namespace
- [ ] ViewBlock — 🔌 — ruled out: `viewblock.io/arweave/address/{address}` is
  keyed by a base58 wallet address, not a claimable name; the HackerNoon footer
  link is a wallet, not a username
- [ ] Trustpilot — 🕸 — ruled out: profiles are business units keyed by domain
  (`trustpilot.com/review/{domain}`), not a claimable name — a 200 means "a
  profile exists for that domain" (including units with no reviews) and a 404
  says nothing about availability, so neither maps to taken/free. The Business
  Units API is keyed (401 without `apikey`), `robots.txt` is `User-agent: *` /
  `Disallow: /`, and keyless requests get a Cloudflare challenge (403), so there
  is no verifiable signal
