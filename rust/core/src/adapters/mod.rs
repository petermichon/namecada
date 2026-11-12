//! Body-parsing, multi-request, POST/GraphQL and RDAP providers.

use std::collections::BTreeMap;
use std::sync::Arc;

use serde_json::Value;

use crate::manifest::{Manifest, Source};
use crate::provider::{Error, HttpRequest, HttpResponse, Outcome, Provider, Status, Transport};
use status::{DEFAULT_ACCEPT, USER_AGENT};

mod api;
pub mod dns;
mod domains;
mod page;
pub mod status;

use api::*;
use domains::*;
use page::*;

pub(crate) fn unknown() -> Outcome {
    // A 200 whose body did not match the expected shape.
    Outcome::new(Status::Unknown).with_detail("unexpected response")
}

/// unknown_response explains a non-200 that left a source without a verdict.
pub(crate) fn unknown_response(response: &HttpResponse) -> Outcome {
    let detail = match response.status {
        429 => "rate limited",
        401 | 403 | 999 => "blocked",
        500..=599 => "server error",
        _ => "unexpected response",
    };
    Outcome::new(Status::Unknown).with_detail(detail)
}

pub(crate) fn available() -> Outcome {
    Outcome::new(Status::Available)
}

pub(crate) fn taken() -> Outcome {
    Outcome::new(Status::Taken)
}

pub(crate) fn get(
    transport: &dyn Transport,
    url: &str,
    accept: &str,
) -> Result<HttpResponse, Error> {
    Ok(get_follow(transport, url, accept)?.response)
}

pub(crate) fn get_once(
    transport: &dyn Transport,
    url: &str,
    accept: &str,
) -> Result<HttpResponse, Error> {
    let mut headers = BTreeMap::new();
    headers.insert("Accept".to_string(), accept.to_string());
    headers.insert("User-Agent".to_string(), USER_AGENT.to_string());
    transport.execute(HttpRequest {
        method: "GET".to_string(),
        url: url.to_string(),
        headers,
        body: None,
    })
}

pub(crate) fn post_json(
    transport: &dyn Transport,
    url: &str,
    body: &Value,
) -> Result<HttpResponse, Error> {
    let mut headers = BTreeMap::new();
    headers.insert("Accept".to_string(), DEFAULT_ACCEPT.to_string());
    headers.insert("Content-Type".to_string(), "application/json".to_string());
    headers.insert("User-Agent".to_string(), USER_AGENT.to_string());
    transport.execute(HttpRequest {
        method: "POST".to_string(),
        url: url.to_string(),
        headers,
        body: Some(serde_json::to_vec(body)?),
    })
}

pub(crate) fn json_body(response: &HttpResponse) -> Option<Value> {
    serde_json::from_slice(&response.body).ok()
}

pub(crate) fn text_body(response: &HttpResponse) -> String {
    String::from_utf8_lossy(&response.body).into_owned()
}

pub(crate) fn check_status(response: &HttpResponse, taken: Status) -> Option<Status> {
    if response.status == 200 {
        Some(taken)
    } else if response.status == 404 {
        Some(Status::Available)
    } else {
        None
    }
}

pub(crate) struct Fetched {
    response: HttpResponse,
    url: String,
}

pub(crate) fn get_follow(
    transport: &dyn Transport,
    url: &str,
    accept: &str,
) -> Result<Fetched, Error> {
    let mut current = url.to_string();
    for _ in 0..=10 {
        let response = get_once(transport, &current, accept)?;
        if matches!(response.status, 301 | 302 | 303 | 307 | 308) {
            if let Some(location) = response.headers.get("location") {
                current = join_url(&current, location)?;
                continue;
            }
        }
        return Ok(Fetched {
            response,
            url: current,
        });
    }
    Err(Error::Transport("too many redirects".to_string()))
}

pub(crate) fn join_url(base: &str, location: &str) -> Result<String, Error> {
    let base = url::Url::parse(base)?;
    Ok(base.join(location)?.to_string())
}

pub(crate) fn render(template: &str, label: &str) -> String {
    template.replace("%s", label)
}

/// CheckFn is one custom provider's implementation.
pub type CheckFn = fn(&Source, &dyn Transport, &str) -> Result<Outcome, Error>;

/// CustomSpec ties a manifest id to its check function.
pub struct CustomSpec {
    pub id: &'static str,
    pub check: CheckFn,
}

