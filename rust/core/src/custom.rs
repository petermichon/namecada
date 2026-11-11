//! Body-parsing and multi-request providers. Each check mirrors the source
//! semantics: transport failures propagate; malformed or unexpected bodies
//! degrade to `unknown`.

use std::collections::BTreeMap;
use std::sync::Arc;

use serde_json::{json, Value};

use crate::manifest::{Manifest, Source};
use crate::provider::{Error, HttpRequest, HttpResponse, Outcome, Provider, Status, Transport};
use crate::status::{DEFAULT_ACCEPT, USER_AGENT};

const HTML_ANY_ACCEPT: &str = "text/html,*/*";

fn unknown() -> Outcome {
    Outcome::new(Status::Unknown)
}

fn available() -> Outcome {
    Outcome::new(Status::Available)
}

fn taken() -> Outcome {
    Outcome::new(Status::Taken)
}

fn get(transport: &dyn Transport, url: &str, accept: &str) -> Result<HttpResponse, Error> {
    Ok(get_follow(transport, url, accept)?.response)
}

fn get_once(transport: &dyn Transport, url: &str, accept: &str) -> Result<HttpResponse, Error> {
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

fn post_json(transport: &dyn Transport, url: &str, body: &Value) -> Result<HttpResponse, Error> {
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

fn json_body(response: &HttpResponse) -> Option<Value> {
    serde_json::from_slice(&response.body).ok()
}

fn text_body(response: &HttpResponse) -> String {
    String::from_utf8_lossy(&response.body).into_owned()
}

fn check_status(response: &HttpResponse, taken: Status) -> Option<Status> {
    if response.status == 200 {
        Some(taken)
    } else if response.status == 404 {
        Some(Status::Available)
    } else {
        None
    }
}

fn check_ens(_: &Source, transport: &dyn Transport, label: &str) -> Result<Outcome, Error> {
    let url = format!("https://api.ensideas.com/ens/resolve/{label}.eth");
    let response = get(transport, &url, DEFAULT_ACCEPT)?;
    if response.status != 200 {
        return Ok(unknown());
    }
    let Some(value) = json_body(&response) else {
        return Ok(unknown());
    };
    match value.get("address") {
        None | Some(Value::Null) => Ok(available()),
        Some(Value::String(address)) if address.is_empty() => Ok(available()),
        Some(Value::String(_)) => Ok(taken()),
        _ => Ok(unknown()),
    }
}

fn check_sns(_: &Source, transport: &dyn Transport, label: &str) -> Result<Outcome, Error> {
    let url = format!("https://sdk-proxy.sns.id/resolve/{label}.sol");
    let response = get(transport, &url, DEFAULT_ACCEPT)?;
    if response.status != 200 {
        return Ok(unknown());
    }
    let Some(value) = json_body(&response) else {
        return Ok(unknown());
    };
    match value.get("s").and_then(Value::as_str) {
        Some("ok") => {
            if value
                .get("result")
                .and_then(Value::as_str)
                .is_some_and(|result| !result.is_empty())
            {
                Ok(taken())
            } else {
                Ok(unknown())
            }
        }
        Some("error") => Ok(available()),
        _ => Ok(unknown()),
    }
}

fn check_hns(_: &Source, transport: &dyn Transport, label: &str) -> Result<Outcome, Error> {
    let url = format!("https://hnsscan.com/api/names/{label}/details");
    let response = get(transport, &url, DEFAULT_ACCEPT)?;
    if response.status != 200 {
        return Ok(unknown());
    }
    let Some(value) = json_body(&response) else {
        return Ok(unknown());
    };
    match value.pointer("/info/state").and_then(Value::as_str) {
        Some("CLOSED") => Ok(taken()),
        Some("AVAILABLE" | "UNKNOWN") => Ok(available()),
        _ => Ok(unknown()),
    }
}

fn check_namecoin(_: &Source, transport: &dyn Transport, label: &str) -> Result<Outcome, Error> {
    let url = format!("https://explorer.namecoin.co/api/name/d%2F{label}");
    let response = get(transport, &url, DEFAULT_ACCEPT)?;
    if response.status == 404 {
        return Ok(available());
    }
    if response.status != 200 {
        return Ok(unknown());
    }
    let Some(value) = json_body(&response) else {
        return Ok(unknown());
    };
    match value.get("expired").and_then(Value::as_bool) {
        Some(true) => Ok(available()),
        Some(false) => Ok(taken()),
        None => Ok(unknown()),
    }
}

fn check_lens(_: &Source, transport: &dyn Transport, label: &str) -> Result<Outcome, Error> {
    let body = json!({
        "query": "query ($name: String) { account(request: {username: {localName: $name}}) { username { localName value } } }",
        "variables": { "name": label },
    });
    let response = post_json(transport, "https://api.lens.xyz/graphql", &body)?;
    let Some(value) = json_body(&response) else {
        return Ok(unknown());
    };
    match value.pointer("/data/account") {
        None => Ok(unknown()),
        Some(Value::Null) => Ok(available()),
        Some(_) => Ok(taken()),
    }
}

fn check_anilist(_: &Source, transport: &dyn Transport, label: &str) -> Result<Outcome, Error> {
    let body = json!({
        "query": "query ($name: String) { User(name: $name) { id } }",
        "variables": { "name": label },
    });
    let response = post_json(transport, "https://graphql.anilist.co", &body)?;
    let Some(value) = json_body(&response) else {
        return Ok(unknown());
    };
    match value.pointer("/data/User") {
        None => Ok(unknown()),
        Some(Value::Null) => Ok(available()),
        Some(_) => Ok(taken()),
    }
}

fn check_roblox(_: &Source, transport: &dyn Transport, label: &str) -> Result<Outcome, Error> {
    let body = json!({
        "usernames": [label],
        "excludeBannedUsers": false,
    });
    let response = post_json(
        transport,
        "https://users.roblox.com/v1/usernames/users",
        &body,
    )?;
    if response.status != 200 {
        return Ok(unknown());
    }
    let Some(value) = json_body(&response) else {
        return Ok(unknown());
    };
    match value.get("data").and_then(Value::as_array) {
        None => Ok(unknown()),
        Some(entries) if entries.is_empty() => Ok(available()),
        Some(_) => Ok(taken()),
    }
}

fn check_codeberg(_: &Source, transport: &dyn Transport, label: &str) -> Result<Outcome, Error> {
    let url = format!("https://codeberg.org/api/v1/users/{label}");
    let response = get(transport, &url, DEFAULT_ACCEPT)?;
    match response.status {
        200 => Ok(taken()),
        404 => {
            let message = json_body(&response).and_then(|value| {
                value
                    .get("message")
                    .and_then(Value::as_str)
                    .map(str::to_string)
            });
            match message {
                None => Ok(unknown()),
                Some(message) if message.contains("does not exist") => Ok(available()),
                Some(_) => Ok(taken()),
            }
        }
        _ => Ok(unknown()),
    }
}

fn check_wikipedia(_: &Source, transport: &dyn Transport, label: &str) -> Result<Outcome, Error> {
    let url = format!(
        "https://meta.wikimedia.org/w/api.php?action=query&format=json&guiuser={label}&meta=globaluserinfo&origin=%2A"
    );
    let response = get(transport, &url, DEFAULT_ACCEPT)?;
    if response.status != 200 {
        return Ok(unknown());
    }
    let Some(value) = json_body(&response) else {
        return Ok(unknown());
    };
    match value.pointer("/query/globaluserinfo") {
        Some(Value::Object(info)) => {
            if info.contains_key("missing") {
                Ok(available())
            } else {
                Ok(taken())
            }
        }
        _ => Ok(unknown()),
    }
}

fn check_devto(_: &Source, transport: &dyn Transport, label: &str) -> Result<Outcome, Error> {
    let user_url = format!("https://dev.to/api/users/by_username?url={label}");
    let user = get(transport, &user_url, DEFAULT_ACCEPT)?;
    match user.status {
        200 => return Ok(taken()),
        404 => {}
        _ => return Ok(unknown()),
    }

    let org_url = format!("https://dev.to/api/organizations/{label}");
    let org = get(transport, &org_url, DEFAULT_ACCEPT)?;
    Ok(check_status(&org, Status::Taken).map_or_else(unknown, Outcome::new))
}

fn check_minehut(_: &Source, transport: &dyn Transport, label: &str) -> Result<Outcome, Error> {
    let url = format!("https://api.minehut.com/server/{label}?byName=true");
    let response = get(transport, &url, DEFAULT_ACCEPT)?;
    if response.status != 200 {
        return Ok(unknown());
    }
    let Some(value) = json_body(&response) else {
        return Ok(unknown());
    };
    match value.get("server") {
        None => Ok(unknown()),
        Some(Value::Null) => Ok(available()),
        Some(_) => Ok(taken()),
    }
}

fn check_bluesky(_: &Source, transport: &dyn Transport, label: &str) -> Result<Outcome, Error> {
    let url = format!(
        "https://bsky.social/xrpc/com.atproto.temp.checkHandleAvailability?handle={label}.bsky.social"
    );
    let response = get(transport, &url, DEFAULT_ACCEPT)?;
    if response.status != 200 {
        return Ok(unknown());
    }
    let Some(value) = json_body(&response) else {
        return Ok(unknown());
    };
    match value.pointer("/result/$type").and_then(Value::as_str) {
        None | Some("") => Ok(unknown()),
        Some(kind) if kind.ends_with("#resultAvailable") => Ok(available()),
        Some(_) => Ok(taken()),
    }
}

fn check_opencollective(
    _: &Source,
    transport: &dyn Transport,
    label: &str,
) -> Result<Outcome, Error> {
    let url = format!("https://opencollective.com/{label}.json");
    let response = get(transport, &url, DEFAULT_ACCEPT)?;
    Ok(check_status(&response, Status::Taken).map_or_else(unknown, Outcome::new))
}

fn check_instagram(_: &Source, transport: &dyn Transport, label: &str) -> Result<Outcome, Error> {
    let url = format!("https://www.instagram.com/{label}/");
    let response = get(transport, &url, DEFAULT_ACCEPT)?;
    if response.status != 200 {
        return Ok(unknown());
    }
    let text = text_body(&response);
    if text.contains(&format!("\"username\":\"{label}\"")) {
        return Ok(taken());
    }
    if let Some(index) = text.find("<title>") {
        if text[index + "<title>".len()..].starts_with(label) {
            return Ok(taken());
        }
    }
    if text.contains("<title>Instagram</title>") {
        return Ok(available());
    }
    Ok(unknown())
}

fn check_dribbble(_: &Source, transport: &dyn Transport, label: &str) -> Result<Outcome, Error> {
    let url = format!("https://dribbble.com/{label}");
    let response = get(transport, &url, HTML_ANY_ACCEPT)?;
    match response.status {
        200 => Ok(taken()),
        404 => {
            if text_body(&response).contains("public yet") {
                Ok(taken())
            } else {
                Ok(available())
            }
        }
        _ => Ok(unknown()),
    }
}

fn check_hackernews(_: &Source, transport: &dyn Transport, label: &str) -> Result<Outcome, Error> {
    let url = format!("https://news.ycombinator.com/user?id={label}");
    let response = get(transport, &url, DEFAULT_ACCEPT)?;
    if response.status != 200 {
        return Ok(unknown());
    }
    let text = text_body(&response);
    if text.contains("No such user.") {
        return Ok(available());
    }
    if text.contains("karma") {
        return Ok(taken());
    }
    Ok(unknown())
}

fn check_giphy(_: &Source, transport: &dyn Transport, label: &str) -> Result<Outcome, Error> {
    let url = format!("https://giphy.com/{label}");
    let fetched = get_follow(transport, &url, HTML_ANY_ACCEPT)?;
    if fetched.response.status != 200 {
        return Ok(unknown());
    }
    let path = url::Url::parse(&fetched.url)
        .map(|url| url.path().to_string())
        .unwrap_or_default();
    if path.starts_with("/explore/") {
        return Ok(available());
    }
    if path == format!("/{label}") {
        return Ok(taken());
    }
    if text_body(&fetched.response)
        .to_ascii_lowercase()
        .contains("gifs on giphy")
    {
        return Ok(taken());
    }
    Ok(unknown())
}

fn check_minds(_: &Source, transport: &dyn Transport, label: &str) -> Result<Outcome, Error> {
    let url = format!("https://www.minds.com/api/v1/channel/{label}");
    let response = get(transport, &url, DEFAULT_ACCEPT)?;
    if response.status != 200 {
        return Ok(unknown());
    }
    let Some(value) = json_body(&response) else {
        return Ok(unknown());
    };
    let status = value.get("status").and_then(Value::as_str);
    let kind = value.get("type").and_then(Value::as_str);
    let channel = value.get("channel");
    match status {
        Some("success") if channel.is_some_and(|value| !value.is_null()) => Ok(taken()),
        Some("error") if kind == Some("ChannelNotFoundException") => Ok(available()),
        _ => Ok(unknown()),
    }
}

/// The Snap Store answers 404 for an unclaimed name and 200 for a registered
/// one, but only when the request names the device series, so this check sends
/// its own headers instead of using `get`.
fn check_snap(_: &Source, transport: &dyn Transport, label: &str) -> Result<Outcome, Error> {
    let mut headers = BTreeMap::new();
    headers.insert("Accept".to_string(), DEFAULT_ACCEPT.to_string());
    headers.insert("Snap-Device-Series".to_string(), "16".to_string());
    headers.insert("User-Agent".to_string(), USER_AGENT.to_string());
    let response = transport.execute(HttpRequest {
        method: "GET".to_string(),
        url: format!("https://api.snapcraft.io/v2/snaps/info/{label}"),
        headers,
        body: None,
    })?;
    Ok(check_status(&response, Status::Taken).map_or_else(unknown, Outcome::new))
}

fn check_steam(_: &Source, transport: &dyn Transport, label: &str) -> Result<Outcome, Error> {
    let url = format!("https://steamcommunity.com/id/{label}?xml=1");
    let response = get(transport, &url, DEFAULT_ACCEPT)?;
    if response.status != 200 {
        return Ok(unknown());
    }
    let text = text_body(&response);
    if text.contains("<steamID64>") {
        return Ok(taken());
    }
    if text.contains("<error>") {
        return Ok(available());
    }
    Ok(unknown())
}

fn check_wordpress(_: &Source, transport: &dyn Transport, label: &str) -> Result<Outcome, Error> {
    let url = format!(
        "https://api.wordpress.org/plugins/info/1.2/?action=plugin_information&request%5Bslug%5D={label}"
    );
    let response = get(transport, &url, DEFAULT_ACCEPT)?;
    match response.status {
        200 => Ok(taken()),
        404 => {
            let closed = json_body(&response)
                .and_then(|value| value.get("closed").and_then(Value::as_bool))
                .unwrap_or(false);
            if closed {
                Ok(taken())
            } else {
                Ok(available())
            }
        }
        _ => Ok(unknown()),
    }
}

fn check_gitlab(_: &Source, transport: &dyn Transport, label: &str) -> Result<Outcome, Error> {
    let users_url = format!("https://gitlab.com/api/v4/users?username={label}");
    let users = get(transport, &users_url, DEFAULT_ACCEPT)?;
    if users.status != 200 {
        return Ok(unknown());
    }
    let Some(value) = json_body(&users) else {
        return Ok(unknown());
    };
    match value.as_array() {
        None => Ok(unknown()),
        Some(users) if !users.is_empty() => Ok(taken()),
        Some(_) => {
            let group_url = format!("https://gitlab.com/api/v4/groups/{label}");
            let group = get(transport, &group_url, DEFAULT_ACCEPT)?;
            Ok(check_status(&group, Status::Taken).map_or_else(unknown, Outcome::new))
        }
    }
}

fn check_peertube(
    source: &Source,
    transport: &dyn Transport,
    label: &str,
) -> Result<Outcome, Error> {
    let account_url = format!("https://framatube.org/api/v1/accounts/{label}");
    let account = get(transport, &account_url, DEFAULT_ACCEPT)?;
    if account.status == 200 {
        let mut outcome = taken();
        outcome.url = Some(render(&source.url_template, label));
        return Ok(outcome);
    }

    let channel_url = format!("https://framatube.org/api/v1/video-channels/{label}");
    let channel = get(transport, &channel_url, DEFAULT_ACCEPT)?;
    if channel.status == 200 {
        let mut outcome = taken();
        outcome.url = Some(render(&source.url_template.replace("/a/", "/c/"), label));
        return Ok(outcome);
    }

    if account.status == 404 && channel.status == 404 {
        return Ok(available());
    }
    Ok(unknown())
}

const RESERVED_LABELS: &[&str] = &["example", "test", "invalid", "localhost"];

fn check_domain(source: &Source, transport: &dyn Transport, label: &str) -> Result<Outcome, Error> {
    if RESERVED_LABELS.contains(&label) {
        return Ok(taken());
    }
    let tld = source
        .id
        .strip_prefix("domain.")
        .ok_or_else(|| Error::Invalid(format!("domain source has id {}", source.id)))?;
    let base = rdap_base(tld).ok_or_else(|| Error::Invalid(format!("no RDAP base for .{tld}")))?;
    let url = format!("{base}domain/{label}.{tld}");
    let response = get(transport, &url, DEFAULT_ACCEPT)?;
    match response.status {
        200 => Ok(taken()),
        404 => Ok(available()),
        _ => Ok(unknown()),
    }
}

fn rdap_base(tld: &str) -> Option<&'static str> {
    match tld {
        "com" => Some("https://rdap.verisign.com/com/v1/"),
        "net" => Some("https://rdap.verisign.com/net/v1/"),
        "org" => Some("https://rdap.publicinterestregistry.org/rdap/"),
        "dev" | "app" => Some("https://pubapi.registry.google/rdap/"),
        "io" | "ai" => Some("https://rdap.identitydigital.services/rdap/"),
        _ => None,
    }
}

struct Fetched {
    response: HttpResponse,
    url: String,
}

fn get_follow(transport: &dyn Transport, url: &str, accept: &str) -> Result<Fetched, Error> {
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

fn join_url(base: &str, location: &str) -> Result<String, Error> {
    let base = url::Url::parse(base)?;
    Ok(base.join(location)?.to_string())
}

fn render(template: &str, label: &str) -> String {
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

static CUSTOM_SPECS: &[CustomSpec] = &[
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
];

/// CustomProvider wraps a custom check function with manifest metadata.
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
}
