# Platform backlog

Candidate sources for checking whether a name is taken. Flat list — one entry per
line.

**Rule:** nothing gets added to the app until its CORS is verified on **both** the
taken and free response (see the README's "Adding a source"). Reachability below
is best-effort and must be re-checked before implementing.

Legend:

- ✅ verified browser-direct (CORS tested)
- 🌐 likely browser-direct (public CORS API — verify)
- 🔑 needs an API key and a proxy
- 🕸 no public API / scrape / TOS-blocked

Status: `[x]` = already a provider, `[ ]` = candidate.

- [x] Domains (RDAP TLDs) — ✅ — `.com .org .net .dev` + any TLD in the IANA bootstrap
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
- [ ] Maven Central — 🌐 — `search.maven.org` / `central.sonatype.com`
- [ ] Hex (Elixir) — 🌐 — `hex.pm/api/packages/{name}`
- [ ] pub.dev (Dart) — 🌐 — `pub.dev/api/packages/{name}`
- [ ] JSR (JavaScript) — 🌐 — `jsr.io`
- [ ] Deno (`deno.land/x`) — 🌐 — `deno.land/x/{name}`
- [ ] Docker Hub — 🕸 — no CORS
- [ ] Homebrew — 🌐 — `formulae.brew.sh/api/formula/{name}.json`
- [ ] Chocolatey — 🌐 — `community.chocolatey.org` API
- [ ] Snap Store — 🌐 — `api.snapcraft.io`
- [ ] Flathub — 🌐 — `flathub.org` API
- [ ] AUR (Arch) — 🌐 — `aur.archlinux.org/rpc`
- [ ] CPAN — 🌐 — `metacpan.org` API
- [ ] Hackage — 🌐 — `hackage.haskell.org`
- [ ] Clojars — 🌐 — `clojars.org` API
- [ ] LuaRocks — 🌐 — `luarocks.org` API
- [ ] Terraform Registry — 🌐 — `registry.terraform.io`
- [ ] VS Code Marketplace — 🌐 — marketplace API
- [ ] JetBrains Marketplace — 🌐 — `plugins.jetbrains.com` API
- [ ] WordPress plugins/themes — 🌐 — `api.wordpress.org`
- [ ] Figma Community — 🕸 — no public API
- [ ] Shopify App Store — 🕸
- [ ] Go modules — 🕸 — modules are paths, not single labels
- [x] GitHub — ✅ — `api.github.com/users/{name}`
- [x] Bluesky — ✅ — `public.api.bsky.app` resolveHandle (`{name}.bsky.social`)
- [x] Mastodon (`mastodon.social`) — ✅ — `/api/v1/accounts/lookup?acct={name}`
- [ ] GitLab — 🌐 — `gitlab.com/api/v4/users?username={name}`
- [ ] Bitbucket — 🌐 — `api.bitbucket.org/2.0/users/{name}`
- [ ] Nostr (NIP-05) — 🌐 — `/.well-known/nostr.json`
- [ ] Farcaster — 🌐 — fnames registry API
- [ ] Lemmy — 🌐 — `/api/v3/user?username={name}`
- [ ] Dev.to — 🌐 — `dev.to/api/users/by_username`
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
- [ ] Telegram — 🕸 — `t.me/{name}` scrape
- [ ] Medium — 🕸
- [ ] Hacker News — 🕸 — Algolia API has no availability signal
- [ ] USPTO (trademarks) — 🔑
- [ ] EUIPO (trademarks) — 🔑
- [ ] WIPO Global Brand Database — 🕸
- [ ] Companies House UK — 🔑
- [ ] OpenCorporates — 🔑
- [ ] SEC EDGAR — 🔑
- [ ] Crunchbase — 🕸
- [ ] US state business registries — 🕸
- [ ] Chrome Web Store — 🌐
- [ ] Firefox Add-ons — 🌐 — `addons.mozilla.org/api/v5`
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
