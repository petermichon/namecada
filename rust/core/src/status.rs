//! Status-only providers: the largest adapter family. A request URL plus a
//! status code table covers every source whose verdict is just the HTTP status.

use std::collections::BTreeMap;
use std::sync::Arc;

use crate::manifest::{Manifest, Source};
use crate::provider::{Error, HttpRequest, HttpResponse, Outcome, Provider, Status, Transport};

pub const DEFAULT_ACCEPT: &str = "application/json";
pub const HTML_ACCEPT: &str = "text/html";
pub const USER_AGENT: &str = "namecheck/1.0 (+https://github.com/petermichon/namecheck)";

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
        "goodreads",
        "https://www.goodreads.com/%s",
        DEFAULT_CODES,
        HTML_ACCEPT,
    ),
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
    spec("spotify", "https://open.spotify.com/user/%s"),
    spec("paragraph", "https://paragraph.com/@%s"),
    spec("linktree", "https://linktr.ee/%s"),
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
    spec("vercel", "https://%s.vercel.app/"),
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

    fn classify(&self, response: &HttpResponse) -> Status {
        let status = response.status;
        if self.spec.codes.taken.contains(&status) {
            if is_challenge(response) {
                return Status::Unknown;
            }
            return Status::Taken;
        }
        if self.spec.codes.available.contains(&status) {
            return Status::Available;
        }
        Status::Unknown
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
        Ok(Outcome::new(self.classify(&response)))
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
