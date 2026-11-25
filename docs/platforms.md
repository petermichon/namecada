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
- [x] Linktree — ✅ — `linktr.ee/{name}` (200 taken / 404 free; needs an HTML Accept, a JSON Accept 406s; the earlier cold-key 429s were the CDN cache, a capture now holds)
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
- [x] LinkedIn Company — ✅ — `linkedin.com/company/{name}` (keyless company page: 200 taken / 404 free; status-only, HTML Accept; the route 301s to `www.linkedin.com`, followed; `acema` is itself a claimed company, so the free branch was proven with `zzqxnvc`; a bot wall answers 999, which reads as unknown)
- [x] Loops — ✅ — `loops.video/@{name}` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] LyricsTranslate — ✅ — `lyricstranslate.com/en/translator/{name}` (keyless translator profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Mixlr — ✅ — `mixlr.com/{name}` (keyless broadcaster profile: 200 taken / 404 free; status-only, HTML Accept — a JSON Accept is answered 406 even for a claimed profile)
- [x] Moj — ✅ — `mojapp.in/@{name}` (keyless profile: 200 taken / 404 free; status-only, HTML Accept; a free name redirects to `/404`, followed)
- [x] MusicBrainz — ✅ — `musicbrainz.org/user/{name}` (keyless editor profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Musixmatch — ✅ — `musixmatch.com/artist/{name}` (keyless artist page: 200 taken / 404 free; status-only, HTML Accept; `musixmatch.com` 301s to `www.musixmatch.com`, followed)
- [x] MyDramaList — ✅ — `mydramalist.com/profile/{name}` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Naver Blog — ✅ — `blog.naver.com/{name}` (keyless blog: 200 taken / 404 free; status-only, HTML Accept; `acema` is itself a claimed blog, so the free branch was proven with `zzqxnvc`)
- [x] NuGet profiles — ✅ — `nuget.org/profiles/{name}` (keyless profile: 200 taken / 404 free; status-only, HTML Accept; `nuget.org` 301s to `www.nuget.org`, followed; separate namespace from the package id check at `nuget.org/packages/{name}`)
- [x] Open Library — ✅ — `openlibrary.org/people/{name}` (keyless reader profile: 200 taken / 404 free; status-only, HTML Accept; a profile's display name can differ from its username, so the page title is not the label)
- [x] Owler — ✅ — `owler.com/company/{name}` (keyless company profile: 200 taken / 404 free; status-only, HTML Accept; `owler.com` 301s to `www.owler.com`, followed; `acema` is itself a claimed company, so the free branch was proven with `zzqxnvc`; a bot wall answers 403, which reads as unknown)
- [x] Pillowfort — ✅ — `pillowfort.social/{name}` (keyless profile: 200 taken / 404 free; status-only, HTML Accept; `pillowfort.social` 301s to `www.pillowfort.social`, followed)
- [x] Pixelfed — ✅ — `pixelfed.social/{name}` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] PlayCode — ✅ — `playcode.io/{name}` (keyless playground project: 200 taken / 404 free; status-only, HTML Accept; a claimed slug returns the project app shell, whose title is the project's own, while a missing slug falls back to the marketing page with 404)
- [x] Poshmark — ✅ — `poshmark.com/closet/{name}` (keyless closet: 200 taken / 404 free; status-only, HTML Accept; `acema` is itself a claimed closet, so the free branch was proven with `zzqxnvc`)
- [x] Racket Packages — ✅ — `pkgs.racket-lang.org/package/{name}` (keyless package page: 200 taken / 404 free; status-only, HTML Accept)
- [x] ReverbNation — ✅ — `reverbnation.com/{name}` (keyless artist profile: 200 taken / 404 free; status-only, HTML Accept; a claimed profile 301s to `legacy.reverbnation.com` and a free name to `www.reverbnation.com`, both followed)
- [x] RubyGems profiles — ✅ — `rubygems.org/profiles/{name}` (keyless account profile: 200 taken / 404 free; status-only, HTML Accept; separate namespace from the gem check at `rubygems.org/gems/{name}`)
- [x] Skool — ✅ — `skool.com/{name}` (keyless community: 200 taken / 404 free; status-only, HTML Accept; `skool.com` 301s to `www.skool.com/{name}/about`, followed)
- [x] Slashdot — ✅ — `slashdot.org/~{name}` (keyless user journal: 200 taken / 404 free; status-only, HTML Accept; a missing user's 404 page carries the same shell as a real one, only with "user you requested does not exist")
- [x] Spotify — ✅ — `open.spotify.com/user/{name}` (keyless listener profile: 200 taken / 404 free; status-only, HTML Accept; the web player serves a login shell for a claimed profile, but a missing user is a genuine 404; `acema` is itself a claimed profile, so the free branch was proven with `zzqxnvc`)
- [x] stats.fm — ✅ — `stats.fm/{name}` (keyless listening profile: 200 taken / 404 free; status-only, HTML Accept; a profile's display name can differ from its username, so the page title is not the label; `acema` is itself a claimed profile, so the free branch was proven with `zzqxnvc`)
- [x] Steemit — ✅ — `steemit.com/@{name}` (keyless blog: 200 taken / 404 free; status-only, HTML Accept)
- [x] Streamlabs — ✅ — `streamlabs.com/{name}` (keyless tip page: 200 taken / 404 free; status-only, HTML Accept; a claimed name 302s to `streamlabs.com/{name}/tip`, followed; `acema` is itself a claimed page, so the free branch was proven with `zzqxnvc`)
- [x] SuperRare — ✅ — `superrare.com/{name}` (keyless artist profile: 200 taken / 404 free; status-only, HTML Accept; a missing name still renders the app shell, but answers 404)
- [x] Tildes — ✅ — `tildes.net/user/{name}` (keyless user profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Tipeee — ✅ — `tipeee.com/{name}` (keyless creator page: 200 taken / 404 free; status-only, HTML Accept; `tipeee.com` 301s to `www.tipeee.com`, which 302s to the locale host `en.tipeee.com`, both followed)
- [x] V2EX — ✅ — `v2ex.com/member/{name}` (keyless member profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Visual Studio Marketplace — ✅ — `marketplace.visualstudio.com/publishers/{name}` (keyless publisher page: 200 taken / 404 free; status-only, HTML Accept; separate namespace from the extension ids, which stay out of scope)
- [x] Wattpad — ✅ — `wattpad.com/user/{name}` (keyless profile: 200 taken / 404 free; status-only, HTML Accept; `wattpad.com` 301s to `www.wattpad.com`, followed)
- [x] Wellfound — ✅ — `wellfound.com/company/{name}` (keyless company profile: 200 taken / 404 free; status-only, HTML Accept; separate namespace from the `/u/{name}` people route)
- [x] Whatnot — ✅ — `whatnot.com/user/{name}` (keyless profile: 200 taken / 404 free; status-only, HTML Accept; `whatnot.com` 301s to `www.whatnot.com`, followed; `acema` is itself a claimed profile, so the free branch was proven with `zzqxnvc`)
- [x] Whop — ✅ — `whop.com/{name}` (keyless community page: 200 taken / 404 free; status-only, HTML Accept; `whop.com/{name}` 308s to the trailing-slash form, followed)
- [x] Write.as — ✅ — `write.as/{name}` (keyless blog: 200 taken / 404 free; status-only, HTML Accept; a claimed blog 301s to the trailing-slash form, followed)
- [x] ddns.net — ✅ — `{name}.ddns.net` (wildcard-free DNS: resolves taken / NXDOMAIN free; registering one needs a provider account, the check is keyless)
- [x] ddnsgeek.com — ✅ — `{name}.ddnsgeek.com` (wildcard-free DNS: resolves taken / NXDOMAIN free; registering one needs a provider account, the check is keyless)
- [x] hopto.org — ✅ — `{name}.hopto.org` (wildcard-free DNS: resolves taken / NXDOMAIN free; registering one needs a provider account, the check is keyless)
- [x] myftp.org — ✅ — `{name}.myftp.org` (wildcard-free DNS: resolves taken / NXDOMAIN free; registering one needs a provider account, the check is keyless)
- [x] zapto.org — ✅ — `{name}.zapto.org` (wildcard-free DNS: resolves taken / NXDOMAIN free; registering one needs a provider account, the check is keyless)

### Namespaces and profiles added 2026-11-25

Each was measured server-side on a taken and a free label with the shipped
`namecada/1.0` User-Agent and no redirect following, then captured for both
fixture labels.

- [x] Docker Hub — ✅ — `hub.docker.com/u/{name}` (keyless user namespace: 200 taken / 404 free; status-only, JSON Accept; user names are 4–30 lowercase characters. Separate from the `user/repo` repository path, which is composite and stays ruled out)
- [x] npm orgs — ✅ — `registry.npmjs.org/-/org/{name}/user` (200 taken / 404 free; org names follow the package grammar, 214 characters max)
- [x] PyPI orgs — ✅ — `pypi.org/org/{name}/` (200 taken / 404 free; separate from the project namespace)
- [x] ArtStation — ✅ — `artstation.com/api/v2/user_profiles/{name}.json` (200 taken / 404 free; profile usernames allow `_`)
- [x] AtCoder — ✅ — `atcoder.jp/users/{name}` (200 taken / 404 free)
- [x] Bugcrowd — ✅ — `bugcrowd.com/profile-service/v1/profiles/{name}` (200 taken / 404 free; `google` is unclaimed there, so the taken branch was proven with `BlackMist`)
- [x] Carrd — ✅ — `{name}.carrd.co` (200 taken / 404 free; `google` is unclaimed there, the taken branch was proven with `liam`)
- [x] Crowdin — ✅ — `crowdin.com/profile/{name}` (200 taken / 404 free)
- [x] Freesound — ✅ — `freesound.org/people/{name}/` (200 taken / 404 free)
- [x] HackerRank — ✅ — `hackerrank.com/rest/contests/master/hackers/{name}/profile` (200 taken / 404 free; the HTML profile route 302s for both names)
- [x] Monkeytype — ✅ — `api.monkeytype.com/users/{name}/profile` (200 taken / 404 free)
- [x] MyAnimeList — ✅ — `myanimelist.net/profile/{name}` (200 taken / 404 free)
- [x] osu! — ✅ — `osu.ppy.sh/users/{name}` (200 taken / 404 free)
- [x] Pastebin — ✅ — `pastebin.com/u/{name}` (200 taken / 404 free)
- [x] TradingView — ✅ — `tradingview.com/u/{name}/` (200 taken / 404 free)
- [x] Trello — ✅ — `trello.com/1/Members/{name}` (200 taken / 404 free; the `trello.com/{name}` page is an always-200 shell)
- [x] Venmo — ✅ — `venmo.com/u/{name}` (200 taken / 404 free; `google` is unclaimed there, the taken branch was proven with `venmo`)
- [x] WakaTime — ✅ — `wakatime.com/api/v1/users/{name}` (200 taken / 404 free; `google` is unclaimed there, the taken branch was proven with `jake`)


### Corpus pass 2026-11-25

WhatsMyName, Sherlock and Maigret entries that were not in the catalogue,
each probed on a known-taken handle and two random free labels with the
shipped `namecada/1.0` client and no redirect following.

- [x] 35photo — ✅ — `35photo.pro/@%s/` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] 7dach — ✅ — `7dach.ru/profile/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Airbit — ✅ — `airbit.com/%s` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] Alik — ✅ — `www.alik.cz/u/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] AmericanThinker — ✅ — `www.americanthinker.com/author/%s/` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Aparat — ✅ — `www.aparat.com/api/fa/v1/user/user/information/username/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Arduino (Forum) — ✅ — `forum.arduino.cc/u/%s.json` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Artbreeder — ✅ — `www.artbreeder.com/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Asciinema — ✅ — `asciinema.org/~%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] AudioJungle — ✅ — `audiojungle.net/user/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] BandLab — ✅ — `www.bandlab.com/api/v1.3/users/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Bezuzyteczna — ✅ — `bezuzyteczna.pl/uzytkownicy/%s` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] Bimpos — ✅ — `ask.bimpos.com/user/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Bio Site — ✅ — `bio.site/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] BioHacking — ✅ — `forum.dangerousthings.com/u/%s` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] Blipfoto — ✅ — `www.blipfoto.com/%s` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] Blitz Tactics — ✅ — `blitztactics.com/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Boosty — ✅ — `api.boosty.to/v1/blog/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Bunpro — ✅ — `community.bunpro.jp/u/%s.json` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] BuzzFeed — ✅ — `www.buzzfeed.com/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Caddy Community — ✅ — `caddy.community/u/%s/summary` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] Calendy — ✅ — `calendly.com/api/booking/profiles/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Car Talk Community — ✅ — `community.cartalk.com/u/%s/summary` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] Carbonmade — ✅ — `%s.carbonmade.com/` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Career.Habr — ✅ — `career.habr.com/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Cfx.re — ✅ — `forum.cfx.re/u/%s.json` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Championat — ✅ — `www.championat.com/user/%s/` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Chatujme.Cz — ✅ — `profil.chatujme.cz/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Choice Community — ✅ — `choice.community/u/%s/summary` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] Choko.Link — ✅ — `choko.link/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Cnet — ✅ — `www.cnet.com/profiles/%s/` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Coda — ✅ — `coda.io/@%s/` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Codeforces — ✅ — `codeforces.com/api/user.info?handles=%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Codementor — ✅ — `www.codementor.io/@%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Commudle — ✅ — `json.commudle.com/api/v2/users?username=%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Coub — ✅ — `coub.com/api/v2/channels/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Credly — ✅ — `www.credly.com/users/%s` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] Cropty — ✅ — `api.cropty.io/v1/auth/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Cryptomator Forum — ✅ — `community.cryptomator.org/u/%s` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] Cytoid — ✅ — `cytoid.io/profile/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Daily Kos — ✅ — `www.dailykos.com/user/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] DC Inside — ✅ — `gallog.dcinside.com/%s` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] Destream — ✅ — `api.destream.net/siteapi/v2/live/details/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] devRant — ✅ — `devrant.com/api/get-user-id?app=3&username=%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Diablo — ✅ — `diablo2.io/member/%s/` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] DIBIZ — ✅ — `www.dibiz.com/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Discourse — ✅ — `meta.discourse.org/u/%s/summary.json` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Discuss.Elastic.Co — ✅ — `discuss.elastic.co/u/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Dojoverse — ✅ — `dojoverse.com/members/%s/` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Donatello — ✅ — `donatello.to/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] eToro — ✅ — `www.etoro.com/api/logininfo/v1.1/users/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Fixya — ✅ — `www.fixya.com/users/%s` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] Flipboard — ✅ — `flipboard.com/@%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Flowcode — ✅ — `www.flow.page/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Football — ✅ — `www.rusfootball.info/user/%s/` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] FreeCodeCamp — ✅ — `api.freecodecamp.org/users/get-public-profile?username=%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Freelance.RU — ✅ — `freelance.ru/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Freelance.Ua — ✅ — `freelance.ua/user/%s/` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Game Jolt — ✅ — `gamejolt.com/site-api/web/profile/@%s/` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Geocaching — ✅ — `www.geocaching.com/p/?u=%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Gradle — ✅ — `plugins.gradle.org/u/%s` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] Habtium — ✅ — `habtium.es/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Hackaday.Io — ✅ — `hackaday.io/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] HackMD — ✅ — `hackmd.io/@%s` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] Hackster — ✅ — `www.hackster.io/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] HiHello — ✅ — `www.hihello.com/author/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Hoo.Be — ✅ — `hoo.be/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Houzz — ✅ — `www.houzz.com/user/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Icons8 Community — ✅ — `community.icons8.com/u/%s/summary` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] Icq-Chat — ✅ — `icq.icqchat.co/members/%s/` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] ILGM Growers Forum — ✅ — `ilgmforum.com/u/%s/summary.json` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] ImageShack — ✅ — `imageshack.com/user/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Inaturalist — ✅ — `api.inaturalist.org/v1/users/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Instructables — ✅ — `www.instructables.com/json-api/showAuthorExists?screenName=%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Intigriti — ✅ — `app.intigriti.com/api/user/public/profile/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Ionic Forum — ✅ — `forum.ionicframework.com/u/%s` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] Issuu — ✅ — `issuu.com/%s` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] Joplin Forum — ✅ — `discourse.joplinapp.org/u/%s` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] Jupyter Community Forum — ✅ — `discourse.jupyter.org/u/%s/summary` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Komi — ✅ — `api.komi.io/api/talent/usernames/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Laracast — ✅ — `laracasts.com/@%s` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] Leasehackr — ✅ — `forum.leasehackr.com/u/%s/summary/` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] LessWrong — ✅ — `www.lesswrong.com/users/%s` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] Linux.Org.Ru — ✅ — `www.linux.org.ru/people/%s/profile` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Livejournal — ✅ — `%s.livejournal.com` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] LottieFiles — ✅ — `lottiefiles.com/%s` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] Magnific — ✅ — `www.magnific.com/app/api/community/creator-info-by-name?creator_name=%s&lang=en_US` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Martech — ✅ — `martech.org/author/%s/` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] MAX — ✅ — `max.ru/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Mintme — ✅ — `www.mintme.com/token/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Mssg.Me — ✅ — `%s.mssg.me/` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] MyLot — ✅ — `www.mylot.com/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] MySpace — ✅ — `myspace.com/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] N8n Community — ✅ — `community.n8n.io/u/%s/summary` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] Needrom — ✅ — `www.needrom.com/author/%s/` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] NICommunityForum — ✅ — `community.native-instruments.com/profile/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Nifty Gateway — ✅ — `api.niftygateway.com/user/profile-and-offchain-nifties-by-url/?profile_url=%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Nightbot — ✅ — `api.nightbot.tv/1/channels/t/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Nih Buat Jajan — ✅ — `www.nihbuatjajan.com/manifest/%s.json` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Note — ✅ — `note.com/api/v2/creators/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Nothing — ✅ — `nothing.community/u/%s` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] Oglaszamy24h.Pl — ✅ — `oglaszamy24h.pl/profil,%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] OK.ru — ✅ — `ok.ru/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] omg.lol — ✅ — `api.omg.lol/address/%s/info` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] OpenGameArt — ✅ — `opengameart.org/users/%s` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] OpenSource — ✅ — `opensource.com/users/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] OpenStreetMap — ✅ — `www.openstreetmap.org/user/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Patronite — ✅ — `patronite.pl/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Planoly — ✅ — `planoly.store/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Platzi — ✅ — `platzi.com/p/%s/` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] Playstrategy — ✅ — `playstrategy.org/@/%s` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] Pokec — ✅ — `pokec.azet.sk/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] pr0gramm — ✅ — `pr0gramm.com/api/profile/info?name=%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Pravda.Me — ✅ — `pravda.me/@%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Privacy Guides — ✅ — `discuss.privacyguides.net/u/%s.json` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Pronouns.Page — ✅ — `pronouns.page/api/profile/get/%s?version=2` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Pychess — ✅ — `www.pychess.org/@/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Python Discussions — ✅ — `discuss.python.org/u/%s/summary` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Rclone Forum — ✅ — `forum.rclone.org/u/%s` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] Refsheet — ✅ — `refsheet.net/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Revolut — ✅ — `revolut.me/api/web-profile/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Risk.Ru — ✅ — `risk.ru/people/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Roberts Space Industries — ✅ — `robertsspaceindustries.com/community-hub/user/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] SEOClerks — ✅ — `www.seoclerks.com/user/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] SEOForum — ✅ — `seoforum.com/@%s` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] Sessionize — ✅ — `sessionize.com/%s` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] Shikimori — ✅ — `shikimori.io/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] ShitpostBot5000 — ✅ — `www.shitpostbot.com/user/%s` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] Signal Community — ✅ — `community.signalusers.org/u/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Slides — ✅ — `slides.com/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] SmashRun — ✅ — `smashrun.com/%s/` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] SmugMug — ✅ — `%s.smugmug.com` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Social.Bund.De — ✅ — `social.bund.de/@%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] SpaceHey — ✅ — `spacehey.com/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Speaker Deck — ✅ — `speakerdeck.com/%s/` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Spells8 — ✅ — `forum.spells8.com/u/%s` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] Spletnik — ✅ — `spletnik.ru/user/%s` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] Status Cafe — ✅ — `status.cafe/users/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] StreamElements — ✅ — `api.streamelements.com/kappa/v2/channels/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] SublimeForum — ✅ — `forum.sublimetext.com/u/%s` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] Suzuri — ✅ — `suzuri.jp/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] SWAPD — ✅ — `swapd.co/u/%s` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] Tanuki.Pl — ✅ — `tanuki.pl/profil/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Tenor — ✅ — `tenor.com/users/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] TETR.IO — ✅ — `ch.tetr.io/api/users/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] The Guardian — ✅ — `www.theguardian.com/profile/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] ThemeForest — ✅ — `themeforest.net/user/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] TipeeeStream — ✅ — `www.tipeeestream.com/v3.0/pages/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] TRAKTRAIN — ✅ — `traktrain.com/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Träwelling — ✅ — `traewelling.de/@%s` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] Twitcast — ✅ — `frontendapi.twitcasting.tv/users/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Twpro — ✅ — `twpro.jp/%s/q-data.json` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] uwu.ai — ✅ — `%s.uwu.ai/` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Uwumarket — ✅ — `uwumarket.us/collections/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Valorant Forums — ✅ — `valorantforums.com/u/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Velog — ✅ — `velog.io/@%s/posts` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] Vero — ✅ — `vero.co/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Vi Bilägare — ✅ — `www.vibilagare.se/users/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Vivino — ✅ — `api.vivino.com/users/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] VLR — ✅ — `www.vlr.gg/user/%s` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] VseTop — ✅ — `vsetop.org/user/%s/` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Wago Addons — ✅ — `addons.wago.io/user/%s` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] Warmerise — ✅ — `warmerise.com/profile/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] WebNode — ✅ — `%s.webnode.cz/` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)
- [x] WeTransfer — ✅ — `wepresent.wetransfer.com/artists/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Wimkin — ✅ — `wimkin.com/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Wowhead — ✅ — `www.wowhead.com/user=%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Wykop — ✅ — `wykop.pl/ludzie/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Xakep.Ru — ✅ — `xakep.ru/author/%s/` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] XING — ✅ — `www.xing.com/profile/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] YesWeHack — ✅ — `api.yeswehack.com/hunters/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] YouPic — ✅ — `youpic.com/photographer/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Zenn — ✅ — `zenn.dev/%s` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Zhihu — ✅ — `api.zhihu.com/books/people/%s/publications?offset=0&limit=5` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)
- [x] Zomato — ✅ — `www.zomato.com/%s/reviews` (keyless profile: 200 taken / 404 free; status-only, HTML Accept)

- [x] Algorand NFD — ✅ — `api.nf.domains/nfd/%s.algo` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept; no documented limit)
- [x] Ansible Galaxy — ✅ — `galaxy.ansible.com/api/v3/namespaces/%s/` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept; no documented limit)
- [x] ArtifactHub — ✅ — `artifacthub.io/api/v1/orgs/%s` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept; no documented limit)
- [x] CRAN — ✅ — `cran.r-project.org/web/packages/%s/index.html` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; no documented limit)
- [x] CTAN — ✅ — `ctan.org/pkg/%s` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; no documented limit)
- [x] Factorio Mod Portal — ✅ — `mods.factorio.com/api/mods/%s` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept; no documented limit)
- [x] Grafana Plugins — ✅ — `grafana.com/api/plugins/%s` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept; no documented limit)
- [x] Homebrew Cask — ✅ — `formulae.brew.sh/api/cask/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept; no documented limit)
- [x] Jenkins Plugins — ✅ — `plugins.jenkins.io/%s/` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; no documented limit)
- [x] LuaRocks — ✅ — `luarocks.org/modules/%s` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; no documented limit)
- [x] Modrinth — ✅ — `api.modrinth.com/v2/project/%s` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept; no documented limit)
- [x] opam — ✅ — `opam.ocaml.org/packages/%s/` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; no documented limit)
- [x] packagecloud — ✅ — `packagecloud.io/%s` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; no documented limit)
- [x] PECL — ✅ — `pecl.php.net/package/%s` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; no documented limit)
- [x] Puppet Forge — ✅ — `forgeapi.puppet.com/v3/users/%s` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept; no documented limit)
- [x] Quay.io — ✅ — `quay.io/api/v1/organization/%s` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept; no documented limit)
- [x] Scoop — ✅ — `raw.githubusercontent.com/ScoopInstaller/Main/master/bucket/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept; no documented limit)
- [x] Sublime Package Control — ✅ — `packagecontrol.io/packages/%s` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; no documented limit)
- [x] Zenodo Communities — ✅ — `zenodo.org/api/communities/%s` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept; no documented limit)
- [x] Zora — ✅ — `zora.co/@%s` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; no documented limit)

- [x] Bear Blog — ✅ — `%s.bearblog.dev` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; subdomain-style blogs; 404 body is a plain 'not found' page)
- [x] Big Cartel — ✅ — `%s.bigcartel.com` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; artist storefront subdomains; 404 for unclaimed shops)
- [x] BookWyrm — ✅ — `bookwyrm.social/user/%s` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; federated Goodreads alternative; canonical instance flagships the software)
- [x] Brunch — ✅ — `brunch.co.kr/@%s` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; Korean long-form publishing (Kakao); @ prefix required)
- [x] D&D Beyond — ✅ — `www.dndbeyond.com/members/%s` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; tabletop RPG community; /members/ prefix required)
- [x] Fiverr — ✅ — `www.fiverr.com/%s` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; seller profile is the bare username path, not /sellers/…)
- [x] Freelancer.com — ✅ — `www.freelancer.com/u/%s` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; /u/ segment is required; bare /{account} is not a profile)
- [x] Hatena — ✅ — `profile.hatena.ne.jp/%s/` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; Hatena's shared profile namespace across its Japanese services)
- [x] Inkitt — ✅ — `www.inkitt.com/%s` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; fiction/wattpad-style platform; brand handle inkitt is a normal user profile)
- [x] Ko-fi — ✅ — `ko-fi.com/%s` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; creator tip jar; /ko-fi itself is not a user (404), so probe real creator handles)
- [x] Kompasiana — ✅ — `www.kompasiana.com/%s` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; Indonesia's largest citizen-blogging platform (Kompas Gramedia))
- [x] Qiita — ✅ — `qiita.com/%s` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; organisations (e.g. increments) 404 here; users 404 cleanly)
- [x] Scratch — ✅ — `scratch.mit.edu/users/%s/` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; MIT kids' coding community; trailing slash matters)
- [x] TabNews — ✅ — `www.tabnews.com.br/%s` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; Brazilian HN-style community founded by Filipe Deschamps)
- [x] Tapas — ✅ — `tapas.io/%s` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; webcomic/webnovel platform; /notion is 404 so names are real profiles)

- [x] Cash App — ✅ — `cash.app/$%s` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept)
- [x] GeeksforGeeks — ✅ — `authapi.geeksforgeeks.org/api-get/user-profile-info/?handle=%s` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] HackAdvisor — ✅ — `hackadvisor.io/api/v2/profile/%s/` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] imood — ✅ — `www.imood.com/users/%s` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept)
- [x] Tistory — ✅ — `%s.tistory.com/` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept)

- [x] Baserow Community — ✅ — `community.baserow.io/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] Bitwarden Community — ✅ — `community.bitwarden.com/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] Celo Forum — ✅ — `forum.celo.org/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] community.influxdata.com — ✅ — `community.influxdata.com/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] Cosmos Forum — ✅ — `forum.cosmos.network/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] Crystal Forum — ✅ — `forum.crystal-lang.org/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] Cursor Forum — ✅ — `forum.cursor.com/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] e Foundation Community — ✅ — `community.e.foundation/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] Elixir Forum — ✅ — `forum.elixirforum.com/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] EndeavourOS Forum — ✅ — `forum.endeavouros.com/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] Garuda Linux Forum — ✅ — `forum.garudalinux.org/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] Ghost Forum — ✅ — `forum.ghost.org/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] Home Assistant Community — ✅ — `community.home-assistant.io/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] Kotlin Discussions — ✅ — `discuss.kotlinlang.org/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] Kubernetes Discuss — ✅ — `discuss.kubernetes.io/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] Lets Encrypt Community — ✅ — `community.letsencrypt.org/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] Logseq Forum — ✅ — `discuss.logseq.com/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] Mattermost Community — ✅ — `forum.mattermost.com/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] NocoDB Community — ✅ — `community.nocodb.com/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] Obsidian Forum — ✅ — `forum.obsidian.md/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] OctoPrint Community — ✅ — `community.octoprint.org/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] OpenAI Community — ✅ — `community.openai.com/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] openHAB Community — ✅ — `community.openhab.org/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] OpenZeppelin Forum — ✅ — `forum.openzeppelin.com/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] Polkadot Forum — ✅ — `forum.polkadot.network/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] PyTorch Discuss — ✅ — `discuss.pytorch.org/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] Rasa Community — ✅ — `forum.rasa.com/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] Ruby on Rails Discussions — ✅ — `discuss.rubyonrails.org/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] Rust Internals (lang) — ✅ — `internals.rust-lang.org/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] Rust Users — ✅ — `users.rust-lang.org/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] Solana Forum — ✅ — `forum.solana.com/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] Traefik Community — ✅ — `community.traefik.io/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] Vercel Community — ✅ — `community.vercel.com/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)