/// custom_specs returns every provider this adapter currently covers.
pub fn custom_specs() -> &'static [CustomSpec] {
    CUSTOM_SPECS
}

pub(crate) static CUSTOM_SPECS: &[CustomSpec] = &[
    CustomSpec {
        id: "ens",
        check: check_ens,
    },
    CustomSpec {
        id: "sns",
        check: check_sns,
    },
    CustomSpec {
        id: "hns",
        check: check_hns,
    },
    CustomSpec {
        id: "namecoin",
        check: check_namecoin,
    },
    CustomSpec {
        id: "lens",
        check: check_lens,
    },
    CustomSpec {
        id: "anilist",
        check: check_anilist,
    },
    CustomSpec {
        id: "roblox",
        check: check_roblox,
    },
    CustomSpec {
        id: "codeberg",
        check: check_codeberg,
    },
    CustomSpec {
        id: "wikipedia",
        check: check_wikipedia,
    },
    CustomSpec {
        id: "devto",
        check: check_devto,
    },
    CustomSpec {
        id: "minehut",
        check: check_minehut,
    },
    CustomSpec {
        id: "bluesky",
        check: check_bluesky,
    },
    CustomSpec {
        id: "opencollective",
        check: check_opencollective,
    },
    CustomSpec {
        id: "instagram",
        check: check_instagram,
    },
    CustomSpec {
        id: "dribbble",
        check: check_dribbble,
    },
    CustomSpec {
        id: "hackernews",
        check: check_hackernews,
    },
    CustomSpec {
        id: "giphy",
        check: check_giphy,
    },
    CustomSpec {
        id: "minds",
        check: check_minds,
    },
    CustomSpec {
        id: "snap",
        check: check_snap,
    },
    CustomSpec {
        id: "bandcamp",
        check: check_bandcamp,
    },
    CustomSpec {
        id: "steam",
        check: check_steam,
    },
    CustomSpec {
        id: "wordpress",
        check: check_wordpress,
    },
    CustomSpec {
        id: "gitlab",
        check: check_gitlab,
    },
    CustomSpec {
        id: "peertube",
        check: check_peertube,
    },
    CustomSpec {
        id: "domain.com",
        check: check_domain,
    },
    CustomSpec {
        id: "domain.org",
        check: check_domain,
    },
    CustomSpec {
        id: "domain.dev",
        check: check_domain,
    },
    CustomSpec {
        id: "domain.io",
        check: check_domain,
    },
    CustomSpec {
        id: "domain.net",
        check: check_domain,
    },
    CustomSpec {
        id: "domain.ai",
        check: check_domain,
    },
    CustomSpec {
        id: "domain.app",
        check: check_domain,
    },
    CustomSpec {
        id: "tiktok",
        check: check_tiktok,
    },
    CustomSpec {
        id: "threads",
        check: check_threads,
    },
    CustomSpec {
        id: "twitch",
        check: check_twitch,
    },
    CustomSpec {
        id: "duolingo",
        check: check_duolingo,
    },
    CustomSpec {
        id: "telegram",
        check: check_telegram,
    },
];

pub struct CustomProvider {
    source: Source,
    check: CheckFn,
    transport: Arc<dyn Transport>,
}

impl CustomProvider {
    pub fn new(source: Source, check: CheckFn, transport: Arc<dyn Transport>) -> Self {
        Self {
            source,
            check,
            transport,
        }
    }
}

impl Provider for CustomProvider {
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
        (self.check)(&self.source, self.transport.as_ref(), label)
    }
}

/// build_custom_provider returns the provider for id when the manifest and a
/// custom spec both know it.
pub fn build_custom_provider(
    manifest: &Manifest,
    id: &str,
    transport: Arc<dyn Transport>,
) -> Result<CustomProvider, Error> {
    let spec = custom_specs()
        .iter()
        .find(|spec| spec.id == id)
        .ok_or_else(|| Error::NotFound(format!("custom spec {id}")))?;
    let source = manifest
        .source(id)
        .ok_or_else(|| Error::NotFound(format!("manifest source {id}")))?
        .clone();
    Ok(CustomProvider::new(source, spec.check, transport))
}

#[cfg(test)]
mod tests {
    use super::api::{check_gitlab, check_hns, check_snap};
    use super::page::check_threads;
    use super::*;
    use crate::fixture::{Fixture, ReplayTransport};

