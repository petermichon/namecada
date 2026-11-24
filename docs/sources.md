# Sources

Every source Namecada checks, the request it makes and the rule that decides its
verdict.

| Source    | Request                                                        | Verdict                   |
| --------- | -------------------------------------------------------------- | ------------------------- |
| Domains   | RDAP registry via IANA bootstrap                               | 200 registered / 404 free |
| ENS       | `api.ensideas.com/ens/resolve/{name}.eth`                     | non-null `address` taken / `null` free |
| SNS       | `sdk-proxy.sns.id/resolve/{name}.sol`                          | `"s":"ok"` taken / `"s":"error"` free |
| HNS       | `hnsscan.com/api/names/{name}/details`                         | `info.state` `CLOSED` taken / `AVAILABLE`/`UNKNOWN` free |
| Namecoin  | `explorer.namecoin.co/api/name/d%2F{name}`                     | 404 free / 200 `expired:true` free, `false` taken |
| Lens      | `api.lens.xyz/graphql` `account(username.localName:)` (POST)   | non-null `account` taken / null free |
| npm       | `registry.npmjs.org/{name}`                                    | 200 taken / 404 free      |
| PyPI      | `pypi.org/pypi/{name}/json`                                    | 200 taken / 404 free      |
| crates.io | `crates.io/api/v1/crates/{name}`                               | 200 taken / 404 free      |
| NuGet     | `api.nuget.org/v3-flatcontainer/{name}/index.json`             | 200 taken / 404 free      |
| RubyGems  | `rubygems.org/api/v1/gems/{name}.json`                         | 200 taken / 404 free      |
| Hex       | `hex.pm/api/packages/{name}`                                   | 200 taken / 404 free      |
| Hex user  | `hex.pm/api/users/{name}`                                      | 200 taken / 404 free      |
| pub.dev   | `pub.dev/api/packages/{name}`                                  | 200 taken / 404 free      |
| Homebrew  | `formulae.brew.sh/api/formula/{name}.json`                     | 200 taken / 404 free      |
| Hackage   | `hackage.haskell.org/package/{name}`                           | 200 taken / 404 free      |
| Clojars   | `clojars.org/api/artifacts/{name}`                             | 200 taken / 404 free      |
| Chocolatey | `community.chocolatey.org/packages/{name}`                    | 200 taken / 404 free      |
| AUR       | `aur.archlinux.org/packages/{name}`                            | 200 taken / 404 free      |
| Snap Store | `api.snapcraft.io/v2/snaps/info/{name}` (`Snap-Device-Series: 16`) | 200 taken / 404 free  |
| Chef Supermarket | `supermarket.chef.io/api/v1/cookbooks/{name}`               | 200 taken / 404 free      |
| CocoaPods | `cocoapods.org/pods/{name}`                                   | 200 taken / 404 free      |
| CPAN      | `fastapi.metacpan.org/v1/module/{name}` (case-sensitive)      | 200 taken / 404 free      |
| Lobsters  | `lobste.rs/~{name}`                                           | 200 taken / 404 free      |
| Blogger   | `{name}.blogspot.com`                                         | 200 taken / 404 free      |
| Flickr    | `flickr.com/people/{name}/`                                   | 200 taken / 404 free      |
| StackBlitz | `stackblitz.com/@{name}` (HTML Accept)                       | 200 taken / 404 free      |
| Slack     | `{name}.slack.com`                                            | 200/403 taken / 404 free  |
| Firefox   | `addons.mozilla.org/api/v5/addons/addon/{name}/`               | 200/401 taken / 404 free  |
| JSR       | `jsr.io/api/scopes/{name}`                                     | 200/400 taken / 404 free  |
| WordPress | `api.wordpress.org/plugins/info/1.2/?request[slug]={name}`     | 200/closed taken / 404 free |
| Open VSX  | `open-vsx.org/api/{name}` (namespace)                          | 200 taken / 404 free      |
| GitHub    | `api.github.com/users/{name}`                                  | 200 taken / 404 free      |
| GitLab    | `gitlab.com/api/v4/users?username=` + `/groups/{name}`         | user or group = taken     |
| Bitbucket | `api.bitbucket.org/2.0/workspaces/{name}`                      | 200/403 taken / 404 free  |
| Codeberg  | `codeberg.org/api/v1/users/{name}`                             | 200 taken / 404 body      |
| SourceForge | `sourceforge.net/projects/{name}/`                           | 200 taken / 404 free      |
| Wikipedia | `meta.wikimedia.org/w/api.php` globaluserinfo               | `missing` = free          |
| DEV       | `dev.to/api/users/by_username` + `/organizations/{name}`       | 200 taken / 404 free      |
| Farcaster | `fnames.farcaster.xyz/transfers/current?name={name}`           | 200 taken / 404 free      |
| Minehut   | `api.minehut.com/server/{name}?byName=true`                    | `server: null` = free     |
| Lichess   | `lichess.org/api/user/{name}`                                  | 200 taken / 404 free      |
| Chess.com | `api.chess.com/pub/player/{name}`                              | 200 taken / 404 free      |
| Hugging Face | `huggingface.co/api/users/{name}/overview`                 | 200 taken / 404 free      |
| Dailymotion | `api.dailymotion.com/user/{name}`                            | 200 taken / 404 free      |
| Mixcloud  | `api.mixcloud.com/{name}/`                                     | 200 taken / 404 free      |
| SoundCloud | `soundcloud.com/{name}` (profile)                             | 200 taken / 404 free      |
| Bandcamp  | `{name}.bandcamp.com` (redirect target)                        | `/signup` redirect free / 200 or elsewhere taken |
| AniList   | `graphql.anilist.co` `User(name:)` (POST)                      | non-null `User` taken / null free |
| Roblox    | `users.roblox.com/v1/usernames/users` (POST)                   | non-empty `data` taken / empty free |
| Minecraft | `api.mojang.com/users/profiles/minecraft/{name}`              | 200 taken / 404 free      |
| Open Collective | `opencollective.com/{name}.json`                         | 200 taken / 404 free      |
| Liberapay | `liberapay.com/{name}`                                        | 200 taken / 404 free      |
| Bluesky   | `bsky.social/xrpc/com.atproto.temp.checkHandleAvailability` (`{name}.bsky.social`) | resultAvailable free / resultUnavailable taken |
| Mastodon  | `mastodon.social/@{name}` (profile)                            | 200/410 taken / 404 free  |
| YouTube   | `youtube.com/@{name}` (profile)                                | 200 taken / 404 free      |
| Kaggle    | `kaggle.com/{name}` (profile)                                  | 200 taken / 404 free      |
| skills.sh | `skills.sh/{name}` (publisher profile)                         | 200 taken / 404 free      |
| Dribbble  | `dribbble.com/{name}` (profile)                                | 200 taken / 404 free; a non-public profile 404s an "isn't public yet" body and is taken |
| Behance   | `behance.net/{name}` (profile)                                 | 200 taken / 404 free      |
| Replicate | `replicate.com/{name}` (profile)                               | 200 taken / 404 free      |
| DeviantArt | `deviantart.com/{name}` (profile)                             | 200 taken / 404 free      |
| Snapchat  | `snapchat.com/add/{name}` (profile)                            | 200 taken / 404 free      |
| Strava    | `strava.com/athletes/{name}` (profile)                         | 200 taken / 404 free      |
| Swift Package Index | `swiftpackageindex.com/{name}` (publisher)            | 200 taken / 404 free; publishing nothing also 404s |
| TikTok | `tiktok.com/oembed?url=…/@{name}` | 200 `author_url` taken / 400 free |
| Threads | `threads.net/@{name}` | `(@name)` in title taken / `Threads • Log in` free |
| Twitch | `twitch.tv/{name}` | `{name} - Twitch` taken / `Twitch` free |
| Duolingo | `duolingo.com/profile/{name}` | og:title `on Duolingo` taken / `Learn a language` free |
| Telegram | `t.me/{name}` | `Telegram: View @` taken / `Contact @` free |
| HackerNoon | `hackernoon.com/u/{name}` (profile)                           | 200 taken / 404 free      |
| Paragraph | `paragraph.com/@{name}` (profile)                              | 200 taken / 404 free      |
| Steam     | `steamcommunity.com/id/{name}?xml=1`                           | `<steamID64>` taken / `<error>` free |
| Instagram | `instagram.com/{name}/` (profile)                              | `"username":"{name}"` taken / generic shell free |
| Hacker News | `hacker-news.firebaseio.com/v0/user/{name}.json`             | `null` free / object with `id` taken |
| Giphy     | `giphy.com/{name}` (profile)                                   | 200 canonical `/{name}` taken / 301 → `/explore/{name}` free |
| Minds     | `api/v1/channel/{name}` (profile `minds.com/{name}/`)           | `status:"success"` taken / `ChannelNotFoundException` free |
| X / Twitter | `x.com/{name}` (profile)                                     | 200 taken / 404 free      |
| Vercel    | `{name}.vercel.app` (subdomain)                                | 200/401 (SSO) taken / 404 free |
| Netlify   | `{name}.netlify.app` (subdomain)                               | 200 taken / 404 free      |
| Surge     | `{name}.surge.sh` (subdomain)                                  | 200 taken / 404 free      |
| Firebase Hosting | `{name}.web.app` (subdomain)                            | 200 taken / 404 free      |
| Neocities | `{name}.neocities.org` (subdomain)                             | 200 taken / 404 free      |
| Render    | `{name}.onrender.com` (subdomain)                              | 200 taken / 404 free      |
| Railway   | `{name}.up.railway.app` (subdomain)                            | 200 taken / 404 free      |
| Heroku    | `{name}.herokuapp.com` (subdomain)                             | 200 taken / 404 free      |
| Read the Docs | `{name}.readthedocs.io` (subdomain)                        | 200 taken / 404 free      |
| Gumroad   | `{name}.gumroad.com` (subdomain)                               | 200 taken / 404 free      |
| Substack  | `{name}.substack.com` (subdomain)                              | 200 taken / 404 free      |
| Shopify   | `{name}.myshopify.com` (subdomain)                             | 200 taken / 404 free      |
| Weebly    | `{name}.weebly.com` (subdomain)                                | 200 taken / 404 free      |
| eu.org    | `{name}.eu.org` (DNS, no wildcard)                             | resolves taken / NXDOMAIN free |
| DuckDNS   | `{name}.duckdns.org` (DNS, no wildcard)                        | resolves taken / NXDOMAIN free |
| No-IP     | `{name}.no-ip.org` (DNS, no wildcard)                          | resolves taken / NXDOMAIN free |
| dynv6     | `{name}.dynv6.net` (DNS, no wildcard)                          | resolves taken / NXDOMAIN free |
| noted.lol | `{name}.noted.lol` (DNS, no wildcard)                          | resolves taken / NXDOMAIN free |
| is-a-good.dev | `{name}.is-a-good.dev` (DNS, no wildcard)                  | resolves taken / NXDOMAIN free |
| Cloudflare Pages | `{name}.pages.dev` (DNS, no wildcard)                    | resolves taken / NXDOMAIN free |
| Fly.io    | `{name}.fly.dev` (DNS, no wildcard)                            | resolves taken / NXDOMAIN free |
| 9GAG      | `9gag.com/u/{name}` (profile)                                  | 200 taken / 404 free      |
| About.me  | `about.me/{name}` (profile)                                    | 200 taken / 404 free      |
| Clubhouse | `clubhouse.com/@{name}` (profile)                              | 200 taken / 404 free      |
| Codewars  | `codewars.com/users/{name}` (profile, HTML Accept)             | 200 taken / 404 free      |
| Gravatar  | `gravatar.com/{name}` (profile)                                | 200 taken / 404 free      |
| Launchpad | `launchpad.net/~{name}` (profile)                              | 200 taken / 404 free      |
| Lichess Team | `lichess.org/team/{name}` (team)                            | 200 taken / 404 free      |
| AngelList | `angel.co/company/{name}` (company)                            | 200 taken / 404 free      |
| Buttondown | `buttondown.com/{name}` (newsletter)                          | 200 taken / 404 free      |
| Built In | `builtin.com/company/{name}` (company)                        | 200 taken / 404 free      |
| Disqus | `disqus.com/by/{name}` (profile)                               | 200 taken / 404 free      |
| Donorbox | `donorbox.org/{name}` (campaign)                             | 200 taken / 404 free      |
| Douyu | `douyu.com/{name}` (streamer)                                 | 200 taken / 404 free      |
| Drupal.org Projects | `drupal.org/project/{name}` (module or theme)       | 200 taken / 404 free      |
| Experts Exchange | `experts-exchange.com/members/{name}`                   | 200 taken / 404 free      |
| Gab | `gab.com/{name}` (profile)                                      | 200 taken / 404 free      |
| Gitea | `gitea.com/{name}` (profile)                                  | 200 taken / 404 free      |
| Gitee | `gitee.com/{name}` (profile)                                  | 200 taken / 404 free      |
| GOG | `gog.com/u/{name}` (profile)                                    | 200 taken / 404 free      |
| Goodreads | `goodreads.com/{name}` (profile)                            | 200 taken / 404 free      |
| House-Mixes | `house-mixes.com/profile/{name}` (DJ profile)               | 200 taken / 404 free      |
| Hype Machine | `hypem.com/{name}` (profile)                              | 200 taken / 404 free      |
| Indie Hackers | `indiehackers.com/{name}` (profile)                        | 200 taken / 404 free      |
| Kick | `kick.com/{name}` (channel)                                     | 200 taken / 404 free      |
| Komoot | `komoot.com/user/{name}` (profile)                            | 200 taken / 404 free      |
| Kongregate | `kongregate.com/accounts/{name}` (profile)                  | 200 taken / 404 free      |
| Libib | `libib.com/u/{name}` (profile)                                   | 200 taken / 404 free      |