- [x] Ameba Blog — ✅ — `ameblo.jp/%s/` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; Japanese blog host)
- [x] Ghost — ✅ — `%s.ghost.io/` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; Ghost(Pro) publication subdomains)
- [x] Hatena Blog — ✅ — `%s.hatenablog.com/` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; Japanese blog host)
- [x] Livedoor Blog — ✅ — `%s.livedoor.blog/` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; Japanese blog host)
- [x] Softr — ✅ — `%s.softr.app/` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; no-code site subdomains)

- [x] Blender Video — ✅ — `video.blender.org/api/v1/accounts/%s` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept; PeerTube instance)
- [x] Lemmy.ml — ✅ — `lemmy.ml/api/v3/user?username=%s` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept; second Lemmy instance next to lemmy.world)
- [x] TILvids — ✅ — `tilvids.com/api/v1/accounts/%s` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept; PeerTube instance)
- [x] TON DNS — ✅ — `tonapi.io/v2/dns/%s.ton` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept; third-party TON API; names are 4+ characters)

- [x] Alire (Ada) — ✅ — `alire.ada.dev/crates/%s` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; no documented limit)
- [x] Bioconductor — ✅ — `bioconductor.org/packages/release/bioc/html/%s.html` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; no documented limit; release/bioc path only, annotation/experiment packages use other paths)
- [x] ConanCenter — ✅ — `conan.io/center/recipes/%s` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; no documented limit)
- [x] Fedora dist-git — ✅ — `src.fedoraproject.org/rpms/%s` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; no documented limit)
- [x] GNU ELPA — ✅ — `elpa.gnu.org/packages/%s.html` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; no documented limit)
- [x] Haxe Library — ✅ — `lib.haxe.org/p/%s/` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; no documented limit)
- [x] Hugo Themes — ✅ — `themes.gohugo.io/themes/%s/` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; no documented limit)
- [x] Keybase — ✅ — `keybase.io/%s` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; no documented limit)
- [x] MacPorts — ✅ — `ports.macports.org/api/v1/ports/%s/` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept; no documented limit)
- [x] Node-RED Library — ✅ — `flows.nodered.org/node/%s` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; no documented limit)
- [x] NonGNU ELPA — ✅ — `elpa.nongnu.org/nongnu/%s.html` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; no documented limit)
- [x] Shopify App Store — ✅ — `apps.shopify.com/%s` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; no documented limit)
- [x] Stackage — ✅ — `www.stackage.org/package/%s` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; no documented limit)
- [x] Zapier Apps — ✅ — `zapier.com/apps/%s/integrations` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; no documented limit)

