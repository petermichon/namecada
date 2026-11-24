# Platform backlog

Candidate sources for checking whether a name is taken. Flat list — one entry per
line.

**Rule:** nothing gets added until its signal is verified **server-side** on both
the taken and free response (see the README's "Adding a source"). CORS is no
longer a constraint — checks run in the Rust core — so the reachability notes
below, written in the browser-direct era, must be re-checked before implementing.

## Markers

- `✅` — checked: a provider is implemented and shipped.
- `🔑` — keyed: a unique claimable label, but the check needs credentials.
- `🕸` — deferred: the platform has a unique claimable label, but no keyless
  signal yet (challenge, bot wall, always-200 shell, dead third-party API).
  Re-probe before promoting.
- `🔌` — ruled out: there is no unique claimable label to probe — composite
  (`owner/name`), ephemeral or generated hosts, auto-assigned ids, display
  names, or a namespace another source already covers.
- `🧪` — verified, not ported: the probe saw a taken/free split on that route,
  so a spec can be written from it, but no provider ships it yet.

The test is the **platform's namespace, not the fetch**: when the names are
unique and claimable the entry belongs under 🔑/🕸 even if today's probe sees
nothing, and a source that answers the same for every name is out either way.

## Providers

- [x] Domains (RDAP TLDs) — ✅ — `.com .net .org .dev .io .ai .app` (`.io` via fallback) + any TLD in the IANA bootstrap
- [x] ENS (`.eth`) — ✅ — `api.ensideas.com/ens/resolve/{name}.eth` (200 always; non-null `address` = taken, `null` = free; third-party resolver — a follow-up could query a public Ethereum RPC against the ENS registry directly)
- [x] Handshake (HNS) — ✅ — `hnsscan.com/api/names/{name}/details` (200 for a known name; `info.state` `CLOSED` = taken, `AVAILABLE`/`UNKNOWN` = free — never use the historical `info.registered` flag alone; the ICANN special-use names — `example`, `test`, `invalid`, `localhost`, `local` — answer 502 and are reported unavailable without a request; any other 502 degrades to unknown; third-party explorer with an undocumented API)
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
- [x] CPAN — ✅ — `fastapi.metacpan.org/v1/module/{name}` (200 taken / 404 free; module names are case-sensitive — DBI exists, dbi does not — so this row is the only one marked `caseSensitive`, which keeps the user's raw case while every other source folds to lowercase; `metacpan.org/pod/{name}` is the page, but it Cloudflare-challenges keyless requests)
- [x] Lobsters — ✅ — `lobste.rs/~{name}` (200 taken / 404 free)
- [x] Blogger — ✅ — `{name}.blogspot.com` (wildcard DNS; HTTP 200 taken / 404 free)
- [x] Flickr — ✅ — `flickr.com/people/{name}/` (200 taken / 404 free)
- [x] StackBlitz — ✅ — `stackblitz.com/@{name}` (200 taken / 404 free; needs an HTML Accept — JSON Accept 406s every name)
- [x] Slack — ✅ — `{name}.slack.com` (an existing workspace answers 200/403, a free subdomain 404s; 403 is the login wall, not a challenge)
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
- [x] Twitch — ✅ — `twitch.tv/{name}` (keyless profile title: `{name} - Twitch` taken / plain `Twitch` free; the Helix API needs a Client-Id plus an app token, so it is not used)
- [x] Instagram — ✅ — `instagram.com/{name}/` (body signal: a claimed profile embeds `"username":"{name}"` — or a `<title>` starting with the handle — while a free name returns the generic `<title>Instagram</title>` shell; every name is a 200, so status alone has no signal)
- [x] Hacker News — ✅ — `hacker-news.firebaseio.com/v0/user/{name}.json` (keyless official API: `null` free / an object with `id` taken; the profile page answers 403 "Sorry." to non-browser clients, and `null` does cover dormant accounts — 37 accounts from 2010 comments, down to karma 1, all resolve)
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
- [x] DeviantArt — ✅ — `deviantart.com/{name}` (keyless profile: 200 taken / 404 free; a name with no profile 404s — the old "always-200 SPA" reading no longer holds)
- [x] TikTok — ✅ — `www.tiktok.com/oembed?url=https://www.tiktok.com/@{name}` (keyless oEmbed: 200 JSON with `author_url` taken / 400 free; the profile page itself is an always-200 shell)
- [x] Threads — ✅ — `threads.net/@{name}` (keyless: the server-rendered `<title>` names the profile, after a display name or alone — `Sofyan (@sofyan) • Threads` and `@acema • Threads` taken / `Threads • Log in` free)
- [x] Duolingo — ✅ — `duolingo.com/profile/{name}` (keyless `og:title`: `Google on Duolingo` taken / `Learn a language for free` free; the page `<title>` is just `Duolingo`)
- [x] Telegram — ✅ — `t.me/{name}` (keyless `<title>`: `Telegram: View @name` taken / `Telegram: Contact @name` free)
- [x] Snapchat — ✅ — `snapchat.com/add/{name}` (keyless profile: 200 taken / 404 free; the old note probed a URL shape that no longer exists)
- [x] Strava — ✅ — `strava.com/athletes/{name}` (keyless public athlete page: 200 taken / 404 free; previously listed as keyed — the profile route needs no token; no documented length rule found, so the row stays unbounded rather than guessing)
- [x] Swift Package Index — ✅ — `swiftpackageindex.com/{name}` (keyless publisher page: 200 taken / 404 free; the page lists packages, so an account that publishes none 404s and reads as free)
- [ ] Companies House UK — 🔑 — `api.company-information.service.gov.uk/search/companies?q={name}` (key; Basic auth; exact company-name hit taken / else free) — deferred, not ported: no keyless signal
- [ ] OpenCorporates — 🔑 — `api.opencorporates.com/v0.4/companies/search?q={name}` (token; `X-API-TOKEN` header; exact company-name hit taken / else free) — deferred, not ported: no keyless signal
- [x] Firefox Add-ons — ✅ — `addons.mozilla.org/api/v5`
- [x] SoundCloud — ✅ — `soundcloud.com/{name}` (keyless profile: 200 taken / 404 free)
- [x] Bandcamp — ✅ — `{name}.bandcamp.com` (a free name 303s to `bandcamp.com/signup?new_domain=`, a claimed one to its own content, so the redirect target is the verdict and must not be followed; the fan path `bandcamp.com/{name}` is not the signal — it 404s for artist-only accounts)
- [x] Gumroad — ✅ — `{name}.gumroad.com` (wildcard DNS; HTTP 404, no CORS; HTML Accept; user-chosen storefront names)
- [x] Vercel — ✅ — `{name}.vercel.app` (wildcard DNS; free names 404; a claimed deployment that sits behind Vercel SSO answers 401 for a JSON request (302 to `vercel.com/sso-api` for HTML), which is taken, so the row maps 200/401 taken / 404 free)
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
- [ ] Linktree — 🕸 — `linktr.ee/{name}`: a claimed name answers 200 with a `<title>` naming the profile (`Google | Linktree`, `github Official: … | Linktree`), a free one 404s. Re-probed 2026-11-17 on three label pairs and the split held, which contradicts the earlier 2026-11 reading (every cold key 429 with an empty body, so the check would echo the CDN cache); deferred pending a capture, but the current signal looks real
- [x] Mixcloud — ✅ — `api.mixcloud.com/{name}/`
- [ ] Discogs — 🔑 — `api.discogs.com/users/{name}` (token; the keyless route 403s; 200 taken / 404 free) — deferred, not ported: keyed only
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
- [x] 9GAG — ✅ — `9gag.com/u/{name}` (keyless profile: 200 taken / 404 free; status-only)
- [x] About.me — ✅ — `about.me/{name}` (keyless profile: 200 taken / 404 free; status-only)
- [x] Clubhouse — ✅ — `clubhouse.com/@{name}` (keyless profile: 200 taken / 404 free; `clubhouse.com` 301s to `www.clubhouse.com`, followed; status-only)
- [x] Codewars — ✅ — `codewars.com/users/{name}` (keyless profile: 200 taken / 404 free; status-only, but the page needs an HTML Accept — a JSON Accept makes a free name answer 200)
- [x] Gravatar — ✅ — `gravatar.com/{name}` (keyless profile: 200 taken / 404 free; status-only)
- [x] Launchpad — ✅ — `launchpad.net/~{name}` (keyless profile: 200 taken / 404 free; status-only)
- [x] Lichess Team — ✅ — `lichess.org/team/{name}` (keyless team page: 200 taken / 404 free; status-only)
- [x] AngelList — ✅ — `angel.co/company/{name}` (keyless company page: 200 taken / 404 free; status-only, `angel.co` 301s to `wellfound.com` and the redirect is followed)
- [x] Buttondown — ✅ — `buttondown.com/{name}` (keyless newsletter: 200 taken / 404 free; status-only)
- [x] Built In — ✅ — `builtin.com/company/{name}` (keyless company page: 200 taken / 404 free; status-only)
- [x] Disqus — ✅ — `disqus.com/by/{name}` (keyless profile: 200 taken / 404 free; status-only, `/by/{name}` 301s to the trailing-slash form, followed)
- [x] Donorbox — ✅ — `donorbox.org/{name}` (keyless campaign page: 200 taken / 404 free; status-only)
- [x] Douyu — ✅ — `douyu.com/{name}` (keyless streamer page: 200 taken / 404 free; status-only, `douyu.com/{name}` 301s to `www.douyu.com/{name}`, which 302s to the numeric room, both followed)
- [x] Drupal.org Projects — ✅ — `drupal.org/project/{name}` (keyless project page: 200 taken / 404 free; status-only; modules and themes share one namespace, so one row covers both)
- [x] Experts Exchange — ✅ — `experts-exchange.com/members/{name}` (keyless profile: 200 taken / 404 free; status-only, the route 301s to the trailing-slash form and then to `www.experts-exchange.com/members/{name}.html`, followed)
- [x] Gab — ✅ — `gab.com/{name}` (keyless profile: 200 taken / 404 free; status-only)
- [x] Gitea — ✅ — `gitea.com/{name}` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Gitee — ✅ — `gitee.com/{name}` (keyless profile: 200 taken / 404 free; status-only, HTML Accept — a JSON Accept answers 404 even for a claimed profile)
- [x] GOG — ✅ — `gog.com/u/{name}` (keyless profile: 200 taken / 404 free; status-only, HTML Accept; `gog.com/u/{name}` 301s to `www.gog.com`, followed, and a free name ends on `/404`)
- [x] Goodreads — ✅ — `goodreads.com/{name}` (keyless profile: 200 taken / 404 free; status-only, HTML Accept — a JSON Accept is answered 406 even for a claimed profile; the route 301s to `www.goodreads.com`, followed)
- [x] House-Mixes — ✅ — `house-mixes.com/profile/{name}` (keyless profile: 200 taken / 404 free; status-only, HTML Accept; `/profile/{name}` 308s to the bare `house-mixes.com/{name}` profile, followed)
- [x] Hype Machine — ✅ — `hypem.com/{name}` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Indie Hackers — ✅ — `indiehackers.com/{name}` (keyless profile: 200 taken / 404 free; status-only, HTML Accept; `indiehackers.com` 301s to `www.indiehackers.com`, followed)
- [x] Kick — ✅ — `kick.com/{name}` (keyless channel page: 200 taken / 404 free; status-only, HTML Accept; `acema` is itself a claimed channel, so the free branch was proven with `zzqxnvc`, which 404s with `Channel Not Found`)
- [x] Komoot — ✅ — `komoot.com/user/{name}` (keyless profile: 200 taken / 404 free; status-only, HTML Accept; `komoot.com` 301s to `www.komoot.com`, followed)
- [x] Kongregate — ✅ — `kongregate.com/accounts/{name}` (keyless profile: 200 taken / 404 free; status-only, HTML Accept; the route 301s twice, to `www.kongregate.com/en/accounts/{name}`, followed)
- [x] Libib — ✅ — `libib.com/u/{name}` (keyless profile: 200 taken / 404 free; status-only, HTML Accept; `libib.com` 301s to `www.libib.com`, followed)

## Candidates

### Verified signals, not ported

Triaged 2026-11-24, in place of the `similar-sites.md` scout list (1,476 rows,
never verified): 163 were keyed, 287 had no claimable route (search boxes,
opaque ids), and 1,026 had a handle or subdomain shape. 173 of those already
appear in this file or in the registry, leaving 853 to probe — 692 handle
routes over HTTP and 161 subdomains over DNS. Of the handle routes 123 split on
`google` against `zzqxnvc`; after dropping soft-404s, bot walls, redirect-only
diffs and 13 routes that are another source's namespace, 77 are below. The
remainder had no signal: 155 shells, 102 challenges, 22 blocked, 34 dead and
243 that answer 404 for every label.

The split was measured once, so re-probe before writing a spec.

- [ ] Canva — 🧪 — `canva.com/%s` (probe 2026-11-24: 403 taken / 404 free)
- [ ] Home Assistant Integrations — 🧪 — `home-assistant.io/integrations/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] LinkedIn Company — 🧪 — `linkedin.com/company/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] Loops — 🧪 — `loops.video/@%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] LyricsTranslate — 🧪 — `lyricstranslate.com/en/translator/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] Metacritic — 🧪 — `metacritic.com/user/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] Mixlr — 🧪 — `mixlr.com/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] Moj — 🧪 — `mojapp.in/@%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] Mojang — 🧪 — `api.mojang.com/users/profiles/minecraft/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] MusicBrainz — 🧪 — `musicbrainz.org/user/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] Musixmatch — 🧪 — `musixmatch.com/artist/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] MyDramaList — 🧪 — `mydramalist.com/profile/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] Naver Blog — 🧪 — `blog.naver.com/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] NuGet profiles — 🧪 — `nuget.org/profiles/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] Open Library — 🧪 — `openlibrary.org/people/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] Open Source Collective — 🧪 — `opencollective.com/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] Owler — 🧪 — `owler.com/company/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] Pillowfort — 🧪 — `pillowfort.social/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] Pixelfed — 🧪 — `pixelfed.social/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] PlayCode — 🧪 — `playcode.io/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] Poshmark — 🧪 — `poshmark.com/closet/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] Racket Packages — 🧪 — `pkgs.racket-lang.org/package/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] ReverbNation — 🧪 — `reverbnation.com/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] RubyGems profiles — 🧪 — `rubygems.org/profiles/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] Skool — 🧪 — `skool.com/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] Slashdot — 🧪 — `slashdot.org/~%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] Spotify — 🧪 — `open.spotify.com/user/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] stats.fm — 🧪 — `stats.fm/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] Steemit — 🧪 — `steemit.com/@%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] Streamlabs — 🧪 — `streamlabs.com/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] SuperRare — 🧪 — `superrare.com/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] Tildes — 🧪 — `tildes.net/user/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] Tipeee — 🧪 — `tipeee.com/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] Uneed — 🧪 — `uneed.best/tool/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] V2EX — 🧪 — `v2ex.com/member/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] Vidyard — 🧪 — `vidyard.com/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] Visual Studio Marketplace — 🧪 — `marketplace.visualstudio.com/publishers/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] Wattpad — 🧪 — `wattpad.com/user/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] webhook.site — 🧪 — `webhook.site/%s` (probe 2026-11-24: 200 taken / 404 free; confirm the segment is a vanity label, not a request token)
- [ ] Wellfound — 🧪 — `wellfound.com/company/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] Whatnot — 🧪 — `whatnot.com/user/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] Whop — 🧪 — `whop.com/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] Workana — 🧪 — `workana.com/freelancer/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] Write.as — 🧪 — `write.as/%s` (probe 2026-11-24: 200 taken / 404 free)
- [ ] Zeerk — 🧪 — `zeerk.com/%s` (probe 2026-11-24: 200 taken / 404 free)

- [ ] ddns.net — 🧪 — `%s.ddns.net` (DNS probe 2026-11-24: `google.ddns.net` resolves, `zzqxnvc.ddns.net` is NXDOMAIN; registration needs an account, the check does not)
- [ ] ddnsgeek.com — 🧪 — `%s.ddnsgeek.com` (DNS probe 2026-11-24: `google.ddnsgeek.com` resolves, `zzqxnvc.ddnsgeek.com` is NXDOMAIN; registration needs an account, the check does not)
- [ ] hopto.org — 🧪 — `%s.hopto.org` (DNS probe 2026-11-24: `google.hopto.org` resolves, `zzqxnvc.hopto.org` is NXDOMAIN; registration needs an account, the check does not)
- [ ] myftp.org — 🧪 — `%s.myftp.org` (DNS probe 2026-11-24: `google.myftp.org` resolves, `zzqxnvc.myftp.org` is NXDOMAIN; registration needs an account, the check does not)
- [ ] zapto.org — 🧪 — `%s.zapto.org` (DNS probe 2026-11-24: `google.zapto.org` resolves, `zzqxnvc.zapto.org` is NXDOMAIN; registration needs an account, the check does not)

Status alone was a false signal on these four, so they are deferred rather than
portable:

- [ ] AppSumo — 🕸 — `appsumo.com/products/{name}`: a missing product 301s to the
  homepage and answers 200 with a generic title, while a real product stays on
  its product page, so the redirect target is the signal and the status is not
- [ ] Chess.com Club — 🕸 — `chess.com/club/{name}`: a missing club 301s to
  `/clubs/home` (200, generic landing) while a real club stays on
  `/club/{slug}`; the redirect target is the signal, not the status
- [ ] CSS Deck — 🕸 — `cssdeck.com/user/{name}`: WordPress fuzzy redirection —
  unknown labels 301 to unrelated blog posts (`google` lands on a Google
  Analytics post, `tim` on a CSS timetable), so this is not a profile namespace
- [ ] Android Stack Exchange — 🕸 — `android.stackexchange.com/users/{name}`: the
  slug is user-settable (`/users/username-settings/`) and a real vanity profile
  did split 200/404 once, but every later request, probe and curl alike, is 429
  behind a Cloudflare challenge, so no fixture can be captured

- [ ] ccTLDs / new gTLDs — 🔌 — ruled out: not a distinct source — the IANA RDAP bootstrap only covers some TLDs, and any TLD it lists is already checked by the Domains provider
- [ ] Unstoppable Domains — 🔑 — deferred: the resolver API is keyed (401 unauthenticated)
- [ ] Packagist — 🔌 — composite `vendor/name`, no single-label namespace
- [ ] Maven Central — 🔌 — composite `group:artifact`, no single-label namespace
- [ ] Deno (`deno.land/x`) — 🕸 — `/x/{name}` 302-redirects and even taken names resolve 404, no signal
- [ ] Docker Hub — 🔌 — composite `user/repo`, no single-label namespace
- [ ] Flathub — 🔌 — ruled out: the app-id is reverse-DNS, not a single label, so a `{name}` probe never names a claimable app
- [ ] LuaRocks — 🔌 — composite `owner/name`, no single-label namespace
- [ ] Terraform Registry — 🔌 — scoped `namespace/type` path, not a single label
- [ ] VS Code Marketplace — 🕸 — no simple API (use Open VSX)
- [ ] JetBrains Marketplace — 🔌 — ruled out: `/api/searchPlugins?search=` is a fuzzy substring search, not an exact-name signal, and plugin ids are numeric
- [ ] Figma Community — 🕸 — no public API
- [ ] 99designs — 🕸 — re-checked server-side (2026-11): from a non-US egress `99designs.com/profiles/{name}` GeoIP-redirects to `en.99designs.fr`, which 404s every name including `99designs` itself, so status alone cannot separate taken from free; needs a locale-pinned host, deferred, not ported
- [ ] Shopify App Store — 🕸 — re-checked server-side (2026-11): `apps.shopify.com/{slug}` 404s for both a known app and a fake slug on the probed shapes, so there is no confirmed signal
- [ ] Nostr (NIP-05) — 🔌 — ruled out: names are per-domain (`https://{domain}/.well-known/nostr.json?name={name}`), not a global namespace, so no single-label source exists to probe
- [ ] Fandom — 🕸 — re-checked server-side (2026-11): `{name}.fandom.com` is wildcard DNS and every host answers 403 (Cloudflare) for real and fake names alike, so there is no signal
- [ ] Bilibili — 🕸 — API returns 412 (banned)
- [ ] Discord — 🕸 — usernames have no public API

