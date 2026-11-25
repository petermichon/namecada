//! Status-only providers: the largest adapter family. A request URL plus a
//! status code table covers every source whose verdict is just the HTTP status.

use std::collections::BTreeMap;
use std::sync::Arc;

use crate::manifest::{Manifest, Source};
use crate::provider::{Error, HttpRequest, HttpResponse, Outcome, Provider, Status, Transport};

pub const DEFAULT_ACCEPT: &str = "application/json";
pub const HTML_ACCEPT: &str = "text/html";
/// USER_AGENT identifies the tool politely but carries no repository link, so
/// it cannot be correlated with the author. Browser-shaped strings are worse
/// than useless here: sites fingerprint the TLS stack too, and a Chrome claim
/// from a non-browser stack is blocked more often than an honest token.
pub const USER_AGENT: &str = "namecada/1.0";
pub const ACCEPT_LANGUAGE: &str = "en-US,en;q=0.9";

/// CodeMap maps HTTP status codes to a verdict.
#[derive(Debug, Clone, Copy)]
pub struct CodeMap {
    pub available: &'static [u16],
    pub taken: &'static [u16],
}

pub const DEFAULT_CODES: CodeMap = CodeMap {
    available: &[404],
    taken: &[200],
};

/// StatusSpec describes one status-only source's request and mapping. The
/// public name/URL come from the manifest, so only the API call lives here.
#[derive(Debug, Clone, Copy)]
pub struct StatusSpec {
    pub id: &'static str,
    pub request: &'static str,
    pub codes: CodeMap,
    pub accept: &'static str,
}

const fn spec(id: &'static str, request: &'static str) -> StatusSpec {
    StatusSpec {
        id,
        request,
        codes: DEFAULT_CODES,
        accept: DEFAULT_ACCEPT,
    }
}

const fn spec_with(
    id: &'static str,
    request: &'static str,
    codes: CodeMap,
    accept: &'static str,
) -> StatusSpec {
    StatusSpec {
        id,
        request,
        codes,
        accept,
    }
}

/// status_specs returns every provider this adapter currently covers.
pub fn status_specs() -> &'static [StatusSpec] {
    SPECS
}