- [x] Apollo GraphQL Community — ✅ — `community.apollographql.com/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] Audacity Forum — ✅ — `forum.audacityteam.org/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] Django Forum — ✅ — `forum.djangoproject.com/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] fast.ai Forums — ✅ — `forums.fast.ai/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] Godot Forum — ✅ — `forum.godotengine.org/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] Golang Bridge Forum — ✅ — `forum.golangbridge.org/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] Julia Discourse — ✅ — `discourse.julialang.org/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] Krita Artists — ✅ — `krita-artists.org/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] Plotly Community — ✅ — `community.plotly.com/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] QIIME 2 Forum — ✅ — `forum.qiime2.org/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] Shotcut Forum — ✅ — `forum.shotcut.org/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] Zorin Forum — ✅ — `forum.zorin.com/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)

- [x] Cargo — ✅ — `%s.cargo.site/` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; portfolio subdomains)

- [x] Audius — ✅ — `api.audius.co/v1/users/handle/%s` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept; Music distribution; JSON API returns 404 'no rows in result set' for free handles.)
- [x] Bionluk — ✅ — `bionluk.com/%s` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; Turkish freelance-services marketplace; slug is the seller username.)
- [x] Blockdit — ✅ — `www.blockdit.com/%s` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; Thai long-form creator platform; known page is Blockdit's official profile.)
- [x] Cnblogs — ✅ — `www.cnblogs.com/%s` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; China dev blogging; 404 body is the site's own 404 page, no echo.)
- [x] Douban — ✅ — `www.douban.com/people/%s/` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; China books/film/social profile; known handle is Douban's founder.)
- [x] Exblog — ✅ — `%s.exblog.jp` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; Excite Japan blog host; missing subdomain returns the Excite 404 page.)
- [x] Instamojo — ✅ — `www.instamojo.com/%s` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; India digital-goods/store storefront; 404 for unclaimed store slugs.)
- [x] Kakuyomu — ✅ — `kakuyomu.jp/users/%s` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; Kadokawa web-novel platform; string author handles, not numeric ids.)
- [x] Mataroa — ✅ — `%s.mataroa.blog` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; Minimalist no-JS blog host; 404 body is Mataroa's 'Page Not Found' page.)
- [x] Pensador — ✅ — `www.pensador.com/autor/%s/` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; Brazilian quotes/poetry author pages; 404 title is '404 - Pensador'.)
- [x] POSTYPE — ✅ — `www.postype.com/@%s` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; Korean creator/novel/comic subscription platform; @handle namespace.)
- [x] Ruby China — ✅ — `ruby-china.org/%s` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; Chinese Ruby developer community profile; open-source Rails app.)
- [x] Teachable — ✅ — `%s.teachable.com` (keyless namespace: 200 taken / 404 free; status-only, HTML Accept; Course-creator school subdomain; known handle is Teachable's own school.)

- [x] Defold Forum — ✅ — `forum.defold.com/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] GDevelop Forum — ✅ — `forum.gdevelop.io/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] LangChain Forum — ✅ — `forum.langchain.com/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] OpenCV Forum — ✅ — `forum.opencv.org/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] Perplexity Community — ✅ — `community.perplexity.ai/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] Pinecone Community — ✅ — `community.pinecone.io/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)
- [x] Ultralytics Community — ✅ — `community.ultralytics.com/u/%s.json` (keyless namespace: 200 taken / 404 free; status-only, JSON Accept)