    /// Both checked-in labels are unclaimed on the Snap Store, so the taken
    /// branch is covered here instead of by a fixture.
    fn snap_outcome(status: u16) -> Outcome {
        let fixture: Fixture = serde_json::from_value(serde_json::json!({
            "source": "snap",
            "platform": "snap",
            "label": "spotify",
            "name": "spotify",
            "url": "https://snapcraft.io/spotify",
            "requests": [{
                "method": "GET",
                "url": "https://api.snapcraft.io/v2/snaps/info/spotify",
                "response": { "status": status }
            }]
        }))
        .expect("fixture");
        let transport = ReplayTransport::new(&fixture);
        let source = Manifest::embedded()
            .expect("manifest")
            .source("snap")
            .expect("snap source")
            .clone();
        check_snap(&source, &transport, "spotify").expect("check")
    }

    #[test]
    fn snap_maps_the_store_status() {
        assert_eq!(snap_outcome(200).status, Status::Taken);
        assert_eq!(snap_outcome(404).status, Status::Available);
        assert_eq!(snap_outcome(500).status, Status::Unknown);
    }

    #[test]
    fn gitlab_says_when_it_is_rate_limited() {
        let fixture: Fixture = serde_json::from_value(serde_json::json!({
            "source": "gitlab",
            "platform": "gitlab",
            "label": "google",
            "name": "google",
            "url": "https://gitlab.com/google",
            "requests": [{
                "method": "GET",
                "url": "https://gitlab.com/api/v4/users?username=google",
                "response": { "status": 429, "body": "Rate limit exceeded" }
            }]
        }))
        .expect("fixture");
        let transport = ReplayTransport::new(&fixture);
        let source = Manifest::embedded()
            .expect("manifest")
            .source("gitlab")
            .expect("gitlab source")
            .clone();
        let outcome = check_gitlab(&source, &transport, "google").expect("check");
        assert_eq!(outcome.status, Status::Unknown);
        assert_eq!(outcome.detail, "rate limited");
    }

    /// A profile title names the handle whether or not the account has a
    /// display name; the missing page is the login shell.
    #[test]
    fn threads_reads_the_handle_out_of_the_title() {
        fn outcome(title: &str) -> Status {
            let fixture: Fixture = serde_json::from_value(serde_json::json!({
                "source": "threads",
                "platform": "threads",
                "label": "hardvoxel",
                "name": "hardvoxel",
                "url": "https://www.threads.net/@hardvoxel",
                "requests": [{
                    "method": "GET",
                    "url": "https://www.threads.net/@hardvoxel",
                    "response": {
                        "status": 200,
                        "body": format!("<html><head><title>{title}</title></head></html>")
                    }
                }]
            }))
            .expect("fixture");
            let transport = ReplayTransport::new(&fixture);
            let source = Manifest::embedded()
                .expect("manifest")
                .source("threads")
                .expect("threads source")
                .clone();
            check_threads(&source, &transport, "hardvoxel")
                .expect("check")
                .status
        }

        assert_eq!(
            outcome("Hardvoxel (&#064;hardvoxel) &#x2022; Threads, Say more"),
            Status::Taken
        );
        // An account with no display name puts the handle first.
        assert_eq!(
            outcome("&#064;hardvoxel &#x2022; Threads, Say more"),
            Status::Taken
        );
        assert_eq!(outcome("Threads &#x2022; Log in"), Status::Available);
        assert_eq!(outcome("Threads"), Status::Unknown);
    }

    /// hnsscan 502s on these, so they are answered without a request. An empty
    /// replay queue would fail loudly if the check tried to reach the network.
    #[test]
    fn hns_reports_special_use_names_as_taken() {
        let fixture: Fixture = serde_json::from_value(serde_json::json!({
            "source": "hns",
            "platform": "hns",
            "label": "example",
            "name": "example",
            "url": "https://hnsscan.com/names/example",
            "requests": []
        }))
        .expect("fixture");
        let transport = ReplayTransport::new(&fixture);
        let source = Manifest::embedded()
            .expect("manifest")
            .source("hns")
            .expect("hns source")
            .clone();

        for label in ["example", "test", "invalid", "localhost", "local"] {
            let outcome = check_hns(&source, &transport, label).expect("check");
            assert_eq!(outcome.status, Status::Taken, "{label}");
        }
    }
}