static SPECS: &[StatusSpec] = &[
    spec("npm", "https://registry.npmjs.org/%s"),
    spec("pypi", "https://pypi.org/pypi/%s/json"),
    spec("crates", "https://crates.io/api/v1/crates/%s"),
    spec(
        "nuget",
        "https://api.nuget.org/v3-flatcontainer/%s/index.json",
    ),
    spec("rubygems", "https://rubygems.org/api/v1/gems/%s.json"),
    spec("hex", "https://hex.pm/api/packages/%s"),
    spec("hex-user", "https://hex.pm/api/users/%s"),
    spec("pub", "https://pub.dev/api/packages/%s"),
    spec("homebrew", "https://formulae.brew.sh/api/formula/%s.json"),
    spec("hackage", "https://hackage.haskell.org/package/%s"),
    spec("clojars", "https://clojars.org/api/artifacts/%s"),
    spec("chocolatey", "https://community.chocolatey.org/packages/%s"),
    spec("aur", "https://aur.archlinux.org/packages/%s"),
    spec("cocoapods", "https://cocoapods.org/pods/%s"),
    spec("cpan", "https://fastapi.metacpan.org/v1/module/%s"),
    spec("lobsters", "https://lobste.rs/~%s"),
    spec("blogger", "https://%s.blogspot.com/"),
    spec("flickr", "https://www.flickr.com/people/%s/"),
    spec_with(
        "stackblitz",
        "https://stackblitz.com/@%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "slack",
        "https://%s.slack.com/",
        CodeMap {
            available: &[404],
            taken: &[200, 403],
        },
        DEFAULT_ACCEPT,
    ),
    spec(
        "chef-supermarket",
        "https://supermarket.chef.io/api/v1/cookbooks/%s",
    ),
    spec("sourceforge", "https://sourceforge.net/projects/%s/"),
    spec("liberapay", "https://liberapay.com/%s"),
    spec(
        "minecraft",
        "https://api.mojang.com/users/profiles/minecraft/%s",
    ),
    spec_with(
        "firefox",
        "https://addons.mozilla.org/api/v5/addons/addon/%s/",
        CodeMap {
            available: &[404],
            taken: &[200, 401],
        },
        DEFAULT_ACCEPT,
    ),
    spec_with(
        "jsr",
        "https://jsr.io/api/scopes/%s",
        CodeMap {
            available: &[404],
            taken: &[200, 400],
        },
        DEFAULT_ACCEPT,
    ),
    spec("openvsx", "https://open-vsx.org/api/%s"),
    spec("hackernoon", "https://hackernoon.com/u/%s"),
    spec("paragraph", "https://paragraph.com/@%s"),
    spec("github", "https://api.github.com/users/%s"),
    spec_with(
        "bitbucket",
        "https://api.bitbucket.org/2.0/workspaces/%s",
        CodeMap {
            available: &[404],
            taken: &[200, 403],
        },
        DEFAULT_ACCEPT,
    ),
    spec(
        "farcaster",
        "https://fnames.farcaster.xyz/transfers/current?name=%s",
    ),
    spec("lichess", "https://lichess.org/api/user/%s"),
    spec("chesscom", "https://api.chess.com/pub/player/%s"),
    spec(
        "huggingface",
        "https://huggingface.co/api/users/%s/overview",
    ),
    spec("dailymotion", "https://api.dailymotion.com/user/%s"),
    spec("mixcloud", "https://api.mixcloud.com/%s/"),
    spec("soundcloud", "https://soundcloud.com/%s"),
    spec_with(
        "mastodon",
        "https://mastodon.social/@%s",
        CodeMap {
            available: &[404],
            taken: &[200, 410],
        },
        HTML_ACCEPT,
    ),
    spec("youtube", "https://www.youtube.com/@%s"),
    spec("kaggle", "https://www.kaggle.com/%s"),
    spec("skills", "https://www.skills.sh/%s"),
    spec("behance", "https://www.behance.net/%s"),
    spec("replicate", "https://replicate.com/%s"),
    spec("x", "https://x.com/%s"),
    spec("deviantart", "https://www.deviantart.com/%s"),
    spec("snapchat", "https://www.snapchat.com/add/%s"),
    spec("strava", "https://www.strava.com/athletes/%s"),
    spec("swift-package-index", "https://swiftpackageindex.com/%s"),
    spec("lemmy", "https://lemmy.world/api/v3/user?username=%s"),
    // A claimed deployment behind Vercel SSO answers 401 for a JSON request
    // (302 to /sso-api for HTML); free names still 404, so both are taken.
    spec_with(
        "vercel",
        "https://%s.vercel.app/",
        CodeMap {
            available: &[404],
            taken: &[200, 401, 302],
        },
        DEFAULT_ACCEPT,
    ),
    spec("netlify", "https://%s.netlify.app/"),
    spec("surge", "https://%s.surge.sh/"),
    spec("firebase", "https://%s.web.app/"),
    spec("neocities", "https://%s.neocities.org/"),
    spec("render", "https://%s.onrender.com/"),
    spec("railway", "https://%s.up.railway.app/"),
    spec("heroku", "https://%s.herokuapp.com/"),
    spec("readthedocs", "https://%s.readthedocs.io/"),
    spec("gumroad", "https://%s.gumroad.com/"),
    spec("substack", "https://%s.substack.com/"),
    spec("shopify", "https://%s.myshopify.com/"),
    spec("weebly", "https://%s.weebly.com/"),
    spec("9gag", "https://9gag.com/u/%s"),
    spec("about-me", "https://about.me/%s"),
    spec("clubhouse", "https://clubhouse.com/@%s"),
    spec_with(
        "codewars",
        "https://codewars.com/users/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("gravatar", "https://gravatar.com/%s"),
    spec("launchpad", "https://launchpad.net/~%s"),
    spec("lichess-team", "https://lichess.org/team/%s"),
    spec("angellist", "https://angel.co/company/%s"),
    spec("buttondown", "https://buttondown.com/%s"),
    spec("built-in", "https://builtin.com/company/%s"),
    spec("disqus", "https://disqus.com/by/%s"),
    spec("donorbox", "https://donorbox.org/%s"),
    spec("douyu", "https://douyu.com/%s"),
    // Modules and themes share one drupal.org/project/{name} namespace.
    spec("drupal-org-modules", "https://drupal.org/project/%s"),
    spec(
        "experts-exchange",
        "https://experts-exchange.com/members/%s",
    ),
    spec("gab", "https://gab.com/%s"),
    spec_with("gitea", "https://gitea.com/%s", DEFAULT_CODES, HTML_ACCEPT),
    // A JSON Accept answers 404 even for a claimed profile.
    spec_with("gitee", "https://gitee.com/%s", DEFAULT_CODES, HTML_ACCEPT),
    spec_with("gog", "https://gog.com/u/%s", DEFAULT_CODES, HTML_ACCEPT),
    // A JSON Accept is answered 406 even for a claimed profile.
    spec_with(
        "goodreads",
        "https://goodreads.com/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "house-mixes",
        "https://house-mixes.com/profile/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "hype-machine",
        "https://hypem.com/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "indie-hackers",
        "https://indiehackers.com/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with("kick", "https://kick.com/%s", DEFAULT_CODES, HTML_ACCEPT),
    spec_with(
        "komoot",
        "https://komoot.com/user/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "kongregate",
        "https://kongregate.com/accounts/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "libib",
        "https://libib.com/u/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "linkedin-company",
        "https://linkedin.com/company/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "loops",
        "https://loops.video/@%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "lyricstranslate",
        "https://lyricstranslate.com/en/translator/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "metacritic",
        "https://metacritic.com/user/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    // A JSON Accept is answered 406 even for a claimed profile.
    spec_with("mixlr", "https://mixlr.com/%s", DEFAULT_CODES, HTML_ACCEPT),
    spec_with("moj", "https://mojapp.in/@%s", DEFAULT_CODES, HTML_ACCEPT),
    spec_with(
        "musicbrainz",
        "https://musicbrainz.org/user/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "musixmatch",
        "https://musixmatch.com/artist/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "mydramalist",
        "https://mydramalist.com/profile/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "naver-blog",
        "https://blog.naver.com/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "nuget-profiles",
        "https://nuget.org/profiles/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "open-library",
        "https://openlibrary.org/people/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "owler",
        "https://owler.com/company/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "pillowfort",
        "https://pillowfort.social/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "pixelfed",
        "https://pixelfed.social/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "playcode",
        "https://playcode.io/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "poshmark",
        "https://poshmark.com/closet/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "racket-packages",
        "https://pkgs.racket-lang.org/package/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "reverbnation",
        "https://reverbnation.com/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "rubygems-profiles",
        "https://rubygems.org/profiles/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with("skool", "https://skool.com/%s", DEFAULT_CODES, HTML_ACCEPT),
    spec_with(
        "slashdot",
        "https://slashdot.org/~%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "spotify",
        "https://open.spotify.com/user/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "stats-fm",
        "https://stats.fm/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "steemit",
        "https://steemit.com/@%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "streamlabs",
        "https://streamlabs.com/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "superrare",
        "https://superrare.com/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "tildes",
        "https://tildes.net/user/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "tipeee",
        "https://tipeee.com/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "v2ex",
        "https://v2ex.com/member/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "visual-studio-marketplace",
        "https://marketplace.visualstudio.com/publishers/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "wattpad",
        "https://wattpad.com/user/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "wellfound",
        "https://wellfound.com/company/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "whatnot",
        "https://whatnot.com/user/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with("whop", "https://whop.com/%s", DEFAULT_CODES, HTML_ACCEPT),
    spec_with(
        "write-as",
        "https://write.as/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    // Namespaces that sit next to a source we already check: the handle is the
    // same string but a different claimable name, so it gets its own row.
    spec("docker-hub", "https://hub.docker.com/u/%s"),
    spec("github-orgs", "https://api.github.com/orgs/%s"),
    spec("npm-orgs", "https://registry.npmjs.org/-/org/%s/user"),
    spec("pypi-orgs", "https://pypi.org/org/%s/"),
    // Profiles added in the 2026-11-25 candidate pass.
    spec(
        "artstation",
        "https://www.artstation.com/api/v2/user_profiles/%s.json",
    ),
    spec("atcoder", "https://atcoder.jp/users/%s"),
    spec(
        "bugcrowd",
        "https://bugcrowd.com/profile-service/v1/profiles/%s",
    ),
    spec("carrd", "https://%s.carrd.co"),
    spec("crowdin", "https://crowdin.com/profile/%s"),
    spec("freesound", "https://freesound.org/people/%s/"),
    spec(
        "hackerrank",
        "https://www.hackerrank.com/rest/contests/master/hackers/%s/profile",
    ),
    spec("itchio", "https://itch.io/profile/%s"),
    spec_with(
        "linktree",
        "https://linktr.ee/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("monkeytype", "https://api.monkeytype.com/users/%s/profile"),
    spec("myanimelist", "https://myanimelist.net/profile/%s"),
    spec("osu", "https://osu.ppy.sh/users/%s"),
    spec("pastebin", "https://pastebin.com/u/%s"),
    spec("tradingview", "https://www.tradingview.com/u/%s/"),
    spec("trello", "https://trello.com/1/Members/%s"),
    spec("venmo", "https://venmo.com/u/%s"),
    spec("vimeo", "https://vimeo.com/%s"),
    spec("wakatime", "https://wakatime.com/api/v1/users/%s"),

    spec_with(
        "35photo",
        "https://35photo.pro/@%s/",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "7dach",
        "https://7dach.ru/profile/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("addons-wago-io", "https://addons.wago.io/user/%s"),
    spec("airbit", "https://airbit.com/%s"),
    spec_with(
        "alik",
        "https://www.alik.cz/u/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "americanthinker",
        "https://www.americanthinker.com/author/%s/",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "aparat",
        "https://www.aparat.com/api/fa/v1/user/user/information/username/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "arduino-forum",
        "https://forum.arduino.cc/u/%s.json",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "artbreeder",
        "https://www.artbreeder.com/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "asciinema",
        "https://asciinema.org/~%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "audiojungle",
        "https://audiojungle.net/user/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "bandlab",
        "https://www.bandlab.com/api/v1.3/users/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("bezuzyteczna", "https://bezuzyteczna.pl/uzytkownicy/%s"),
    spec_with(
        "bimpos",
        "https://ask.bimpos.com/user/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "bio-site",
        "https://bio.site/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("biohacking", "https://forum.dangerousthings.com/u/%s"),
    spec("blipfoto", "https://www.blipfoto.com/%s"),
    spec_with(
        "blitz-tactics",
        "https://blitztactics.com/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "boosty",
        "https://api.boosty.to/v1/blog/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "bunpro",
        "https://community.bunpro.jp/u/%s.json",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "buzzfeed",
        "https://www.buzzfeed.com/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("caddy-community", "https://caddy.community/u/%s/summary"),
    spec_with(
        "calendy",
        "https://calendly.com/api/booking/profiles/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("car-talk-community", "https://community.cartalk.com/u/%s/summary"),
    spec_with(
        "carbonmade",
        "https://%s.carbonmade.com/",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "career-habr",
        "https://career.habr.com/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "cfx-re",
        "https://forum.cfx.re/u/%s.json",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "championat",
        "https://www.championat.com/user/%s/",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "chatujme-cz",
        "https://profil.chatujme.cz/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("choice-community", "https://choice.community/u/%s/summary"),
    spec_with(
        "choko-link",
        "https://choko.link/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "cnet",
        "https://www.cnet.com/profiles/%s/",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "coda",
        "https://coda.io/@%s/",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "codeforces",
        "https://codeforces.com/api/user.info?handles=%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "codementor",
        "https://www.codementor.io/@%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "commudle",
        "https://json.commudle.com/api/v2/users?username=%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "coub",
        "https://coub.com/api/v2/channels/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("credly", "https://www.credly.com/users/%s"),
    spec_with(
        "cropty",
        "https://api.cropty.io/v1/auth/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("cryptomator-forum", "https://community.cryptomator.org/u/%s"),
    spec_with(
        "cytoid",
        "https://cytoid.io/profile/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "daily-kos",
        "https://www.dailykos.com/user/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("dcinside", "https://gallog.dcinside.com/%s"),
    spec_with(
        "destream",
        "https://api.destream.net/siteapi/v2/live/details/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "devrant",
        "https://devrant.com/api/get-user-id?app=3&username=%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "diablo",
        "https://diablo2.io/member/%s/",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "dibiz",
        "https://www.dibiz.com/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "discourse",
        "https://meta.discourse.org/u/%s/summary.json",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "discuss-elastic-co",
        "https://discuss.elastic.co/u/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "dojoverse",
        "https://dojoverse.com/members/%s/",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "donatello",
        "https://donatello.to/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "etoro",
        "https://www.etoro.com/api/logininfo/v1.1/users/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("fixya", "https://www.fixya.com/users/%s"),
    spec_with(
        "flipboard",
        "https://flipboard.com/@%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "flowcode",
        "https://www.flow.page/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "football",
        "https://www.rusfootball.info/user/%s/",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "freecodecamp",
        "https://api.freecodecamp.org/users/get-public-profile?username=%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "freelance-ru",
        "https://freelance.ru/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "freelance-ua",
        "https://freelance.ua/user/%s/",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "game-jolt",
        "https://gamejolt.com/site-api/web/profile/@%s/",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "geocaching",
        "https://www.geocaching.com/p/?u=%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "gettr",
        "https://gettr.com/api/s/uinf/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("gradle", "https://plugins.gradle.org/u/%s"),
    spec_with(
        "habtium",
        "https://habtium.es/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "hackaday-io",
        "https://hackaday.io/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("hackmd", "https://hackmd.io/@%s"),
    spec_with(
        "hackster",
        "https://www.hackster.io/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "hihello",
        "https://www.hihello.com/author/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "hoo-be",
        "https://hoo.be/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "houzz",
        "https://www.houzz.com/user/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("icons8-community", "https://community.icons8.com/u/%s/summary"),
    spec_with(
        "icq-chat",
        "https://icq.icqchat.co/members/%s/",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "ilgm-growers-forum",
        "https://ilgmforum.com/u/%s/summary.json",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "imageshack",
        "https://imageshack.com/user/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "inaturalist",
        "https://api.inaturalist.org/v1/users/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "instructables",
        "https://www.instructables.com/json-api/showAuthorExists?screenName=%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "intigriti",
        "https://app.intigriti.com/api/user/public/profile/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("ionic-forum", "https://forum.ionicframework.com/u/%s"),
    spec("issuu", "https://issuu.com/%s"),
    spec("joplin-forum", "https://discourse.joplinapp.org/u/%s"),
    spec_with(
        "jupyter-community-forum",
        "https://discourse.jupyter.org/u/%s/summary",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "komi",
        "https://api.komi.io/api/talent/usernames/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("laracast", "https://laracasts.com/@%s"),
    spec("leasehackr", "https://forum.leasehackr.com/u/%s/summary/"),
    spec("lesswrong", "https://www.lesswrong.com/users/%s"),
    spec_with(
        "linux-org-ru",
        "https://www.linux.org.ru/people/%s/profile",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "livejournal",
        "https://%s.livejournal.com",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("lottiefiles", "https://lottiefiles.com/%s"),
    spec_with(
        "magnific",
        "https://www.magnific.com/app/api/community/creator-info-by-name?creator_name=%s&lang=en_US",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "martech",
        "https://martech.org/author/%s/",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "max",
        "https://max.ru/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "mintme",
        "https://www.mintme.com/token/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "mssg-me",
        "https://%s.mssg.me/",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "mylot",
        "https://www.mylot.com/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "myspace",
        "https://myspace.com/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("n8n-community", "https://community.n8n.io/u/%s/summary"),
    spec("needrom", "https://www.needrom.com/author/%s/"),
    spec_with(
        "nicommunityforum",
        "https://community.native-instruments.com/profile/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "nifty-gateway",
        "https://api.niftygateway.com/user/profile-and-offchain-nifties-by-url/?profile_url=%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "nightbot",
        "https://api.nightbot.tv/1/channels/t/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "nih-buat-jajan",
        "https://www.nihbuatjajan.com/manifest/%s.json",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "note",
        "https://note.com/api/v2/creators/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("nothing-community", "https://nothing.community/u/%s"),
    spec_with(
        "oglaszamy24h-pl",
        "https://oglaszamy24h.pl/profil,%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "ok-ru",
        "https://ok.ru/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "omg-lol",
        "https://api.omg.lol/address/%s/info",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("opengameart", "https://opengameart.org/users/%s"),
    spec_with(
        "opensource",
        "https://opensource.com/users/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "openstreetmap",
        "https://www.openstreetmap.org/user/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "patronite",
        "https://patronite.pl/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "planoly",
        "https://planoly.store/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("platzi", "https://platzi.com/p/%s/"),
    spec("playstrategy", "https://playstrategy.org/@/%s"),
    spec_with(
        "pokec",
        "https://pokec.azet.sk/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "pr0gramm",
        "https://pr0gramm.com/api/profile/info?name=%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "pravda-me",
        "https://pravda.me/@%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "privacy-guides",
        "https://discuss.privacyguides.net/u/%s.json",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "pronouns-page",
        "https://pronouns.page/api/profile/get/%s?version=2",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "pronouny",
        "https://pronouny.xyz/api/users/profile/username/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "pychess",
        "https://www.pychess.org/@/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "python-org-discussions",
        "https://discuss.python.org/u/%s/summary",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("rclone-forum", "https://forum.rclone.org/u/%s"),
    spec_with(
        "refsheet",
        "https://refsheet.net/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "revolut",
        "https://revolut.me/api/web-profile/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "risk-ru",
        "https://risk.ru/people/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "roberts-space-industries",
        "https://robertsspaceindustries.com/community-hub/user/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "seoclerks",
        "https://www.seoclerks.com/user/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("seoforum", "https://seoforum.com/@%s"),
    spec("sessionize", "https://sessionize.com/%s"),
    spec_with(
        "shikimori",
        "https://shikimori.io/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("shitpostbot5000", "https://www.shitpostbot.com/user/%s"),
    spec_with(
        "signal",
        "https://community.signalusers.org/u/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "slides",
        "https://slides.com/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "smashrun",
        "https://smashrun.com/%s/",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "smugmug",
        "https://%s.smugmug.com",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "social-bund-de",
        "https://social.bund.de/@%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "spacehey",
        "https://spacehey.com/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "speaker-deck",
        "https://speakerdeck.com/%s/",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("spells8", "https://forum.spells8.com/u/%s"),
    spec("spletnik", "https://spletnik.ru/user/%s"),
    spec_with(
        "status-cafe",
        "https://status.cafe/users/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "streamelements",
        "https://api.streamelements.com/kappa/v2/channels/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("sublimeforum", "https://forum.sublimetext.com/u/%s"),
    spec_with(
        "suzuri",
        "https://suzuri.jp/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("swapd", "https://swapd.co/u/%s"),
    spec_with(
        "tanuki-pl",
        "https://tanuki.pl/profil/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "tenor",
        "https://tenor.com/users/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "tetr-io",
        "https://ch.tetr.io/api/users/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "theguardian",
        "https://www.theguardian.com/profile/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "themeforest",
        "https://themeforest.net/user/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "tipeeestream",
        "https://www.tipeeestream.com/v3.0/pages/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "traktrain",
        "https://traktrain.com/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("trawelling", "https://traewelling.de/@%s"),
    spec_with(
        "twitcast",
        "https://frontendapi.twitcasting.tv/users/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "twpro",
        "https://twpro.jp/%s/q-data.json",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "uwu-ai",
        "https://%s.uwu.ai/",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "uwumarket",
        "https://uwumarket.us/collections/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "valorant-forums",
        "https://valorantforums.com/u/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("velog", "https://velog.io/@%s/posts"),
    spec_with(
        "vero",
        "https://vero.co/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "vibilagare",
        "https://www.vibilagare.se/users/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "vivino",
        "https://api.vivino.com/users/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("vlr", "https://www.vlr.gg/user/%s"),
    spec_with(
        "vsetop",
        "https://vsetop.org/user/%s/",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "warmerise",
        "https://warmerise.com/profile/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("webnode", "https://%s.webnode.cz/"),
    spec_with(
        "wetransfer",
        "https://wepresent.wetransfer.com/artists/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "wimkin-publicprofile",
        "https://wimkin.com/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "wowhead",
        "https://www.wowhead.com/user=%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "wykop",
        "https://wykop.pl/ludzie/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "xakep-ru",
        "https://xakep.ru/author/%s/",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "xing",
        "https://www.xing.com/profile/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "yeswehack",
        "https://api.yeswehack.com/hunters/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "youpic",
        "https://youpic.com/photographer/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "zenn",
        "https://zenn.dev/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "zhihu",
        "https://api.zhihu.com/books/people/%s/publications?offset=0&limit=5",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "zomato",
        "https://www.zomato.com/%s/reviews",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("algorand-nfd", "https://api.nf.domains/nfd/%s.algo"),
    spec("ansible-galaxy", "https://galaxy.ansible.com/api/v3/namespaces/%s/"),
    spec("artifacthub-org", "https://artifacthub.io/api/v1/orgs/%s"),
    spec_with(
        "cran",
        "https://cran.r-project.org/web/packages/%s/index.html",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "ctan",
        "https://ctan.org/pkg/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("factorio-mods", "https://mods.factorio.com/api/mods/%s"),
    spec("grafana-plugins", "https://grafana.com/api/plugins/%s"),
    spec("homebrew-cask", "https://formulae.brew.sh/api/cask/%s.json"),
    spec_with(
        "jenkins-plugins",
        "https://plugins.jenkins.io/%s/",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "luarocks-user",
        "https://luarocks.org/modules/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("modrinth", "https://api.modrinth.com/v2/project/%s"),
    spec_with(
        "opam",
        "https://opam.ocaml.org/packages/%s/",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "packagecloud",
        "https://packagecloud.io/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "packagecontrol",
        "https://packagecontrol.io/packages/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "pecl",
        "https://pecl.php.net/package/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("puppet-forge", "https://forgeapi.puppet.com/v3/users/%s"),
    spec("quay-org", "https://quay.io/api/v1/organization/%s"),
    spec("scoop", "https://raw.githubusercontent.com/ScoopInstaller/Main/master/bucket/%s.json"),
    spec("zenodo-communities", "https://zenodo.org/api/communities/%s"),
    spec_with(
        "zora",
        "https://zora.co/@%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "bear-blog",
        "https://%s.bearblog.dev",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "bigcartel",
        "https://%s.bigcartel.com",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "bookwyrm",
        "https://bookwyrm.social/user/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "brunch",
        "https://brunch.co.kr/@%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "dnd-beyond",
        "https://www.dndbeyond.com/members/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "fiverr",
        "https://www.fiverr.com/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "freelancer",
        "https://www.freelancer.com/u/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "hatena-profile",
        "https://profile.hatena.ne.jp/%s/",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "inkitt",
        "https://www.inkitt.com/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "ko-fi",
        "https://ko-fi.com/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "kompasiana",
        "https://www.kompasiana.com/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "qiita",
        "https://qiita.com/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "scratch",
        "https://scratch.mit.edu/users/%s/",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "tabnews",
        "https://www.tabnews.com.br/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "tapas",
        "https://tapas.io/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "cashapp",
        "https://cash.app/$%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("geeksforgeeks", "https://authapi.geeksforgeeks.org/api-get/user-profile-info/?handle=%s"),
    spec("hackadvisor", "https://hackadvisor.io/api/v2/profile/%s/"),
    spec_with(
        "imood",
        "https://www.imood.com/users/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "tistory",
        "https://%s.tistory.com/",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("baserow-community", "https://community.baserow.io/u/%s.json"),
    spec("bitwarden-community", "https://community.bitwarden.com/u/%s.json"),
    spec("e-foundation-community", "https://community.e.foundation/u/%s.json"),
    spec("home-assistant-community", "https://community.home-assistant.io/u/%s.json"),
    spec("community-influxdata-com", "https://community.influxdata.com/u/%s.json"),
    spec("lets-encrypt-community", "https://community.letsencrypt.org/u/%s.json"),
    spec("nocodb-community", "https://community.nocodb.com/u/%s.json"),
    spec("octoprint-community", "https://community.octoprint.org/u/%s.json"),
    spec("openai-community", "https://community.openai.com/u/%s.json"),
    spec("openhab-community", "https://community.openhab.org/u/%s.json"),
    spec("traefik-community", "https://community.traefik.io/u/%s.json"),
    spec("vercel-community", "https://community.vercel.com/u/%s.json"),
    spec("kotlin-discussions", "https://discuss.kotlinlang.org/u/%s.json"),
    spec("kubernetes-discuss", "https://discuss.kubernetes.io/u/%s.json"),
    spec("logseq-forum", "https://discuss.logseq.com/u/%s.json"),
    spec("pytorch-discuss", "https://discuss.pytorch.org/u/%s.json"),
    spec("ruby-on-rails-discussions", "https://discuss.rubyonrails.org/u/%s.json"),
    spec("celo-forum", "https://forum.celo.org/u/%s.json"),
    spec("cosmos-forum", "https://forum.cosmos.network/u/%s.json"),
    spec("crystal-forum", "https://forum.crystal-lang.org/u/%s.json"),
    spec("cursor-forum", "https://forum.cursor.com/u/%s.json"),
    spec("elixir-forum", "https://forum.elixirforum.com/u/%s.json"),
    spec("endeavouros-forum", "https://forum.endeavouros.com/u/%s.json"),
    spec("garuda-linux-forum", "https://forum.garudalinux.org/u/%s.json"),
    spec("ghost-forum", "https://forum.ghost.org/u/%s.json"),
    spec("mattermost-community", "https://forum.mattermost.com/u/%s.json"),
    spec("obsidian-forum", "https://forum.obsidian.md/u/%s.json"),
    spec("openzeppelin-forum", "https://forum.openzeppelin.com/u/%s.json"),
    spec("polkadot-forum", "https://forum.polkadot.network/u/%s.json"),
    spec("rasa-community", "https://forum.rasa.com/u/%s.json"),
    spec("solana-forum", "https://forum.solana.com/u/%s.json"),
    spec("rust-internals-lang", "https://internals.rust-lang.org/u/%s.json"),
    spec("rust-users", "https://users.rust-lang.org/u/%s.json"),
    spec_with(
        "ameblo",
        "https://ameblo.jp/%s/",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "ghost",
        "https://%s.ghost.io/",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "hatena-blog",
        "https://%s.hatenablog.com/",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "livedoor-blog",
        "https://%s.livedoor.blog/",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "softr",
        "https://%s.softr.app/",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    // A claimed blog with a custom domain redirects; a free one 404s.
    spec_with("tumblr", "https://%s.tumblr.com/", DEFAULT_CODES, HTML_ACCEPT),
    spec("blender-video", "https://video.blender.org/api/v1/accounts/%s"),
    spec("lemmy-ml", "https://lemmy.ml/api/v3/user?username=%s"),
    spec("tilvids", "https://tilvids.com/api/v1/accounts/%s"),
    spec("ton-dns", "https://tonapi.io/v2/dns/%s.ton"),
    spec_with(
        "alire",
        "https://alire.ada.dev/crates/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "bioconductor",
        "https://bioconductor.org/packages/release/bioc/html/%s.html",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "conan-center",
        "https://conan.io/center/recipes/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "fedora-rpms",
        "https://src.fedoraproject.org/rpms/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "gnu-elpa",
        "https://elpa.gnu.org/packages/%s.html",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "haxe-lib",
        "https://lib.haxe.org/p/%s/",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "hugo-themes",
        "https://themes.gohugo.io/themes/%s/",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "keybase",
        "https://keybase.io/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("macports", "https://ports.macports.org/api/v1/ports/%s/"),
    spec_with(
        "node-red",
        "https://flows.nodered.org/node/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "nongnu-elpa",
        "https://elpa.nongnu.org/nongnu/%s.html",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "shopify-apps",
        "https://apps.shopify.com/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "stackage",
        "https://www.stackage.org/package/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "zapier-apps",
        "https://zapier.com/apps/%s/integrations",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("apollo-graphql-community", "https://community.apollographql.com/u/%s.json"),
    spec("plotly-community", "https://community.plotly.com/u/%s.json"),
    spec("julia-discourse", "https://discourse.julialang.org/u/%s.json"),
    spec("audacity-forum", "https://forum.audacityteam.org/u/%s.json"),
    spec("django-forum", "https://forum.djangoproject.com/u/%s.json"),
    spec("godot-forum", "https://forum.godotengine.org/u/%s.json"),
    spec("golang-bridge-forum", "https://forum.golangbridge.org/u/%s.json"),
    spec("qiime-2-forum", "https://forum.qiime2.org/u/%s.json"),
    spec("shotcut-forum", "https://forum.shotcut.org/u/%s.json"),
    spec("zorin-forum", "https://forum.zorin.com/u/%s.json"),
    spec("fast-ai-forums", "https://forums.fast.ai/u/%s.json"),
    spec("krita-artists", "https://krita-artists.org/u/%s.json"),
    spec_with(
        "cargo-site",
        "https://%s.cargo.site/",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("audius", "https://api.audius.co/v1/users/handle/%s"),
    spec_with(
        "bionluk",
        "https://bionluk.com/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "blockdit",
        "https://www.blockdit.com/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "cnblogs",
        "https://www.cnblogs.com/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "douban-people",
        "https://www.douban.com/people/%s/",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "exblog",
        "https://%s.exblog.jp",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "instamojo",
        "https://www.instamojo.com/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "kakuyomu",
        "https://kakuyomu.jp/users/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "mataroa",
        "https://%s.mataroa.blog",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "pensador",
        "https://www.pensador.com/autor/%s/",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "postype",
        "https://www.postype.com/@%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "ruby-china",
        "https://ruby-china.org/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec_with(
        "teachable",
        "https://%s.teachable.com",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
    spec("perplexity-community", "https://community.perplexity.ai/u/%s.json"),
    spec("pinecone-community", "https://community.pinecone.io/u/%s.json"),
    spec("ultralytics-community", "https://community.ultralytics.com/u/%s.json"),
    spec("defold-forum", "https://forum.defold.com/u/%s.json"),
    spec("gdevelop-forum", "https://forum.gdevelop.io/u/%s.json"),
    spec("langchain-forum", "https://forum.langchain.com/u/%s.json"),
    spec("opencv-forum", "https://forum.opencv.org/u/%s.json"),
    spec("google-ai-developers", "https://discuss.ai.google.dev/u/%s.json"),
];

/// StatusProvider checks a source by requesting one URL and mapping its status.
pub struct StatusProvider {
    source: Source,
    spec: StatusSpec,
    transport: Arc<dyn Transport>,
}

impl StatusProvider {
    pub fn new(source: Source, spec: StatusSpec, transport: Arc<dyn Transport>) -> Self {
        Self {
            source,
            spec,
            transport,
        }
    }

    fn request_url(&self, label: &str) -> String {
        render(self.spec.request, label)
    }

    fn classify(&self, response: &HttpResponse) -> Outcome {
        let status = response.status;
        if self.spec.codes.taken.contains(&status) {
            if is_challenge(response) {
                return Outcome::new(Status::Unknown).with_detail("blocked");
            }
            return Outcome::new(Status::Taken);
        }
        if self.spec.codes.available.contains(&status) {
            return Outcome::new(Status::Available);
        }
        // Say why a source has no verdict, so the UI can tell a transient
        // rate limit apart from a wall or a source that changed shape.
        let detail = match status {
            429 => "rate limited",
            401 | 403 | 999 => "blocked",
            500..=599 => "server error",
            _ => "unexpected response",
        };
        Outcome::new(Status::Unknown).with_detail(detail)
    }
}

impl Provider for StatusProvider {
    fn id(&self) -> &str {
        &self.source.id
    }

    fn platform(&self) -> &str {
        &self.source.platform
    }

    fn platform_name(&self) -> String {
        self.source.platform_name.clone()
    }

    fn name(&self, label: &str) -> String {
        render(&self.source.name_template, label)
    }

    fn url(&self, label: &str) -> String {
        render(&self.source.url_template, label)
    }

    fn paid(&self) -> bool {
        self.source.paid
    }

    fn normalize_label(&self, label: &str) -> String {
        if self.source.case_sensitive {
            label.to_string()
        } else {
            label.to_ascii_lowercase()
        }
    }

    fn chars(&self) -> crate::label::Chars {
        self.source.chars
    }

    fn min_length(&self) -> usize {
        self.source.min_length
    }

    fn max_length(&self) -> Option<usize> {
        (self.source.max_length > 0).then_some(self.source.max_length)
    }

    fn check(&self, label: &str) -> Result<Outcome, Error> {
        let mut url = self.request_url(label);
        let mut response = None;
        for _ in 0..=10 {
            let mut headers = BTreeMap::new();
            headers.insert("Accept".to_string(), self.spec.accept.to_string());
            headers.insert("Accept-Language".to_string(), ACCEPT_LANGUAGE.to_string());
            headers.insert("User-Agent".to_string(), USER_AGENT.to_string());
            let got = self.transport.execute(HttpRequest {
                method: "GET".to_string(),
                url: url.clone(),
                headers,
                body: None,
            })?;
            if let (true, Some(location)) = (is_redirect(got.status), got.headers.get("location")) {
                url = join_url(&url, location)?;
                response = Some(got);
                continue;
            }
            response = Some(got);
            break;
        }
        let Some(response) = response else {
            return Err(Error::Transport("request produced no response".to_string()));
        };
        Ok(self.classify(&response))
    }
}

/// build_status_provider returns the provider for id when the manifest and a
/// status spec both know it.
pub fn build_status_provider(
    manifest: &Manifest,
    id: &str,
    transport: Arc<dyn Transport>,
) -> Result<StatusProvider, Error> {
    let spec = status_specs()
        .iter()
        .find(|spec| spec.id == id)
        .copied()
        .ok_or_else(|| Error::NotFound(format!("status spec {id}")))?;
    let source = manifest
        .source(id)
        .ok_or_else(|| Error::NotFound(format!("manifest source {id}")))?
        .clone();
    Ok(StatusProvider::new(source, spec, transport))
}

fn render(template: &str, label: &str) -> String {
    template.replace("%s", label)
}

fn is_redirect(status: u16) -> bool {
    matches!(status, 301 | 302 | 303 | 307 | 308)
}

fn join_url(base: &str, location: &str) -> Result<String, Error> {
    let base = url::Url::parse(base)?;
    Ok(base.join(location)?.to_string())
}

fn is_challenge(response: &HttpResponse) -> bool {
    if response
        .headers
        .get("cf-mitigated")
        .is_some_and(|value| value.eq_ignore_ascii_case("challenge"))
    {
        return true;
    }
    let body = String::from_utf8_lossy(&response.body).to_ascii_lowercase();
    body.contains("client challenge") || body.contains("just a moment")
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::fixture::{Fixture, ReplayTransport};

    fn fixture(source: &str, url: &str, status: u16) -> Fixture {
        serde_json::from_value(serde_json::json!({
            "source": source,
            "platform": source,
            "label": "acema",
            "name": "acema",
            "url": "https://example.test/acema",
            "requests": [{
                "method": "GET",
                "url": url,
                "response": { "status": status }
            }]
        }))
        .expect("fixture")
    }

    fn provider(id: &str, fixture: Fixture) -> StatusProvider {
        let manifest = Manifest::embedded().expect("manifest");
        build_status_provider(&manifest, id, Arc::new(ReplayTransport::new(&fixture)))
            .expect("provider")
    }

    #[test]
    fn unknown_verdicts_carry_a_reason() {
        for (code, detail) in [
            (429, "rate limited"),
            (403, "blocked"),
            (500, "server error"),
            (418, "unexpected response"),
        ] {
            let provider = provider(
                "npm",
                fixture("npm", "https://registry.npmjs.org/acema", code),
            );
            let outcome = provider.check("acema").expect("check");
            assert_eq!(outcome.status, Status::Unknown, "{code}");
            assert_eq!(outcome.detail, detail, "{code}");
        }
    }

    #[test]
    fn sso_protected_deployment_is_taken() {
        // Vercel answers 401 for a JSON request to a deployment behind SSO
        // (302 to /sso-api for HTML); free names 404.
        let provider = provider(
            "vercel",
            fixture("vercel", "https://acema.vercel.app/", 401),
        );
        let outcome = provider.check("acema").expect("check");
        assert_eq!(outcome.status, Status::Taken);
    }

    #[test]
    fn case_sensitive_sources_keep_the_raw_label() {
        // CPAN modules are case-sensitive: DBI exists, dbi does not.
        let provider = provider(
            "cpan",
            fixture("cpan", "https://fastapi.metacpan.org/v1/module/DBI", 200),
        );
        assert_eq!(provider.normalize_label("DBI"), "DBI");
        let outcome = provider
            .check(&provider.normalize_label("DBI"))
            .expect("check");
        assert_eq!(outcome.status, Status::Taken);
    }

    #[test]
    fn case_insensitive_sources_fold_the_label() {
        let provider = provider(
            "npm",
            fixture("npm", "https://registry.npmjs.org/acema", 404),
        );
        assert_eq!(provider.normalize_label("AceMa"), "acema");
        let outcome = provider
            .check(&provider.normalize_label("AceMa"))
            .expect("check");
        assert_eq!(outcome.status, Status::Available);
    }
}