- [x] Google AI Developers — ✅ — `discuss.ai.google.dev/u/{name}.json` (keyless profile: 200 taken / 404 free; status-only, JSON Accept)

## Candidates

### The 2026-11-24 scout triage

This replaced the `similar-sites.md` scout list (1,476 rows, never verified).
163 were keyed, 287 had no claimable route (search boxes, opaque ids) and 1,026
had a handle or subdomain shape; 173 of those already appeared here or in the
registry, leaving 853 to probe — 692 handle routes over HTTP and 161 subdomains
over DNS. Of the handle routes 123 split on `google` against `zzqxnvc`; after
dropping soft-404s, bot walls, redirect-only diffs and 13 routes another source
owns, 82 rows survived (77 handles + 5 subdomain hosts) and became providers
the same day. Everything below is what did not survive a second look: five
deferred, seven ruled out, two dropped as already covered. The remainder of the
853 never had a signal at all — 155 shells, 102 challenges, 22 blocked, 34 dead
and 243 that answer 404 for every label.

- [ ] Canva — 🕸 — `canva.com/{name}`: creator profiles exist, but every keyless
  request, for a real and a fake label alike, comes back 403, so there is no
  signal to read; re-probe before porting
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
- [ ] USPTO (trademarks) — 🔌 — class-based, fuzzy-similarity trademark search; not a
  single claimable label, so it cannot answer name availability. Re-probed
  2026-11-25: `tsdrapi.uspto.gov/ts/cd/casestatus/…` is now **keyed** (401,
  "you'll need to register for an API key"; free MyUSPTO account), and the new
  `tmsearch.uspto.gov` search is an SPA with no public JSON route
  (`/api/v1/trademarks/search` is a 404 from the static host). A mark is
  registered per Nice class and blocks on similarity, so the verdict would be a
  class-scoped count, not taken/free
