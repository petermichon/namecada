# Platform backlog

Candidate sources for checking whether a name is taken. Flat list — one entry per
line.

**Rule:** nothing gets added to the app until its CORS is verified on **both** the
taken and free response (see the README's "Adding a source"). Reachability below
is best-effort and must be re-checked before implementing.

Legend:

- ✅ verified browser-direct (CORS tested)
- 🌐 likely browser-direct (public CORS API — verify)
- 🔌 needs a proxy (no key) — CORS/DNS-blocked but otherwise reachable
- 🔑 needs an API key and a proxy
- 🕸 no public API / scrape / TOS-blocked

Status: `[x]` = already a provider, `[ ]` = candidate.

- [x] Domains (RDAP TLDs) — ✅ — `.com .org .dev .io .ai .app` (`.io` via fallback) + any TLD in the IANA bootstrap
- [ ] ccTLDs / new gTLDs — 🌐 — RDAP coverage varies; not all are in the bootstrap
- [ ] ENS (`.eth`) — 🌐 — public RPC / resolver
- [ ] Handshake (HNS) — 🌐 — HNS resolver
- [ ] Unstoppable Domains — 🌐 — resolver API
- [ ] Solana Name Service — 🌐 — SNS API
- [x] npm — ✅ — `registry.npmjs.org/{name}`
- [x] PyPI — ✅ — `pypi.org/pypi/{name}/json`
- [x] crates.io — ✅ — `crates.io/api/v1/crates/{name}`
- [x] NuGet — ✅ — `api.nuget.org/v3-flatcontainer/{name}/index.json`
- [ ] RubyGems — 🕸 — no CORS on 404
- [ ] Packagist — 🕸 — no CORS
- [ ] Maven Central — 🕸 — no CORS
- [x] Hex (Elixir) — ✅ — `hex.pm/api/packages/{name}`
- [x] pub.dev (Dart) — ✅ — `pub.dev/api/packages/{name}`
- [x] JSR (JavaScript) — ✅ — `jsr.io/api/scopes/{name}`
- [ ] Deno (`deno.land/x`) — 🌐 — `deno.land/x/{name}`
- [ ] Docker Hub — 🕸 — no CORS
- [x] Homebrew — ✅ — `formulae.brew.sh/api/formula/{name}.json`
- [ ] Chocolatey — 🕸 — redirects / no CORS
- [ ] Snap Store — 🕸 — no CORS + required header
- [ ] Flathub — 🌐 — app-id is reverse-DNS, not a label
- [ ] AUR (Arch) — 🕸 — no CORS
- [ ] CPAN — 🕸 — blocked (402)
- [ ] Hackage — 🕸 — no CORS
- [ ] Clojars — 🕸 — no CORS on 404
- [ ] LuaRocks — 🌐 — `luarocks.org` API
- [ ] Terraform Registry — 🕸 — scoped path + no CORS
- [ ] VS Code Marketplace — 🕸 — no simple API (use Open VSX)
- [ ] JetBrains Marketplace — 🌐 — search API only (body signal)
- [x] WordPress plugins/themes — ✅ — `api.wordpress.org/plugins/info/1.2`
- [ ] Figma Community — 🕸 — no public API
- [ ] Shopify App Store — 🕸
- [ ] Go modules — 🕸 — modules are paths, not single labels
- [x] GitHub — ✅ — `api.github.com/users/{name}`
- [x] Bluesky — ✅ — `public.api.bsky.app` resolveHandle (`{name}.bsky.social`)
- [x] Mastodon (`mastodon.social`) — ✅ — `/api/v1/accounts/lookup?acct={name}`
- [x] GitLab — ✅ — `gitlab.com/api/v4/users?username={name}` (array body)
- [x] Bitbucket — ✅ — `api.bitbucket.org/2.0/workspaces/{name}`
- [x] Codeberg — ✅ — `codeberg.org/api/v1/users/{name}`
- [ ] Nostr (NIP-05) — 🌐 — `/.well-known/nostr.json`
- [x] Farcaster — ✅ — `fnames.farcaster.xyz/transfers/current?name={name}`
- [ ] Lemmy — 🌐 — `/api/v3/user?username={name}`
- [x] DEV — ✅ — `dev.to/api/users/by_username?url={name}`
- [x] Wikipedia — ✅ — `en.wikipedia.org/w/api.php` list=users (`missing` = free)
- [ ] Fandom — 🕸 — `{name}.fandom.com` has no CORS
- [ ] Bilibili — 🕸 — API returns 412 (banned)
- [ ] Discord — 🕸 — usernames have no public API
- [ ] Telegram — 🕸 — `t.me/{name}` scrape
- [ ] Hashnode — 🌐 — GraphQL
- [ ] Substack — 🌐 — `substack.com/@{name}` (scrape)
- [ ] Product Hunt — 🌐 — GraphQL (key)
- [ ] X / Twitter — 🔑
- [ ] YouTube — 🔑 — Data API
- [ ] Twitch — 🔑
- [ ] LinkedIn — 🔑
- [ ] Instagram — 🕸
- [ ] TikTok — 🕸
- [ ] Threads — 🕸
- [ ] Facebook — 🕸
- [ ] Reddit — 🕸 — 403 from server IPs
- [ ] Snapchat — 🕸
- [ ] Pinterest — 🕸
- [ ] Medium — 🕸
- [x] Hacker News — ✅ — `hacker-news.firebaseio.com/v0/user/{name}.json` (`null` = free)
- [x] Open Collective — ✅ — `opencollective.com/{name}.json`
- [x] Minehut — ✅ — `api.minehut.com/server/{name}?byName=true` (`server: null` = free)
- [x] Lichess — ✅ — `lichess.org/api/user/{name}`
- [x] Chess.com — ✅ — `api.chess.com/pub/player/{name}`
- [x] Hugging Face — ✅ — `huggingface.co/api/users/{name}/overview`
- [ ] Minecraft (Mojang) — 🕸 — API has no CORS
- [ ] Roblox — 🕸 — rate-limited / no CORS
- [ ] Stack Exchange — 🕸 — display names aren't unique (no availability)
- [ ] SourceForge — 🕸 — 200 has CORS, 404 does not
- [ ] Dribbble — 🔑 — API needs a token
- [ ] Behance — 🔑 — Adobe API needs a key
- [ ] XDA Forums — 🕸 — Cloudflare-blocked (403)
- [ ] WhatsApp — 🕸 — phone-based, no username
- [ ] USPTO (trademarks) — 🔑
- [ ] EUIPO (trademarks) — 🔑
- [ ] WIPO Global Brand Database — 🕸
- [ ] Companies House UK — 🔑
- [ ] OpenCorporates — 🔑
- [ ] SEC EDGAR — 🔑
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
- [ ] SoundCloud — 🌐
- [ ] Gumroad — 🌐
- [ ] Kickstarter — 🌐
- [ ] Indiegogo — 🕸
- [ ] Etsy — 🕸
- [ ] Amazon — 🕸
- [ ] eBay — 🕸
- [ ] IMDb — 🕸

Hosting subdomains (need a proxy to read the 404; wildcard DNS defeats DoH):

- [ ] Vercel — 🔌 — `{name}.vercel.app` (wildcard DNS; HTTP 404, no CORS)
- [ ] GitHub Pages — 🔌 — `{name}.github.io` (wildcard DNS; no CORS)
- [ ] Netlify — 🔌 — `{name}.netlify.app` (wildcard DNS; no CORS)
- [ ] Cloudflare Pages — 🔌 — `{name}.pages.dev` (not wildcard; DoH may work)
- [ ] GitLab Pages — 🔌 — `{name}.gitlab.io`
- [ ] Render — 🔌 — `{name}.onrender.com`
- [ ] Fly.io — 🔌 — `{name}.fly.dev`
- [ ] Railway — 🔌 — `{name}.up.railway.app`
- [ ] Heroku — 🔌 — `{name}.herokuapp.com`
- [ ] Surge — 🔌 — `{name}.surge.sh`
- [ ] Deno Deploy — 🔌 — `{name}.deno.dev`
- [ ] Firebase Hosting — 🔌 — `{name}.web.app` / `.firebaseapp.com`
- [ ] Azure Static Web Apps — 🔌 — `{name}.azurestaticapps.net`
- [ ] AWS Amplify — 🔌 — `{name}.amplifyapp.com`
- [ ] Replit — 🔌 — `{name}.repl.co`
- [ ] Glitch — 🔌 — `{name}.glitch.me`
- [ ] Read the Docs — 🔌 — `{name}.readthedocs.io`
- [ ] Neocities — 🔌 — `{name}.neocities.org`

More package registries:

- [ ] Open VSX — 🌐 — `open-vsx.org/api/{namespace}/{name}` (scoped)
- [ ] CocoaPods — 🌐 — `trunk.cocoapods.org`
- [ ] Swift Package Index — 🌐 — `swiftpackageindex.com`
- [ ] Gradle Plugin Portal — 🌐 — `plugins.gradle.org`
- [ ] Ansible Galaxy — 🌐 — `galaxy.ansible.com`
- [ ] Puppet Forge — 🌐 — `forge.puppet.com`
- [ ] Chef Supermarket — 🌐 — `supermarket.chef.io`
- [ ] Elm packages — 🌐 — `package.elm-lang.org`
- [ ] Quay (containers) — 🌐 — `quay.io`
- [ ] Meteor — 🌐 — `packages.meteor.com`
- [ ] Nimble (Nim) — 🌐 — `nimble.directory`
- [ ] Crystal Shards — 🌐 — `shards.info`

More social / community:

- [ ] Stack Overflow / Stack Exchange — 🌐 — `api.stackexchange.com`
- [ ] Quora — 🕸
- [ ] Lobste.rs — 🕸
- [ ] HackerNoon — 🌐
- [ ] Beehiiv — 🕸
- [ ] Ghost — 🕸
- [ ] WordPress.com — 🌐 — `{name}.wordpress.com`
- [ ] Blogger — 🕸
- [ ] Tumblr — 🔑
- [ ] Vimeo — 🔑
- [ ] Dailymotion — 🌐
- [ ] Odysee / LBRY — 🌐
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

- [ ] Ko-fi — 🌐
- [ ] Buy Me a Coffee — 🌐
- [ ] Lemon Squeezy — 🌐
- [ ] thanks.dev — 🕸 — Cloudflare-blocked (403)
- [ ] Liberapay — 🕸 — 404 lacks CORS
- [ ] PayPal.me — 🕸 — 301 redirect / always 200, no signal
- [ ] Itch.io — 🕸
- [ ] Spotify (artist) — 🔑
- [ ] Apple Music — 🔑
- [ ] Audiomack — 🌐
- [ ] Mixcloud — 🌐
- [ ] Discogs — 🔑
- [ ] Last.fm — 🔑
- [ ] Letterboxd — 🕸
- [ ] Goodreads — 🕸
- [ ] MyAnimeList — 🌐
- [ ] AniList — 🌐 — GraphQL
- [ ] Trakt — 🔑
- [ ] Genius — 🔑
- [ ] Strava — 🔑
- [ ] Duolingo — 🕸

AI / dev tools:

- [ ] Replicate — 🔑
- [ ] Kaggle — 🔑
- [ ] CodeSandbox — 🌐
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

- [ ] eu.org — 🌐 — DoH-checkable (no wildcard); note: sends the name to the resolver
- [ ] duckdns.org — 🌐 — DoH-checkable (no wildcard)
- [ ] freedns.afraid.org — 🌐 — DoH (verify no wildcard)
- [ ] no-ip.org — 🌐 — DoH (verify no wildcard)
- [ ] dynv6.net — 🌐 — DoH (verify no wildcard)
- [ ] noted.lol — 🌐 — DoH-checkable (no wildcard)
- [ ] js.org — 🔌 — wildcard DNS; needs HTTP status via proxy
- [ ] is-a.dev — 🔌 — wildcard DNS
- [ ] is-a-good.dev — 🔌 — wildcard DNS
- [ ] ngrok-free.app — 🔌 — needs proxy
- [ ] localhost.run — 🔌 — needs proxy
- [ ] trycloudflare.com — 🔌 — needs proxy
