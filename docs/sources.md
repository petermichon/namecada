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
| ddns.net  | `{name}.ddns.net` (DNS, no wildcard)                           | resolves taken / NXDOMAIN free |
| ddnsgeek.com | `{name}.ddnsgeek.com` (DNS, no wildcard)                    | resolves taken / NXDOMAIN free |
| hopto.org | `{name}.hopto.org` (DNS, no wildcard)                          | resolves taken / NXDOMAIN free |
| myftp.org | `{name}.myftp.org` (DNS, no wildcard)                          | resolves taken / NXDOMAIN free |
| zapto.org | `{name}.zapto.org` (DNS, no wildcard)                          | resolves taken / NXDOMAIN free |
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
| LinkedIn Company | `linkedin.com/company/{name}` (company)                 | 200 taken / 404 free      |
| Loops | `loops.video/@{name}` (profile)                                | 200 taken / 404 free      |
| LyricsTranslate | `lyricstranslate.com/en/translator/{name}` (translator) | 200 taken / 404 free      |
| Mixlr | `mixlr.com/{name}` (broadcaster)                             | 200 taken / 404 free |
| Moj | `mojapp.in/@{name}` (profile)                                  | 200 taken / 404 free      |
| MusicBrainz | `musicbrainz.org/user/{name}` (editor)                    | 200 taken / 404 free      |
| Musixmatch | `musixmatch.com/artist/{name}` (artist)                    | 200 taken / 404 free      |
| MyDramaList | `mydramalist.com/profile/{name}` (profile)                 | 200 taken / 404 free      |
| Naver Blog | `blog.naver.com/{name}` (blog)                             | 200 taken / 404 free      |
| NuGet profiles | `nuget.org/profiles/{name}` (profile)                  | 200 taken / 404 free      |
| Open Library | `openlibrary.org/people/{name}` (reader profile)         | 200 taken / 404 free      |
| Owler | `owler.com/company/{name}` (company)                     | 200 taken / 404 free      |
| Pillowfort | `pillowfort.social/{name}` (profile)                    | 200 taken / 404 free      |
| Pixelfed | `pixelfed.social/{name}` (profile)                        | 200 taken / 404 free      |
| PlayCode | `playcode.io/{name}` (playground project)                 | 200 taken / 404 free      |
| Poshmark | `poshmark.com/closet/{name}` (closet)                    | 200 taken / 404 free      |
| Racket Packages | `pkgs.racket-lang.org/package/{name}` (package)    | 200 taken / 404 free      |
| ReverbNation | `reverbnation.com/{name}` (artist profile)             | 200 taken / 404 free      |
| RubyGems profiles | `rubygems.org/profiles/{name}` (profile)          | 200 taken / 404 free      |
| Skool | `skool.com/{name}` (community)                                 | 200 taken / 404 free      |
| Slashdot | `slashdot.org/~{name}` (user journal)                       | 200 taken / 404 free      |
| Spotify | `open.spotify.com/user/{name}` (user profile)               | 200 taken / 404 free      |
| stats.fm | `stats.fm/{name}` (listening profile)                      | 200 taken / 404 free      |
| Steemit | `steemit.com/@{name}` (blog)                                | 200 taken / 404 free      |
| Streamlabs | `streamlabs.com/{name}` (tip page)                        | 200 taken / 404 free      |
| SuperRare | `superrare.com/{name}` (artist profile)                   | 200 taken / 404 free      |
| Tildes | `tildes.net/user/{name}` (user profile)                       | 200 taken / 404 free      |
| Tipeee | `tipeee.com/{name}` (creator page)                            | 200 taken / 404 free      |
| V2EX | `v2ex.com/member/{name}` (member profile)                     | 200 taken / 404 free      |
| Visual Studio Marketplace | `marketplace.visualstudio.com/publishers/{name}` (publisher) | 200 taken / 404 free |
| Wattpad | `wattpad.com/user/{name}` (user profile)                      | 200 taken / 404 free      |
| Wellfound | `wellfound.com/company/{name}` (company page)               | 200 taken / 404 free      |
| Whatnot | `whatnot.com/user/{name}` (user profile)                      | 200 taken / 404 free      |
| Whop | `whop.com/{name}` (community page)                                | 200 taken / 404 free      |
| Write.as | `write.as/{name}` (blog)                                       | 200 taken / 404 free      |
| ArtStation | `artstation.com/api/v2/user_profiles/{name}.json`             | 200 taken / 404 free      |
| AtCoder | `atcoder.jp/users/{name}` (profile)                             | 200 taken / 404 free      |
| Bugcrowd | `bugcrowd.com/profile-service/v1/profiles/{name}`              | 200 taken / 404 free      |
| Carrd | `{name}.carrd.co`                                               | 200 taken / 404 free      |
| Crowdin | `crowdin.com/profile/{name}` (profile)                          | 200 taken / 404 free      |
| Docker Hub | `hub.docker.com/u/{name}` (user)                              | 200 taken / 404 free      |
| Freesound | `freesound.org/people/{name}/` (profile)                      | 200 taken / 404 free      |
| HackerRank | `hackerrank.com/rest/contests/master/hackers/{name}/profile` | 200 taken / 404 free      |
| Itch.io | `itch.io/profile/{name}` (profile)                              | 200 taken / 404 free      |
| Linktree | `linktr.ee/{name}` (profile)                                   | 200 taken / 404 free; needs an HTML Accept |
| Monkeytype | `api.monkeytype.com/users/{name}/profile`                    | 200 taken / 404 free      |
| MyAnimeList | `myanimelist.net/profile/{name}` (profile)                   | 200 taken / 404 free      |
| npm orgs | `registry.npmjs.org/-/org/{name}/user` (organisation)          | 200 taken / 404 free      |
| osu! | `osu.ppy.sh/users/{name}` (profile)                                | 200 taken / 404 free      |
| Pastebin | `pastebin.com/u/{name}` (profile)                              | 200 taken / 404 free      |
| PyPI orgs | `pypi.org/org/{name}/` (organisation)                         | 200 taken / 404 free      |
| TradingView | `tradingview.com/u/{name}/` (profile)                        | 200 taken / 404 free      |
| Trello | `trello.com/1/Members/{name}`                                    | 200 taken / 404 free      |
| Venmo | `venmo.com/u/{name}` (profile)                                   | 200 taken / 404 free      |
| Vimeo | `vimeo.com/{name}` (profile)                                     | 200 taken / 404 free      |
| WakaTime | `wakatime.com/api/v1/users/{name}`                              | 200 taken / 404 free      |