- [ ] EUIPO (trademarks) — 🔌 — class-based trademark search; not a single
  claimable label, so it cannot answer name availability. Re-probed 2026-11-25:
  `api.euipo.europa.eu/trademark-search/trademarks` is real but OAuth2
  client-credentials only (401, "Invalid client id or secret"; free EUIPO
  account), while the `eSearch` web route resets the connection for a non-browser
  client. Same class/similarity limit as USPTO
- [ ] TMview (EUIPO aggregator) — 🔌 — trademark registers by class, not a name
  namespace; re-probed 2026-11-25: `tmdn.org/tmview/api/…` is rejected by the WAF
  ("Request Rejected"), so even a scraping route has no signal
- [ ] WIPO Global Brand Database — 🕸 — re-probed 2026-11-25: 200 HTML gated by an
  altcha proof-of-work challenge, so no keyless signal and still class-based
  search underneath
- [ ] SEC EDGAR — 🔌 — keyless (descriptive User-Agent) but not a claimable namespace: it indexes issuer/filing names, which you cannot register
- [ ] Uneed — 🔌 — ruled out: `uneed.best/tool/{name}` is a product-directory slug submitted by makers, not a user namespace; `google` is only a queued listing ("Launching Soon", no launch date), so a 200 records that some product once took the slug and not that anyone holds the label
- [ ] Vidyard — 🔌 — ruled out: `vidyard.com/{name}` is the WordPress marketing site, and `google` only 301s to an uploaded media file (`/wp-content/uploads/google.svg`, 200 `image/svg+xml`) while an arbitrary label 301s to `www.vidyard.com/{name}`, which 404s — the split is a static-asset artifact, and public Vidyard pages are addressed by opaque ids, not a claimable handle
- [ ] webhook.site — 🔌 — ruled out: `webhook.site/{name}` is a per-request webhook token, not a vanity namespace. `webhook.site/google` 200s with a 145-byte default response ("This URL has no default content configured. View in Webhook.site", pointing at token `ded31719-…`) while a missing token 404s ("The URL was deleted, or expired automatically") — a 200 only records that a live capture URL currently exists at that path, and free URLs expire after a week, so taken/free is not a stable identity claim
- [ ] Workana — 🔌 — ruled out: `workana.com/freelancer/{name}` is a display-name lookup, not the profile key — `google` 302s twice to `www.workana.com/freelancer/8ccdac2706be99d451443c62eb99619d`, an opaque id, while an unclaimed name 404s. The verdict would hinge on the redirect target, and the label itself is neither unique nor claimable
- [ ] Zeerk — 🔌 — ruled out: `zeerk.com/{name}` is a keyword search, not a username namespace — `google` 301s to a matching gig page (`/job/lead-generation/google-ads-adwords-ppc-campaign-setup-from-scratch-for-more-leads-sales/`, whose title is about Google Ads), while `microsoft`, `acme` and a fake label 404. A 200 means some service listing matched the query, not that anyone holds the label
- [ ] Home Assistant Integrations — 🔌 — ruled out: `home-assistant.io/integrations/{name}` is the documentation page for a component id that already exists in Home Assistant's source tree (`/integrations/hue/`), not a label anyone claims; a 200 only means the integration ships, so there is nothing to register and the 200/404 split is a docs index
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
- [x] Tumblr — ✅ — `{name}.tumblr.com` (200 taken / 404 free; re-probed 2026-11-25, the subdomain host now answers without a challenge; a blog on a custom domain 302s and reads as unknown rather than free)
- [x] Vimeo — ✅ — `vimeo.com/{name}` (keyless profile: 200 taken / 404 free; status-only, the 403 wall noted in 2026-11 is gone as of 2026-11-25)
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
- [x] Itch.io — ✅ — `itch.io/profile/{name}` (200 taken / 404 free; the `{name}.itch.io` host stays behind a Cloudflare challenge, the profile route does not; the earlier reading only probed the host)
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