Credential-gated sources (Discogs, Companies House, OpenCorporates) are not
currently included.

Domains run once per TLD (`.com`, `.net`, `.org`, `.dev`, `.io`, `.ai`, `.app`), each a
separate provider. Each TLD uses its registry's RDAP endpoint; the live IANA
bootstrap lookup is not wired into the core yet. Domains are **paid**
sources: a registration costs money, and the amount varies (premium names cost
more). ENS
(`.eth`) is likewise paid — registering costs ETH. SNS (`.sol`) is paid too —
registering costs SOL. HNS is paid as well — a new Handshake name is won in an
auction, and an expired one is reclaimed the same way. Namecoin (`.bit`) is paid
too — registering a `d/` name costs NMC. Heroku is also paid (its free
tier ended in November 2022) but is still surfaced. Every handle, package and
social source is **free** to register. Each source carries a `paid` flag; the
report marks paid rows and offers a cost filter, but otherwise renders one
alphabetical list.

The free-subdomain sources (eu.org, DuckDNS, No-IP, dynv6, noted.lol,
is-a-good.dev, Cloudflare Pages, Fly.io) have no wildcard DNS, so they resolve `{name}.{host}` and treat a
successful answer as taken and an NXDOMAIN as free. Because that is a DNS
lookup, the label is sent to the resolver — the same privacy exposure as
looking the name up directly.

GitHub users and organizations share a single namespace, so one `/users` lookup
covers both.

## Choosing a source

Ordered by preference when adding one:

1. Official keyless API — the default. Verify each on the taken/free response.
2. Body-parse a server-rendered page — fallback when there is no API: match a
   stable marker (e.g. `og:title`/`"username"`), never localized "not found"
   text; return `unknown` on ambiguity.
3. Keyed API — **currently omitted**: no row, request, or cache entry. Add one
   only when neither (1) nor (2) can answer and the source is high-value.
4. Headless-browser scraping — avoid; slow, brittle, usually bot-blocked. Never
   scrape behind a login.

## Semantics

- Domain rows report **registry state**, not registrability: a name may be
  reserved or premium and still show as free. Coverage follows the bootstrap.
  The RFC 6761 special-use labels (`example`, `test`, `invalid`, `localhost`)
  are always shown as unavailable, since some registries reserve them without
  creating an RDAP record.
- Handle and package rows mean "the name exists", not legal or brand
  availability.
- Some platforms enforce a minimum name length (Chess.com 3, Lichess 2); shorter
  names are shown as **unavailable** (not registrable) rather than free.