| 35photo | `35photo.pro/@%s/` (profile) | 200 taken / 404 free |
| 7dach | `7dach.ru/profile/%s` (profile) | 200 taken / 404 free |
| Airbit | `airbit.com/%s` (profile) | 200 taken / 404 free |
| Alik | `www.alik.cz/u/%s` (profile) | 200 taken / 404 free |
| AmericanThinker | `www.americanthinker.com/author/%s/` (profile) | 200 taken / 404 free |
| Aparat | `www.aparat.com/api/fa/v1/user/user/information/username/%s` (profile) | 200 taken / 404 free |
| Arduino (Forum) | `forum.arduino.cc/u/%s.json` (profile) | 200 taken / 404 free |
| Artbreeder | `www.artbreeder.com/%s` (profile) | 200 taken / 404 free |
| Asciinema | `asciinema.org/~%s` (profile) | 200 taken / 404 free |
| AudioJungle | `audiojungle.net/user/%s` (profile) | 200 taken / 404 free |
| BandLab | `www.bandlab.com/api/v1.3/users/%s` (profile) | 200 taken / 404 free |
| Bezuzyteczna | `bezuzyteczna.pl/uzytkownicy/%s` (profile) | 200 taken / 404 free |
| Bimpos | `ask.bimpos.com/user/%s` (profile) | 200 taken / 404 free |
| Bio Site | `bio.site/%s` (profile) | 200 taken / 404 free |
| BioHacking | `forum.dangerousthings.com/u/%s` (profile) | 200 taken / 404 free |
| Blipfoto | `www.blipfoto.com/%s` (profile) | 200 taken / 404 free |
| Blitz Tactics | `blitztactics.com/%s` (profile) | 200 taken / 404 free |
| Boosty | `api.boosty.to/v1/blog/%s` (profile) | 200 taken / 404 free |
| Bunpro | `community.bunpro.jp/u/%s.json` (profile) | 200 taken / 404 free |
| BuzzFeed | `www.buzzfeed.com/%s` (profile) | 200 taken / 404 free |
| Caddy Community | `caddy.community/u/%s/summary` (profile) | 200 taken / 404 free |
| Calendy | `calendly.com/api/booking/profiles/%s` (profile) | 200 taken / 404 free |
| Car Talk Community | `community.cartalk.com/u/%s/summary` (profile) | 200 taken / 404 free |
| Carbonmade | `%s.carbonmade.com/` (profile) | 200 taken / 404 free |
| Career.Habr | `career.habr.com/%s` (profile) | 200 taken / 404 free |
| Cfx.re | `forum.cfx.re/u/%s.json` (profile) | 200 taken / 404 free |
| Championat | `www.championat.com/user/%s/` (profile) | 200 taken / 404 free |
| Chatujme.Cz | `profil.chatujme.cz/%s` (profile) | 200 taken / 404 free |
| Choice Community | `choice.community/u/%s/summary` (profile) | 200 taken / 404 free |
| Choko.Link | `choko.link/%s` (profile) | 200 taken / 404 free |
| Cnet | `www.cnet.com/profiles/%s/` (profile) | 200 taken / 404 free |
| Coda | `coda.io/@%s/` (profile) | 200 taken / 404 free |
| Codeforces | `codeforces.com/api/user.info?handles=%s` (profile) | 200 taken / 404 free |
| Codementor | `www.codementor.io/@%s` (profile) | 200 taken / 404 free |
| Commudle | `json.commudle.com/api/v2/users?username=%s` (profile) | 200 taken / 404 free |
| Coub | `coub.com/api/v2/channels/%s` (profile) | 200 taken / 404 free |
| Credly | `www.credly.com/users/%s` (profile) | 200 taken / 404 free |
| Cropty | `api.cropty.io/v1/auth/%s` (profile) | 200 taken / 404 free |
| Cryptomator Forum | `community.cryptomator.org/u/%s` (profile) | 200 taken / 404 free |
| Cytoid | `cytoid.io/profile/%s` (profile) | 200 taken / 404 free |
| Daily Kos | `www.dailykos.com/user/%s` (profile) | 200 taken / 404 free |
| DC Inside | `gallog.dcinside.com/%s` (profile) | 200 taken / 404 free |
| Destream | `api.destream.net/siteapi/v2/live/details/%s` (profile) | 200 taken / 404 free |
| devRant | `devrant.com/api/get-user-id?app=3&username=%s` (profile) | 200 taken / 404 free |
| Diablo | `diablo2.io/member/%s/` (profile) | 200 taken / 404 free |
| DIBIZ | `www.dibiz.com/%s` (profile) | 200 taken / 404 free |
| Discourse | `meta.discourse.org/u/%s/summary.json` (profile) | 200 taken / 404 free |
| Discuss.Elastic.Co | `discuss.elastic.co/u/%s` (profile) | 200 taken / 404 free |
| Dojoverse | `dojoverse.com/members/%s/` (profile) | 200 taken / 404 free |
| Donatello | `donatello.to/%s` (profile) | 200 taken / 404 free |
| eToro | `www.etoro.com/api/logininfo/v1.1/users/%s` (profile) | 200 taken / 404 free |
| Fixya | `www.fixya.com/users/%s` (profile) | 200 taken / 404 free |
| Flipboard | `flipboard.com/@%s` (profile) | 200 taken / 404 free |
| Flowcode | `www.flow.page/%s` (profile) | 200 taken / 404 free |
| Football | `www.rusfootball.info/user/%s/` (profile) | 200 taken / 404 free |
| FreeCodeCamp | `api.freecodecamp.org/users/get-public-profile?username=%s` (profile) | 200 taken / 404 free |
| Freelance.RU | `freelance.ru/%s` (profile) | 200 taken / 404 free |
| Freelance.Ua | `freelance.ua/user/%s/` (profile) | 200 taken / 404 free |
| Game Jolt | `gamejolt.com/site-api/web/profile/@%s/` (profile) | 200 taken / 404 free |
| Geocaching | `www.geocaching.com/p/?u=%s` (profile) | 200 taken / 404 free |
| Gradle | `plugins.gradle.org/u/%s` (profile) | 200 taken / 404 free |
| Habtium | `habtium.es/%s` (profile) | 200 taken / 404 free |
| Hackaday.Io | `hackaday.io/%s` (profile) | 200 taken / 404 free |
| HackMD | `hackmd.io/@%s` (profile) | 200 taken / 404 free |
| Hackster | `www.hackster.io/%s` (profile) | 200 taken / 404 free |
| HiHello | `www.hihello.com/author/%s` (profile) | 200 taken / 404 free |
| Hoo.Be | `hoo.be/%s` (profile) | 200 taken / 404 free |
| Houzz | `www.houzz.com/user/%s` (profile) | 200 taken / 404 free |
| Icons8 Community | `community.icons8.com/u/%s/summary` (profile) | 200 taken / 404 free |
| Icq-Chat | `icq.icqchat.co/members/%s/` (profile) | 200 taken / 404 free |
| ILGM Growers Forum | `ilgmforum.com/u/%s/summary.json` (profile) | 200 taken / 404 free |
| ImageShack | `imageshack.com/user/%s` (profile) | 200 taken / 404 free |
| Inaturalist | `api.inaturalist.org/v1/users/%s` (profile) | 200 taken / 404 free |
| Instructables | `www.instructables.com/json-api/showAuthorExists?screenName=%s` (profile) | 200 taken / 404 free |
| Intigriti | `app.intigriti.com/api/user/public/profile/%s` (profile) | 200 taken / 404 free |
| Ionic Forum | `forum.ionicframework.com/u/%s` (profile) | 200 taken / 404 free |
| Issuu | `issuu.com/%s` (profile) | 200 taken / 404 free |
| Joplin Forum | `discourse.joplinapp.org/u/%s` (profile) | 200 taken / 404 free |
| Jupyter Community Forum | `discourse.jupyter.org/u/%s/summary` (profile) | 200 taken / 404 free |
| Komi | `api.komi.io/api/talent/usernames/%s` (profile) | 200 taken / 404 free |
| Laracast | `laracasts.com/@%s` (profile) | 200 taken / 404 free |
| Leasehackr | `forum.leasehackr.com/u/%s/summary/` (profile) | 200 taken / 404 free |
| LessWrong | `www.lesswrong.com/users/%s` (profile) | 200 taken / 404 free |
| Linux.Org.Ru | `www.linux.org.ru/people/%s/profile` (profile) | 200 taken / 404 free |
| Livejournal | `%s.livejournal.com` (profile) | 200 taken / 404 free |
| LottieFiles | `lottiefiles.com/%s` (profile) | 200 taken / 404 free |
| Magnific | `www.magnific.com/app/api/community/creator-info-by-name?creator_name=%s&lang=en_US` (profile) | 200 taken / 404 free |
| Martech | `martech.org/author/%s/` (profile) | 200 taken / 404 free |
| MAX | `max.ru/%s` (profile) | 200 taken / 404 free |
| Mintme | `www.mintme.com/token/%s` (profile) | 200 taken / 404 free |
| Mssg.Me | `%s.mssg.me/` (profile) | 200 taken / 404 free |
| MyLot | `www.mylot.com/%s` (profile) | 200 taken / 404 free |
| MySpace | `myspace.com/%s` (profile) | 200 taken / 404 free |
| N8n Community | `community.n8n.io/u/%s/summary` (profile) | 200 taken / 404 free |
| Needrom | `www.needrom.com/author/%s/` (profile) | 200 taken / 404 free |
| NICommunityForum | `community.native-instruments.com/profile/%s` (profile) | 200 taken / 404 free |
| Nifty Gateway | `api.niftygateway.com/user/profile-and-offchain-nifties-by-url/?profile_url=%s` (profile) | 200 taken / 404 free |
| Nightbot | `api.nightbot.tv/1/channels/t/%s` (profile) | 200 taken / 404 free |
| Nih Buat Jajan | `www.nihbuatjajan.com/manifest/%s.json` (profile) | 200 taken / 404 free |
| Note | `note.com/api/v2/creators/%s` (profile) | 200 taken / 404 free |
| Nothing | `nothing.community/u/%s` (profile) | 200 taken / 404 free |
| Oglaszamy24h.Pl | `oglaszamy24h.pl/profil,%s` (profile) | 200 taken / 404 free |
| OK.ru | `ok.ru/%s` (profile) | 200 taken / 404 free |
| omg.lol | `api.omg.lol/address/%s/info` (profile) | 200 taken / 404 free |
| OpenGameArt | `opengameart.org/users/%s` (profile) | 200 taken / 404 free |
| OpenSource | `opensource.com/users/%s` (profile) | 200 taken / 404 free |
| OpenStreetMap | `www.openstreetmap.org/user/%s` (profile) | 200 taken / 404 free |
| Patronite | `patronite.pl/%s` (profile) | 200 taken / 404 free |
| Planoly | `planoly.store/%s` (profile) | 200 taken / 404 free |
| Platzi | `platzi.com/p/%s/` (profile) | 200 taken / 404 free |
| Playstrategy | `playstrategy.org/@/%s` (profile) | 200 taken / 404 free |
| Pokec | `pokec.azet.sk/%s` (profile) | 200 taken / 404 free |
| pr0gramm | `pr0gramm.com/api/profile/info?name=%s` (profile) | 200 taken / 404 free |
| Pravda.Me | `pravda.me/@%s` (profile) | 200 taken / 404 free |
| Privacy Guides | `discuss.privacyguides.net/u/%s.json` (profile) | 200 taken / 404 free |
| Pronouns.Page | `pronouns.page/api/profile/get/%s?version=2` (profile) | 200 taken / 404 free |
| Pychess | `www.pychess.org/@/%s` (profile) | 200 taken / 404 free |
| Python Discussions | `discuss.python.org/u/%s/summary` (profile) | 200 taken / 404 free |
| Rclone Forum | `forum.rclone.org/u/%s` (profile) | 200 taken / 404 free |
| Refsheet | `refsheet.net/%s` (profile) | 200 taken / 404 free |
| Revolut | `revolut.me/api/web-profile/%s` (profile) | 200 taken / 404 free |
| Risk.Ru | `risk.ru/people/%s` (profile) | 200 taken / 404 free |
| Roberts Space Industries | `robertsspaceindustries.com/community-hub/user/%s` (profile) | 200 taken / 404 free |
| SEOClerks | `www.seoclerks.com/user/%s` (profile) | 200 taken / 404 free |
| SEOForum | `seoforum.com/@%s` (profile) | 200 taken / 404 free |
| Sessionize | `sessionize.com/%s` (profile) | 200 taken / 404 free |
| Shikimori | `shikimori.io/%s` (profile) | 200 taken / 404 free |
| ShitpostBot5000 | `www.shitpostbot.com/user/%s` (profile) | 200 taken / 404 free |
| Signal Community | `community.signalusers.org/u/%s` (profile) | 200 taken / 404 free |
| Slides | `slides.com/%s` (profile) | 200 taken / 404 free |
| SmashRun | `smashrun.com/%s/` (profile) | 200 taken / 404 free |
| SmugMug | `%s.smugmug.com` (profile) | 200 taken / 404 free |
| Social.Bund.De | `social.bund.de/@%s` (profile) | 200 taken / 404 free |
| SpaceHey | `spacehey.com/%s` (profile) | 200 taken / 404 free |
| Speaker Deck | `speakerdeck.com/%s/` (profile) | 200 taken / 404 free |
| Spells8 | `forum.spells8.com/u/%s` (profile) | 200 taken / 404 free |
| Spletnik | `spletnik.ru/user/%s` (profile) | 200 taken / 404 free |
| Status Cafe | `status.cafe/users/%s` (profile) | 200 taken / 404 free |
| StreamElements | `api.streamelements.com/kappa/v2/channels/%s` (profile) | 200 taken / 404 free |
| SublimeForum | `forum.sublimetext.com/u/%s` (profile) | 200 taken / 404 free |
| Suzuri | `suzuri.jp/%s` (profile) | 200 taken / 404 free |
| SWAPD | `swapd.co/u/%s` (profile) | 200 taken / 404 free |
| Tanuki.Pl | `tanuki.pl/profil/%s` (profile) | 200 taken / 404 free |
| Tenor | `tenor.com/users/%s` (profile) | 200 taken / 404 free |
| TETR.IO | `ch.tetr.io/api/users/%s` (profile) | 200 taken / 404 free |
| The Guardian | `www.theguardian.com/profile/%s` (profile) | 200 taken / 404 free |
| ThemeForest | `themeforest.net/user/%s` (profile) | 200 taken / 404 free |
| TipeeeStream | `www.tipeeestream.com/v3.0/pages/%s` (profile) | 200 taken / 404 free |
| TRAKTRAIN | `traktrain.com/%s` (profile) | 200 taken / 404 free |
| Träwelling | `traewelling.de/@%s` (profile) | 200 taken / 404 free |
| Twitcast | `frontendapi.twitcasting.tv/users/%s` (profile) | 200 taken / 404 free |
| Twpro | `twpro.jp/%s/q-data.json` (profile) | 200 taken / 404 free |
| uwu.ai | `%s.uwu.ai/` (profile) | 200 taken / 404 free |
| Uwumarket | `uwumarket.us/collections/%s` (profile) | 200 taken / 404 free |
| Valorant Forums | `valorantforums.com/u/%s` (profile) | 200 taken / 404 free |
| Velog | `velog.io/@%s/posts` (profile) | 200 taken / 404 free |
| Vero | `vero.co/%s` (profile) | 200 taken / 404 free |
| Vi Bilägare | `www.vibilagare.se/users/%s` (profile) | 200 taken / 404 free |
| Vivino | `api.vivino.com/users/%s` (profile) | 200 taken / 404 free |
| VLR | `www.vlr.gg/user/%s` (profile) | 200 taken / 404 free |
| VseTop | `vsetop.org/user/%s/` (profile) | 200 taken / 404 free |
| Wago Addons | `addons.wago.io/user/%s` (profile) | 200 taken / 404 free |
| Warmerise | `warmerise.com/profile/%s` (profile) | 200 taken / 404 free |
| WebNode | `%s.webnode.cz/` (profile) | 200 taken / 404 free |
| WeTransfer | `wepresent.wetransfer.com/artists/%s` (profile) | 200 taken / 404 free |
| Wimkin | `wimkin.com/%s` (profile) | 200 taken / 404 free |
| Wowhead | `www.wowhead.com/user=%s` (profile) | 200 taken / 404 free |
| Wykop | `wykop.pl/ludzie/%s` (profile) | 200 taken / 404 free |
| Xakep.Ru | `xakep.ru/author/%s/` (profile) | 200 taken / 404 free |
| XING | `www.xing.com/profile/%s` (profile) | 200 taken / 404 free |
| YesWeHack | `api.yeswehack.com/hunters/%s` (profile) | 200 taken / 404 free |
| YouPic | `youpic.com/photographer/%s` (profile) | 200 taken / 404 free |
| Zenn | `zenn.dev/%s` (profile) | 200 taken / 404 free |
| Zhihu | `api.zhihu.com/books/people/%s/publications?offset=0&limit=5` (profile) | 200 taken / 404 free |
| Zomato | `www.zomato.com/%s/reviews` (profile) | 200 taken / 404 free |
| Algorand NFD | `api.nf.domains/nfd/%s.algo` (namespace) | 200 taken / 404 free |
| Ansible Galaxy | `galaxy.ansible.com/api/v3/namespaces/%s/` (namespace) | 200 taken / 404 free |
| ArtifactHub | `artifacthub.io/api/v1/orgs/%s` (namespace) | 200 taken / 404 free |
| CRAN | `cran.r-project.org/web/packages/%s/index.html` (namespace) | 200 taken / 404 free |
| CTAN | `ctan.org/pkg/%s` (namespace) | 200 taken / 404 free |
| Factorio Mod Portal | `mods.factorio.com/api/mods/%s` (namespace) | 200 taken / 404 free |
| Grafana Plugins | `grafana.com/api/plugins/%s` (namespace) | 200 taken / 404 free |
| Homebrew Cask | `formulae.brew.sh/api/cask/%s.json` (namespace) | 200 taken / 404 free |
| Jenkins Plugins | `plugins.jenkins.io/%s/` (namespace) | 200 taken / 404 free |
| LuaRocks | `luarocks.org/modules/%s` (namespace) | 200 taken / 404 free |
| Modrinth | `api.modrinth.com/v2/project/%s` (namespace) | 200 taken / 404 free |
| opam | `opam.ocaml.org/packages/%s/` (namespace) | 200 taken / 404 free |
| packagecloud | `packagecloud.io/%s` (namespace) | 200 taken / 404 free |
| PECL | `pecl.php.net/package/%s` (namespace) | 200 taken / 404 free |
| Puppet Forge | `forgeapi.puppet.com/v3/users/%s` (namespace) | 200 taken / 404 free |
| Quay.io | `quay.io/api/v1/organization/%s` (namespace) | 200 taken / 404 free |
| Scoop | `raw.githubusercontent.com/ScoopInstaller/Main/master/bucket/%s.json` (namespace) | 200 taken / 404 free |
| Sublime Package Control | `packagecontrol.io/packages/%s` (namespace) | 200 taken / 404 free |
| Zenodo Communities | `zenodo.org/api/communities/%s` (namespace) | 200 taken / 404 free |
| Zora | `zora.co/@%s` (namespace) | 200 taken / 404 free |
| Bear Blog | `%s.bearblog.dev` (namespace) | 200 taken / 404 free |
| Big Cartel | `%s.bigcartel.com` (namespace) | 200 taken / 404 free |
| BookWyrm | `bookwyrm.social/user/%s` (namespace) | 200 taken / 404 free |
| Brunch | `brunch.co.kr/@%s` (namespace) | 200 taken / 404 free |
| D&D Beyond | `www.dndbeyond.com/members/%s` (namespace) | 200 taken / 404 free |
| Fiverr | `www.fiverr.com/%s` (namespace) | 200 taken / 404 free |
| Freelancer.com | `www.freelancer.com/u/%s` (namespace) | 200 taken / 404 free |
| Hatena | `profile.hatena.ne.jp/%s/` (namespace) | 200 taken / 404 free |
| Inkitt | `www.inkitt.com/%s` (namespace) | 200 taken / 404 free |
| Ko-fi | `ko-fi.com/%s` (namespace) | 200 taken / 404 free |
| Kompasiana | `www.kompasiana.com/%s` (namespace) | 200 taken / 404 free |
| Qiita | `qiita.com/%s` (namespace) | 200 taken / 404 free |
| Scratch | `scratch.mit.edu/users/%s/` (namespace) | 200 taken / 404 free |
| TabNews | `www.tabnews.com.br/%s` (namespace) | 200 taken / 404 free |
| Tapas | `tapas.io/%s` (namespace) | 200 taken / 404 free |
| Cash App | `cash.app/$%s` (namespace) | 200 taken / 404 free |
| GeeksforGeeks | `authapi.geeksforgeeks.org/api-get/user-profile-info/?handle=%s` (namespace) | 200 taken / 404 free |
| HackAdvisor | `hackadvisor.io/api/v2/profile/%s/` (namespace) | 200 taken / 404 free |
| imood | `www.imood.com/users/%s` (namespace) | 200 taken / 404 free |
| Tistory | `%s.tistory.com/` (namespace) | 200 taken / 404 free |
| Baserow Community | `community.baserow.io/u/%s.json` (namespace) | 200 taken / 404 free |
| Bitwarden Community | `community.bitwarden.com/u/%s.json` (namespace) | 200 taken / 404 free |
| Celo Forum | `forum.celo.org/u/%s.json` (namespace) | 200 taken / 404 free |
| community.influxdata.com | `community.influxdata.com/u/%s.json` (namespace) | 200 taken / 404 free |
| Cosmos Forum | `forum.cosmos.network/u/%s.json` (namespace) | 200 taken / 404 free |
| Crystal Forum | `forum.crystal-lang.org/u/%s.json` (namespace) | 200 taken / 404 free |
| Cursor Forum | `forum.cursor.com/u/%s.json` (namespace) | 200 taken / 404 free |
| e Foundation Community | `community.e.foundation/u/%s.json` (namespace) | 200 taken / 404 free |
| Elixir Forum | `forum.elixirforum.com/u/%s.json` (namespace) | 200 taken / 404 free |
| EndeavourOS Forum | `forum.endeavouros.com/u/%s.json` (namespace) | 200 taken / 404 free |
| Garuda Linux Forum | `forum.garudalinux.org/u/%s.json` (namespace) | 200 taken / 404 free |
| Ghost Forum | `forum.ghost.org/u/%s.json` (namespace) | 200 taken / 404 free |
| Home Assistant Community | `community.home-assistant.io/u/%s.json` (namespace) | 200 taken / 404 free |
| Kotlin Discussions | `discuss.kotlinlang.org/u/%s.json` (namespace) | 200 taken / 404 free |
| Kubernetes Discuss | `discuss.kubernetes.io/u/%s.json` (namespace) | 200 taken / 404 free |
| Lets Encrypt Community | `community.letsencrypt.org/u/%s.json` (namespace) | 200 taken / 404 free |
| Logseq Forum | `discuss.logseq.com/u/%s.json` (namespace) | 200 taken / 404 free |
| Mattermost Community | `forum.mattermost.com/u/%s.json` (namespace) | 200 taken / 404 free |
| NocoDB Community | `community.nocodb.com/u/%s.json` (namespace) | 200 taken / 404 free |
| Obsidian Forum | `forum.obsidian.md/u/%s.json` (namespace) | 200 taken / 404 free |
| OctoPrint Community | `community.octoprint.org/u/%s.json` (namespace) | 200 taken / 404 free |
| OpenAI Community | `community.openai.com/u/%s.json` (namespace) | 200 taken / 404 free |
| openHAB Community | `community.openhab.org/u/%s.json` (namespace) | 200 taken / 404 free |
| OpenZeppelin Forum | `forum.openzeppelin.com/u/%s.json` (namespace) | 200 taken / 404 free |
| Polkadot Forum | `forum.polkadot.network/u/%s.json` (namespace) | 200 taken / 404 free |
| PyTorch Discuss | `discuss.pytorch.org/u/%s.json` (namespace) | 200 taken / 404 free |
| Rasa Community | `forum.rasa.com/u/%s.json` (namespace) | 200 taken / 404 free |
| Ruby on Rails Discussions | `discuss.rubyonrails.org/u/%s.json` (namespace) | 200 taken / 404 free |
| Rust Internals (lang) | `internals.rust-lang.org/u/%s.json` (namespace) | 200 taken / 404 free |
| Rust Users | `users.rust-lang.org/u/%s.json` (namespace) | 200 taken / 404 free |
| Solana Forum | `forum.solana.com/u/%s.json` (namespace) | 200 taken / 404 free |
| Traefik Community | `community.traefik.io/u/%s.json` (namespace) | 200 taken / 404 free |
| Vercel Community | `community.vercel.com/u/%s.json` (namespace) | 200 taken / 404 free |
| Ameba Blog | `ameblo.jp/%s/` (namespace) | 200 taken / 404 free |
| Ghost | `%s.ghost.io/` (namespace) | 200 taken / 404 free |
| Hatena Blog | `%s.hatenablog.com/` (namespace) | 200 taken / 404 free |
| Livedoor Blog | `%s.livedoor.blog/` (namespace) | 200 taken / 404 free |
| Softr | `%s.softr.app/` (namespace) | 200 taken / 404 free |
| Tumblr | `{name}.tumblr.com` (blog) | 200 taken / 404 free; a custom-domain blog 302s (unknown) |
| Blender Video | `video.blender.org/api/v1/accounts/%s` (namespace) | 200 taken / 404 free |
| Lemmy.ml | `lemmy.ml/api/v3/user?username=%s` (namespace) | 200 taken / 404 free |
| TILvids | `tilvids.com/api/v1/accounts/%s` (namespace) | 200 taken / 404 free |
| TON DNS | `tonapi.io/v2/dns/%s.ton` (namespace) | 200 taken / 404 free |
| Alire (Ada) | `alire.ada.dev/crates/%s` (namespace) | 200 taken / 404 free |
| Bioconductor | `bioconductor.org/packages/release/bioc/html/%s.html` (namespace) | 200 taken / 404 free |
| ConanCenter | `conan.io/center/recipes/%s` (namespace) | 200 taken / 404 free |
| Fedora dist-git | `src.fedoraproject.org/rpms/%s` (namespace) | 200 taken / 404 free |
| GNU ELPA | `elpa.gnu.org/packages/%s.html` (namespace) | 200 taken / 404 free |
| Haxe Library | `lib.haxe.org/p/%s/` (namespace) | 200 taken / 404 free |
| Hugo Themes | `themes.gohugo.io/themes/%s/` (namespace) | 200 taken / 404 free |
| Keybase | `keybase.io/%s` (namespace) | 200 taken / 404 free |
| MacPorts | `ports.macports.org/api/v1/ports/%s/` (namespace) | 200 taken / 404 free |
| Node-RED Library | `flows.nodered.org/node/%s` (namespace) | 200 taken / 404 free |
| NonGNU ELPA | `elpa.nongnu.org/nongnu/%s.html` (namespace) | 200 taken / 404 free |
| Shopify App Store | `apps.shopify.com/%s` (namespace) | 200 taken / 404 free |
| Stackage | `www.stackage.org/package/%s` (namespace) | 200 taken / 404 free |
| Zapier Apps | `zapier.com/apps/%s/integrations` (namespace) | 200 taken / 404 free |
| Apollo GraphQL Community | `community.apollographql.com/u/%s.json` (namespace) | 200 taken / 404 free |
| Audacity Forum | `forum.audacityteam.org/u/%s.json` (namespace) | 200 taken / 404 free |
| Django Forum | `forum.djangoproject.com/u/%s.json` (namespace) | 200 taken / 404 free |
| fast.ai Forums | `forums.fast.ai/u/%s.json` (namespace) | 200 taken / 404 free |
| Godot Forum | `forum.godotengine.org/u/%s.json` (namespace) | 200 taken / 404 free |
| Golang Bridge Forum | `forum.golangbridge.org/u/%s.json` (namespace) | 200 taken / 404 free |
| Julia Discourse | `discourse.julialang.org/u/%s.json` (namespace) | 200 taken / 404 free |
| Krita Artists | `krita-artists.org/u/%s.json` (namespace) | 200 taken / 404 free |
| Plotly Community | `community.plotly.com/u/%s.json` (namespace) | 200 taken / 404 free |
| QIIME 2 Forum | `forum.qiime2.org/u/%s.json` (namespace) | 200 taken / 404 free |
| Shotcut Forum | `forum.shotcut.org/u/%s.json` (namespace) | 200 taken / 404 free |
| Zorin Forum | `forum.zorin.com/u/%s.json` (namespace) | 200 taken / 404 free |
| Cargo | `%s.cargo.site/` (namespace) | 200 taken / 404 free |
| Audius | `api.audius.co/v1/users/handle/%s` (namespace) | 200 taken / 404 free |
| Bionluk | `bionluk.com/%s` (namespace) | 200 taken / 404 free |
| Blockdit | `www.blockdit.com/%s` (namespace) | 200 taken / 404 free |
| Cnblogs | `www.cnblogs.com/%s` (namespace) | 200 taken / 404 free |
| Douban | `www.douban.com/people/%s/` (namespace) | 200 taken / 404 free |
| Exblog | `%s.exblog.jp` (namespace) | 200 taken / 404 free |
| Instamojo | `www.instamojo.com/%s` (namespace) | 200 taken / 404 free |
| Kakuyomu | `kakuyomu.jp/users/%s` (namespace) | 200 taken / 404 free |
| Mataroa | `%s.mataroa.blog` (namespace) | 200 taken / 404 free |
| Pensador | `www.pensador.com/autor/%s/` (namespace) | 200 taken / 404 free |
| POSTYPE | `www.postype.com/@%s` (namespace) | 200 taken / 404 free |
| Ruby China | `ruby-china.org/%s` (namespace) | 200 taken / 404 free |
| Teachable | `%s.teachable.com` (namespace) | 200 taken / 404 free |
| Defold Forum | `forum.defold.com/u/%s.json` (namespace) | 200 taken / 404 free |
| GDevelop Forum | `forum.gdevelop.io/u/%s.json` (namespace) | 200 taken / 404 free |
| LangChain Forum | `forum.langchain.com/u/%s.json` (namespace) | 200 taken / 404 free |
| OpenCV Forum | `forum.opencv.org/u/%s.json` (namespace) | 200 taken / 404 free |
| Perplexity Community | `community.perplexity.ai/u/%s.json` (namespace) | 200 taken / 404 free |
| Pinecone Community | `community.pinecone.io/u/%s.json` (namespace) | 200 taken / 404 free |
| Ultralytics Community | `community.ultralytics.com/u/%s.json` (namespace) | 200 taken / 404 free |
| Google AI Developers | `discuss.ai.google.dev/u/{name}.json` (profile) | 200 taken / 404 free |
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
is-a-good.dev, Cloudflare Pages, Fly.io, ddns.net, ddnsgeek.com, hopto.org, myftp.org,
zapto.org) have no wildcard DNS, so they resolve `{name}.{host}` and treat a
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

