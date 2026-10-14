# Platform backlog

Candidate sources for checking whether a name is taken. Flat list — one entry per
line.

**Rule:** nothing gets added until its signal is verified **server-side** on both
the taken and free response (see the README's "Adding a source"). CORS is no
longer a constraint — checks run from the server — so the reachability notes
below, written in the browser-direct era, must be re-checked before implementing.

Legend — reachability with checks running from the server (the server is the
proxy, so CORS no longer gates anything):

- ✅ verified working
- 🌐 public API, likely works — verify server-side
- 🔌 reachable server-side by HTTP status (no key) — formerly "needs a proxy"
- 🔑 needs an API key / config
- 🕸 needs an HTML/body signal — some in this group are also bot- or TOS-blocked

Status: `[x]` = already a provider, `[ ]` = candidate.

- [x] Domains (RDAP TLDs) — ✅ — `.com .org .dev .io .ai .app` (`.io` via fallback) + any TLD in the IANA bootstrap
- [ ] ccTLDs / new gTLDs — 🌐 — RDAP coverage varies; not all are in the bootstrap
- [x] ENS (`.eth`) — ✅ — `api.ensideas.com/ens/resolve/{name}.eth` (200 always; non-null `address` = taken, `null` = free; third-party resolver — a follow-up could query a public Ethereum RPC against the ENS registry directly)
- [ ] Handshake (HNS) — 🌐 — HNS resolver
- [ ] Unstoppable Domains — 🌐 — resolver API
- [ ] Solana Name Service — 🌐 — SNS API
- [x] npm — ✅ — `registry.npmjs.org/{name}`
- [x] PyPI — ✅ — `pypi.org/pypi/{name}/json`
- [x] crates.io — ✅ — `crates.io/api/v1/crates/{name}`
- [x] NuGet — ✅ — `api.nuget.org/v3-flatcontainer/{name}/index.json`
- [x] RubyGems — ✅ — `rubygems.org/api/v1/gems/{name}.json`
- [ ] Packagist — 🕸 — composite `vendor/name`, no single-label namespace
- [ ] Maven Central — 🕸 — composite `group:artifact`, no single-label namespace
- [x] Hex (Elixir) — ✅ — `hex.pm/api/packages/{name}` (packages) and `hex.pm/api/users/{name}` (users)
- [x] pub.dev (Dart) — ✅ — `pub.dev/api/packages/{name}`
- [x] JSR (JavaScript) — ✅ — `jsr.io/api/scopes/{name}`
- [ ] Deno (`deno.land/x`) — 🕸 — `/x/{name}` 302-redirects and even taken names resolve 404, no signal
- [ ] Docker Hub — 🕸 — composite `user/repo`, no single-label namespace
- [x] Homebrew — ✅ — `formulae.brew.sh/api/formula/{name}.json`
- [x] Chocolatey — ✅ — `community.chocolatey.org/packages/{name}` (status-only HTML; package ids are case-insensitive)
- [ ] Snap Store — 🕸 — viable with the `Snap-Device-Series: 16` header; deferred until the `endpoint` helper supports a custom header
- [ ] Flathub — 🌐 — app-id is reverse-DNS, not a label
- [x] AUR (Arch) — ✅ — `aur.archlinux.org/packages/{name}` (status-only HTML; the RPC endpoint always answers 200)
- [ ] CPAN — 🕸 — case-sensitive CamelCase names; labels are lowercased (`platforms.ts:14-18`), so it would false-available
- [x] Hackage — ✅ — `hackage.haskell.org/package/{name}` (bare name resolves; status-only HTML)
- [x] Clojars — ✅ — `clojars.org/api/artifacts/{name}` (single segment resolves)
- [ ] LuaRocks — 🕸 — composite `owner/name`, no single-label namespace
- [ ] Terraform Registry — 🕸 — scoped path + no CORS
- [ ] VS Code Marketplace — 🕸 — no simple API (use Open VSX)
- [ ] JetBrains Marketplace — 🌐 — search API only (body signal)
- [x] WordPress plugins/themes — ✅ — `api.wordpress.org/plugins/info/1.2`
- [ ] Figma Community — 🕸 — no public API
- [ ] Shopify App Store — 🕸
- [x] GitHub — ✅ — `api.github.com/users/{name}`
- [x] Bluesky — ✅ — `bsky.social/xrpc/com.atproto.temp.checkHandleAvailability` (`{name}.bsky.social`); resolveHandle hides reserved/deactivated/over-length (max 18) handles
- [x] Mastodon (`mastodon.social`) — ✅ — `/@{name}` profile (200/410 taken, 404 free; the lookup API hides suspended accounts as 404)
- [x] GitLab — ✅ — `gitlab.com/api/v4/users?username={name}` (array body), then `/api/v4/groups/{name}` as a fallback
- [x] Bitbucket — ✅ — `api.bitbucket.org/2.0/workspaces/{name}`
- [x] Codeberg — ✅ — `codeberg.org/api/v1/users/{name}`
- [ ] Nostr (NIP-05) — 🌐 — `/.well-known/nostr.json`
- [x] Farcaster — ✅ — `fnames.farcaster.xyz/transfers/current?name={name}`
- [ ] Lemmy — 🌐 — `/api/v3/user?username={name}`
- [x] DEV — ✅ — `dev.to/api/users/by_username?url={name}`, then `/api/organizations/{name}` as a fallback
- [x] Wikipedia — ✅ — `meta.wikimedia.org/w/api.php` globaluserinfo (`missing` = free)
- [ ] Fandom — 🕸 — `{name}.fandom.com` has no CORS
- [ ] Bilibili — 🕸 — API returns 412 (banned)
- [ ] Discord — 🕸 — usernames have no public API
- [ ] Telegram — 🕸 — `t.me/{name}` scrape
- [ ] Hashnode — 🌐 — GraphQL
- [ ] Substack — 🌐 — `substack.com/@{name}` (scrape)
- [ ] Product Hunt — 🌐 — GraphQL (key)
- [x] X / Twitter — ✅ — `api.twitter.com/2/users/by/username/{name}` (App-Only bearer; 200 `data` taken / 404 or no `data` free)
- [x] YouTube — ✅ — `youtube.com/@{name}` (profile redirects and resolves 200 taken / 404 free; keyless, no Data API key needed)
- [x] Twitch — ✅ — `api.twitch.tv/helix/users?login={name}` (Client-Id + app access token via client-credentials grant; non-empty `data` taken / empty free)
- [ ] LinkedIn — 🕸 — no public vanity-availability API; profile pages are auth-walled (999/redirect for both real and fake), so it cannot answer "is this name free"
- [ ] Instagram — 🕸
- [ ] TikTok — 🕸
- [ ] Threads — 🕸
- [ ] Facebook — 🕸
- [ ] Reddit — 🕸 — 403 from server IPs
- [ ] Snapchat — 🕸
- [ ] Pinterest — 🕸
- [ ] Medium — 🕸
- [x] Hacker News — ✅ — `news.ycombinator.com/user?id={name}` (body signal: `No such user.` = free, `karma` = taken; the firebase JSON API only exposes users with public activity, so `null` can't confirm free)
- [x] HackerNoon — ✅ — `hackernoon.com/u/{name}` (keyless profile: 200 taken / 404 free)
- [x] Open Collective — ✅ — `opencollective.com/{name}.json` (200 taken / 404 free). The CDN caches responses for 1h and can cache a transient 5xx per-`Origin`; the provider retries once with a cache-busting param to bypass a poisoned entry
- [x] Minehut — ✅ — `api.minehut.com/server/{name}?byName=true` (`server: null` = free)
- [x] Lichess — ✅ — `lichess.org/api/user/{name}`
- [x] Chess.com — ✅ — `api.chess.com/pub/player/{name}`
- [x] Hugging Face — ✅ — `huggingface.co/api/users/{name}/overview`
- [ ] Minecraft (Mojang) — 🕸 — API has no CORS
- [ ] Roblox — 🕸 — rate-limited / no CORS
- [ ] SourceForge — 🕸 — 200 has CORS, 404 does not
- [x] Dribbble — ✅ — `dribbble.com/{name}` (keyless profile: 200 taken / 404 free; the API v2 is publishing-only with no arbitrary-user endpoint, so the profile route is the signal)
- [x] Behance — ✅ — `behance.net/{name}` (keyless profile: 200 taken / 404 free; the Adobe API needs a key but the profile route does not)
- [x] Replicate — ✅ — `replicate.com/{name}` (keyless profile: 200 taken / 404 free; the API has no arbitrary-username lookup, so the profile page is the signal)
- [ ] XDA Forums — 🕸 — Cloudflare-blocked (403)
- [ ] USPTO (trademarks) — 🕸 — class-based, fuzzy-similarity trademark search; not a single claimable label, so it cannot answer name availability
- [ ] EUIPO (trademarks) — 🕸 — class-based trademark search (eSearch plus, OAuth); not a single claimable label, so it cannot answer name availability
- [ ] WIPO Global Brand Database — 🕸
- [x] Companies House UK — ✅ — `api.company-information.service.gov.uk/search/companies?q={name}` (key; Basic auth; exact company-name hit taken / else free)
- [x] OpenCorporates — ✅ — `api.opencorporates.com/v0.4/companies/search?q={name}` (token; `X-API-TOKEN` header; exact company-name hit taken / else free)
- [ ] SEC EDGAR — 🕸 — keyless (descriptive User-Agent) but not a claimable namespace: it indexes issuer/filing names, which you cannot register
- [ ] Crunchbase — 🕸
- [ ] US state business registries — 🕸
- [ ] Chrome Web Store — 🌐
- [x] Firefox Add-ons — ✅ — `addons.mozilla.org/api/v5`
- [ ] Edge Add-ons — 🌐
- [ ] Apple App Store — 🕸
- [ ] Google Play — 🕸
- [ ] Steam — 🕸
- [ ] Epic Games — 🕸
- [ ] Patreon — 🌐
- [ ] Bandcamp — 🌐
- [x] SoundCloud — ✅ — `soundcloud.com/{name}` (keyless profile: 200 taken / 404 free)
- [x] Gumroad — ✅ — `{name}.gumroad.com` (wildcard DNS; HTTP 404, no CORS; HTML Accept; user-chosen storefront names)
- [ ] Kickstarter — 🌐
- [ ] Indiegogo — 🕸
- [ ] Etsy — 🕸
- [ ] Amazon — 🕸
- [ ] eBay — 🕸
- [ ] IMDb — 🕸

Hosting subdomains (server-side HTTP status is the signal; wildcard DNS defeats
browser DoH):

- [x] Vercel — ✅ — `{name}.vercel.app` (wildcard DNS; HTTP 404, no CORS)
- [x] Netlify — ✅ — `{name}.netlify.app` (wildcard DNS; HTTP 404, no CORS)
- [x] Render — ✅ — `{name}.onrender.com` (wildcard DNS; HTTP 404, no CORS; HTML Accept; spun-down free tier answers 503 → unknown)
- [x] Railway — ✅ — `{name}.up.railway.app` (wildcard DNS; HTTP 404, no CORS; HTML Accept; user-chosen names, collisions get a suffix so the exact label still 404s)
- [x] Heroku — ✅ — `{name}.herokuapp.com` (wildcard DNS; HTTP 404, no CORS; user-chosen names; paid-only since Nov 2022, so marked paid)
- [x] Surge — ✅ — `{name}.surge.sh` (wildcard DNS; HTTP 404, no CORS)
- [x] Deno Deploy — ✅ — `{name}.deno.dev` (wildcard DNS; HTTP 404, no CORS; a claimed name can 301 to a custom domain, followed to 200; user-chosen project names)
- [x] Firebase Hosting — ✅ — `{name}.web.app` (wildcard DNS; HTTP 404, no CORS)
- [x] Read the Docs — ✅ — `{name}.readthedocs.io` (wildcard DNS; HTTP 404, no CORS; a claimed slug 302s to the project docs, followed to 200; user-chosen project slugs)
- [x] Neocities — ✅ — `{name}.neocities.org` (wildcard DNS; HTTP 404, no CORS)

Hosting subdomains — ruled out (status can't isolate a user-chosen name):

- [ ] Azure Static Web Apps — 🔌 — `{name}.azurestaticapps.net` — the host is
  Azure-generated `{two-words}-{hash}.{region}.azurestaticapps.net`; Azure picks
  the phrase, so the name is **not user-choosable** — a single-label probe never
  corresponds to a claimable name.
- [ ] AWS Amplify — 🔌 — `{name}.amplifyapp.com` — host is
  `{branch}.{appid}.amplifyapp.com`, where `appid` is an AWS-generated id, not a
  chosen name; also non-wildcard (random names NXDOMAIN).
- [ ] Replit — 🔌 — `{name}.repl.co` — deprecated old domain; random NXDOMAIN and
  the apex 301s to `replit.com/deployments`.
- [ ] Glitch — 🔌 — `{name}.glitch.me` — Glitch is shut down; every name,
  including random ones, returns **410 Gone**.
- [ ] GitHub Pages — 🔌 — `{name}.github.io` — the subdomain *is* the GitHub
  account name, already covered by the GitHub provider; redundant.
- [ ] GitLab Pages — 🔌 — `{name}.gitlab.io` — the subdomain *is* the GitLab
  account name, already covered by the GitLab provider; redundant.

More package registries:

- [x] Open VSX — ✅ — `open-vsx.org/api/{namespace}` (namespace)
- [ ] CocoaPods — 🕸 — case-sensitive CamelCase names; labels are lowercased (`platforms.ts:14-18`), so it would false-available
- [ ] Swift Package Index — 🕸 — 403
- [ ] Gradle Plugin Portal — 🕸 — no usable API
- [ ] Ansible Galaxy — 🕸 — scoped / no CORS
- [ ] Puppet Forge — 🕸 — requires owner-name slug
- [ ] Chef Supermarket — 🕸 — no CORS
- [ ] Elm packages — 🕸 — no CORS
- [ ] Quay (containers) — 🕸 — no public user endpoint
- [ ] Meteor — 🕸 — no API
- [ ] Nimble (Nim) — 🕸 — 502
- [ ] Crystal Shards — 🕸 — no endpoint

Ruled out — account namespaces (no public availability API):

- [ ] npm users — 🕸 — `www.npmjs.com/~{name}` is a Cloudflare challenge (403 for
  both real and fake); `registry.npmjs.org/-/user/org.couchdb.user:{name}` is 401
  for both (auth required, no CORS); `/-/v1/search?text=maintainer:{name}` has CORS
  and does differ, but reflects published packages only — a real account with zero
  packages returns `total: 0`, so it can't prove the account exists.
- [ ] PyPI users — 🕸 — `pypi.org/user/{name}/` is HTML behind a bot challenge
  (200, byte-identical "Client Challenge" page for real and fake users) and sends
  no CORS; Warehouse exposes no JSON user endpoint.

Ruled out — can't answer "is the name free" (semantic, not reachability; the
server doesn't change this):

- [ ] Stack Exchange — 🕸 — display names aren't unique (no availability)
- [ ] Go modules — 🕸 — modules are paths, not single labels
- [ ] WhatsApp — 🕸 — phone-based, no username

More social / community:

- [ ] Quora — 🕸
- [ ] Lobste.rs — 🕸
- [ ] HackerNoon — 🌐
- [ ] Beehiiv — 🕸
- [ ] Ghost — 🕸
- [ ] WordPress.com — 🕸 — wildcard DNS
- [ ] Blogger — 🕸
- [ ] Tumblr — 🔑
- [ ] Vimeo — 🔑
- [x] Dailymotion — ✅ — `api.dailymotion.com/user/{name}`
- [ ] Odysee / LBRY — 🕸 — no usable API
- [ ] PeerTube — 🌐 — per-instance
- [ ] VK — 🔑
- [ ] Weibo — 🕸
- [ ] Zhihu — 🕸
- [ ] Douyin — 🕸
- [ ] Line — 🔑
- [ ] Kakao — 🔑
- [ ] Matrix — 🌐 — per-homeserver
- [ ] Revolt — 🌐
- [ ] Guilded — 🕸
- [ ] Slack — 🕸

Name systems / crypto:

- [ ] Namecoin (`.bit`) — 🌐 — name lookup
- [ ] OpenNIC — 🌐
- [ ] Lens Protocol — 🌐
- [ ] Paragraph — 🌐

Creators / commerce / media:

- [ ] Ko-fi — 🕸 — redirects, no signal
- [ ] Buy Me a Coffee — 🕸 — no CORS
- [ ] Lemon Squeezy — 🌐
- [ ] thanks.dev — 🕸 — Cloudflare-blocked (403)
- [ ] Liberapay — 🕸 — 404 lacks CORS
- [ ] PayPal.me — 🕸 — 301 redirect / always 200, no signal
- [ ] Itch.io — 🕸
- [ ] Spotify (artist) — 🔑
- [ ] Apple Music — 🔑
- [ ] Audiomack — 🕸 — redirects
- [x] Mixcloud — ✅ — `api.mixcloud.com/{name}/`
- [x] Discogs — ✅ — `api.discogs.com/users/{name}` (token; the keyless route 403s; 200 taken / 404 free)
- [x] Last.fm — ✅ — `last.fm/user/{name}` (keyless profile: 200 taken / 404 free)
- [ ] Letterboxd — 🕸
- [ ] Goodreads — 🕸
- [ ] MyAnimeList — 🕸 — Jikan API unavailable
- [ ] AniList — 🌐 — GraphQL
- [ ] Trakt — 🔑
- [ ] Genius — 🔑
- [ ] Strava — 🔑
- [ ] Duolingo — 🕸

AI / dev tools:

- [x] Kaggle — ✅ — `kaggle.com/{name}` (keyless profile: 200 taken / 404 free)
- [ ] CodeSandbox — 🕸 — Cloudflare-blocked (403)
- [ ] StackBlitz — 🕸
- [ ] CodePen — 🕸
- [ ] Observable — 🕸

Jobs:

- [ ] Wellfound (AngelList) — 🕸
- [ ] HelloWork — 🕸 — job board; no public username namespace/API
- [ ] Fiverr — 🕸
- [ ] Upwork — 🕸
- [ ] Polywork — 🕸

Free subdomain namespaces (developer-oriented):

- [x] eu.org — ✅ — `{name}.eu.org` (DNS, no wildcard: resolves = taken, NXDOMAIN = free; the label is sent to the resolver)
- [x] duckdns.org — ✅ — `{name}.duckdns.org` (DNS, no wildcard: resolves = taken, NXDOMAIN = free; the label is sent to the resolver)
- [ ] freedns.afraid.org — 🌐 — ruled out: FreeDNS hostnames live under shared
  domains (`{name}.{shared-domain}`), not `{name}.freedns.afraid.org`, so a
  single-label probe never names a claimable host.
- [x] no-ip.org — ✅ — `{name}.no-ip.org` (DNS, no wildcard: resolves = taken, NXDOMAIN = free; the label is sent to the resolver)
- [x] dynv6.net — ✅ — `{name}.dynv6.net` (DNS, no wildcard: resolves = taken, NXDOMAIN = free; the label is sent to the resolver)
- [x] noted.lol — ✅ — `{name}.noted.lol` (DNS, no wildcard: resolves = taken, NXDOMAIN = free; the label is sent to the resolver)
- [ ] js.org — 🔌 — ruled out: wildcard DNS — every `{name}.js.org` resolves and
  the host answers 200/302 for both real and fake names, so neither DNS nor HTTP
  status isolates a claim.
- [ ] is-a.dev — 🔌 — ruled out: wildcard DNS — every `{name}.is-a.dev` resolves
  and the host answers 200/302 for both real and fake names, so neither DNS nor
  HTTP status isolates a claim.
- [x] is-a-good.dev — ✅ — `{name}.is-a-good.dev` (DNS, no wildcard: resolves = taken, NXDOMAIN = free; the label is sent to the resolver)
- [x] pages.dev — ✅ — `{name}.pages.dev` (Cloudflare Pages; DNS, no wildcard: resolves = taken, NXDOMAIN = free; the label is sent to the resolver)
- [x] fly.dev — ✅ — `{name}.fly.dev` (Fly.io; DNS, no wildcard: resolves = taken, NXDOMAIN = free; the label is sent to the resolver)
- [ ] ngrok-free.app — 🔌 — ruled out: ephemeral / no stable namespace — tunnels
  get an agent-assigned host, not a user-chosen name, so there is nothing
  claimable to probe.
- [ ] localhost.run — 🔌 — ruled out: ephemeral — each tunnel gets a random
  generated subdomain per session, so the name never resolves.
- [ ] trycloudflare.com — 🔌 — ruled out: ephemeral — Quick Tunnels get a random
  generated host per session; the name never resolves to a claimable namespace.