- [ ] Hashnode — 🕸 — the GraphQL host 301s (reads now paid) and `hashnode.com/{name}` is an always-200 SPA, so no signal
- [ ] Product Hunt — 🔑 — deferred: the GraphQL API is keyed
- [ ] LinkedIn — 🕸 — no public vanity-availability API; profile pages are auth-walled (999/redirect for both real and fake), so it cannot answer "is this name free"


- [ ] Facebook — 🕸 — keyless pass: 200 (present) / 302 (missing) is a login redirect, not a name signal; needs a key or a scrape — deferred, not ported
- [ ] Reddit — 🕸 — keyless pass: 301 → (followed) 200 for both real and fake names, so no status signal (supersedes the old "403 from server IPs"); needs a key or a scrape — deferred, not ported
- [ ] Pinterest — 🕸 — keyless pass: every name returns 200 (JS/SPA shell), so no status signal; needs a key or a scrape — deferred, not ported
- [ ] Medium — 🕸 — keyless pass: 403 (bot-block) for both real and fake names, so no status signal; needs a key or a scrape — deferred, not ported
- [ ] OpenSea — 🕸 — re-checked server-side (2026-11): the whole site now sits behind a Cloudflare interstitial, so `opensea.io/{name}` returns 403 "Just a moment…" for both real and fake names (curl with browser headers, a real browser, and an unrelated third-party proxy were all challenged), so there is no status signal; deferred, not ported
- [ ] XDA Forums — 🕸 — Cloudflare-blocked (403)
- [ ] USPTO (trademarks) — 🔌 — class-based, fuzzy-similarity trademark search; not a single claimable label, so it cannot answer name availability
- [ ] EUIPO (trademarks) — 🔌 — class-based trademark search (eSearch plus, OAuth); not a single claimable label, so it cannot answer name availability
- [ ] WIPO Global Brand Database — 🕸
- [ ] SEC EDGAR — 🔌 — keyless (descriptive User-Agent) but not a claimable namespace: it indexes issuer/filing names, which you cannot register
- [ ] Crunchbase — 🕸
- [ ] US state business registries — 🕸
- [ ] Chrome Web Store — 🔌 — ruled out: extension names are display names and items are keyed by a 32-character id, so there is no claimable label; every name also answers 302
- [ ] Edge Add-ons — 🔌 — ruled out: same shape as the Chrome Web Store — display names plus product ids, so there is no claimable label; every name also answers 200
- [ ] Apple App Store — 🕸
- [ ] Google Play — 🕸
- [x] Steam — ✅ — `steamcommunity.com/id/{name}?xml=1` (body signal: a taken custom id reports a `<steamID64>` element, a free one an `<error>` element; both are 200, so status alone has no signal)
- [ ] Epic Games — 🕸
- [ ] Patreon — 🕸 — bot-blocked: 403 for both real and fake names, so no signal
- [ ] Kickstarter — 🕸 — bot-blocked: 403 for both real and fake names, so no signal
- [ ] Indiegogo — 🕸
- [ ] Etsy — 🕸 — re-checked server-side (2026-11): `etsy.com/shop/{name}` 403s (bot wall) for real and fake shops, so there is no signal
- [ ] Amazon — 🔌 — ruled out: not a name namespace — `amazon.com/{name}` is a brand/search route (200 for `google`, 404 for `amazon`, `github` and fake names), and public profiles are addressed by an internal id, not a claimable label
- [ ] eBay — 🕸 — re-checked server-side (2026-11): `ebay.com/usr/{name}` 403s ("Error Page") for real and fake users, so there is no signal
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
- [ ] Gradle Plugin Portal — 🕸 — no usable API
- [ ] Ansible Galaxy — 🔌 — scoped `namespace/name`, not a single label
- [ ] Puppet Forge — 🕸 — requires owner-name slug
- [ ] Elm packages — 🔌 — `author/project`, not a single label
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
- [ ] Stack Exchange — 🔌 — display names aren't unique (no availability)
- [ ] Go modules — 🔌 — modules are paths, not single labels
- [ ] WhatsApp — 🔌 — phone-based, no username
- [ ] Gmail — 🕸 — the local part of `{name}@gmail.com` is unique and claimable, but availability exists only inside the bot-protected signup flow, which this repo won't scrape; the old keyless probe `mail.google.com/mail/gxlu?email={name}@gmail.com` now returns 204 with no `COMPASS` cookie for real and fake addresses alike (2026-11); dots and `+suffix` are normalized, so a literal local-part check would be wrong even if it worked
- [ ] Proton Mail — 🕸 — deferred: no public availability API. `account.proton.me/api/core/v4/users/available?Name={name}` needs an `x-pm-appversion` header plus an anonymous session (`POST /api/auth/v4/sessions`); with one, every name — including long random ones — returns 409 "Username already used", so taken/free is unverified. The endpoint is undocumented and version-gated, and the namespace is `local@domain` (proton.me/protonmail.com/pm.me/protonmail.ch), not a bare label
- [ ] Quora — 🕸
- [ ] Beehiiv — 🕸
- [ ] Ghost — 🕸
- [ ] WordPress.com — 🕸 — wildcard DNS
- [ ] Tumblr — 🔑
- [ ] Vimeo — 🕸 — bot-blocked: 403 for both real and fake names, so no signal
- [ ] Odysee / LBRY — 🕸 — no usable API
- [ ] VK — 🔑 — keyless profile pages answer 200 for real and fake names alike (the same generic shell) and 404 only for some names, so status has no signal; the availability API needs a token — deferred, not ported
- [ ] Weibo — 🕸
- [ ] Zhihu — 🕸
- [ ] Douyin — 🕸
- [ ] Line — 🔑
- [ ] Kakao — 🔑
- [ ] Matrix — 🕸 — inconclusive: availability is a per-homeserver localpart, and the register/available endpoint only answers on open-registration servers (see plan/results/P2.md)
- [ ] Revolt — 🕸 — the profile route 301s, so no status signal
- [ ] Guilded — 🕸 — re-checked server-side (2026-11): `guilded.gg/{name}` 200s for real and fake names, so there is no signal
- [ ] OpenNIC — 🕸 — the peerage TLDs are a real namespace, but the public DoH endpoints are dead (2026-11), so there is no keyless resolver; re-check before promoting
- [ ] Ko-fi — 🕸 — redirects, no signal
- [ ] Buy Me a Coffee — 🕸 — re-checked server-side (2026-11): `buymeacoffee.com/{name}` 404s for real and fake names alike, so there is no signal
- [ ] Lemon Squeezy — 🕸 — inconsistent: 200 for some stores, 403 (bot wall) for others, 404 for the rest, so there is no reliable status signal
- [ ] thanks.dev — 🕸 — Cloudflare-blocked (403)
- [ ] PayPal.me — 🕸 — 301 redirect / always 200, no signal
- [ ] Itch.io — 🕸 — re-checked server-side (2026-11): Cloudflare challenge (403) for real and fake `{name}.itch.io` hosts
- [ ] Audiomack — 🕸 — redirects
- [ ] Last.fm — 🔑 — the keyless profile route now bot-challenges (`last.fm/user/{name}` returns Cloudflare "Client Challenge" with 200 for every name), so it would need the keyed API (`ws.audioscrobbler.com/2.0/?method=user.getinfo&user={name}&api_key={key}`); deferred
- [ ] Letterboxd — 🕸 — re-checked server-side (2026-11): Cloudflare challenge (403) for real and fake profiles
- [ ] MyAnimeList — 🕸 — Jikan API unavailable
- [ ] Trakt — 🔑
- [ ] Genius — 🔑

- [ ] CodeSandbox — 🕸 — Cloudflare-blocked (403)
- [ ] CodePen — 🕸 — re-checked server-side (2026-11): Cloudflare challenge (403) for real and fake profiles
- [ ] Observable — 🕸 — re-checked server-side (2026-11): 429 Vercel Security Checkpoint for real and fake profiles, so there is no signal
- [ ] Wellfound (AngelList) — 🕸 — re-checked server-side (2026-11): `wellfound.com/u/{name}` 404s for a real and a fake profile alike, so there is no usable shape
- [ ] HelloWork — 🔌 — job board; no public username namespace/API
- [ ] Fiverr — 🕸 — re-checked server-side (2026-11): 403 ("It needs a human touch") for fake sellers, so there is no signal
- [ ] Upwork — 🕸
- [ ] Polywork — 🕸 — re-checked server-side (2026-11): the host no longer answers (connection failure), so it is not viable
- [ ] freedns.afraid.org — 🔌 — ruled out: FreeDNS hostnames live under shared
  domains (`{name}.{shared-domain}`), not `{name}.freedns.afraid.org`, so a
  single-label probe never names a claimable host.
- [ ] js.org — 🕸 — `{name}.js.org` subdomains are unique and claimed (a pull
  request to the site repo), but the zone is wildcard DNS and every host answers
  200/302 for both real and fake names, so neither DNS nor HTTP status isolates a
  claim; deferred, not ported
- [ ] is-a.dev — 🕸 — `{name}.is-a.dev` subdomains are unique and claimed (a pull
  request to the site repo), but the zone is wildcard DNS and every host answers
  200/302 for both real and fake names, so neither DNS nor HTTP status isolates a
  claim; deferred, not ported
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
- [ ] Trustpilot — 🔌 — profiles are business units keyed by domain
  (`trustpilot.com/review/{domain}`), not a claimable name — a 200 means "a
  profile exists for that domain" (including units with no reviews) and a 404
  says nothing about availability, so neither maps to taken/free. The Business
  Units API is keyed (401 without `apikey`), `robots.txt` is `User-agent: *` /
  `Disallow: /`, and keyless requests get a Cloudflare challenge (403), so there
  is no verifiable signal
