//! Sources that expose a structured endpoint: JSON, XML or GraphQL.

use std::collections::BTreeMap;

use serde_json::{json, Value};

use crate::manifest::Source;
use crate::provider::{Error, HttpRequest, Outcome, Status, Transport};

use super::shared::*;

pub(crate) fn check_ens(
    _: &Source,
    transport: &dyn Transport,
    label: &str,
) -> Result<Outcome, Error> {
    let url = format!("https://api.ensideas.com/ens/resolve/{label}.eth");
    let response = get(transport, &url, DEFAULT_ACCEPT)?;
    if response.status != 200 {
        return Ok(unknown_response(&response));
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

pub(crate) fn check_sns(
    _: &Source,
    transport: &dyn Transport,
    label: &str,
) -> Result<Outcome, Error> {
    let url = format!("https://sdk-proxy.sns.id/resolve/{label}.sol");
    let response = get(transport, &url, DEFAULT_ACCEPT)?;
    if response.status != 200 {
        return Ok(unknown_response(&response));
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

pub(crate) fn check_hns(
    _: &Source,
    transport: &dyn Transport,
    label: &str,
) -> Result<Outcome, Error> {
    if HNS_RESERVED.contains(&label) {
        return Ok(taken());
    }
    let url = format!("https://hnsscan.com/api/names/{label}/details");
    let response = get(transport, &url, DEFAULT_ACCEPT)?;
    if response.status != 200 {
        return Ok(unknown_response(&response));
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

pub(crate) fn check_namecoin(
    _: &Source,
    transport: &dyn Transport,
    label: &str,
) -> Result<Outcome, Error> {
    let url = format!("https://explorer.namecoin.co/api/name/d%2F{label}");
    let response = get(transport, &url, DEFAULT_ACCEPT)?;
    if response.status == 404 {
        return Ok(available());
    }
    if response.status != 200 {
        return Ok(unknown_response(&response));
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

pub(crate) fn check_lens(
    _: &Source,
    transport: &dyn Transport,
    label: &str,
) -> Result<Outcome, Error> {
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

pub(crate) fn check_anilist(
    _: &Source,
    transport: &dyn Transport,
    label: &str,
) -> Result<Outcome, Error> {
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

pub(crate) fn check_roblox(
    _: &Source,
    transport: &dyn Transport,
    label: &str,
) -> Result<Outcome, Error> {
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
        return Ok(unknown_response(&response));
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

pub(crate) fn check_codeberg(
    _: &Source,
    transport: &dyn Transport,
    label: &str,
) -> Result<Outcome, Error> {
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

pub(crate) fn check_wikipedia(
    _: &Source,
    transport: &dyn Transport,
    label: &str,
) -> Result<Outcome, Error> {
    let url = format!(
        "https://meta.wikimedia.org/w/api.php?action=query&format=json&guiuser={label}&meta=globaluserinfo&origin=%2A"
    );
    let response = get(transport, &url, DEFAULT_ACCEPT)?;
    if response.status != 200 {
        return Ok(unknown_response(&response));
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

pub(crate) fn check_devto(
    _: &Source,
    transport: &dyn Transport,
    label: &str,
) -> Result<Outcome, Error> {
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

pub(crate) fn check_minehut(
    _: &Source,
    transport: &dyn Transport,
    label: &str,
) -> Result<Outcome, Error> {
    let url = format!("https://api.minehut.com/server/{label}?byName=true");
    let response = get(transport, &url, DEFAULT_ACCEPT)?;
    if response.status != 200 {
        return Ok(unknown_response(&response));
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

pub(crate) fn check_bluesky(
    _: &Source,
    transport: &dyn Transport,
    label: &str,
) -> Result<Outcome, Error> {
    let url = format!(
        "https://bsky.social/xrpc/com.atproto.temp.checkHandleAvailability?handle={label}.bsky.social"
    );
    let response = get(transport, &url, DEFAULT_ACCEPT)?;
    if response.status != 200 {
        return Ok(unknown_response(&response));
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

pub(crate) fn check_opencollective(
    _: &Source,
    transport: &dyn Transport,
    label: &str,
) -> Result<Outcome, Error> {
    let url = format!("https://opencollective.com/{label}.json");
    let response = get(transport, &url, DEFAULT_ACCEPT)?;
    Ok(check_status(&response, Status::Taken).map_or_else(unknown, Outcome::new))
}

pub(crate) fn check_hackernews(
    _: &Source,
    transport: &dyn Transport,
    label: &str,
) -> Result<Outcome, Error> {
    // The profile page answers 403 "Sorry." to non-browser clients, so read the
    // official Firebase API: a free name is `null`, an account is an object.
    let url = format!("https://hacker-news.firebaseio.com/v0/user/{label}.json");
    let response = get(transport, &url, DEFAULT_ACCEPT)?;
    if response.status != 200 {
        return Ok(unknown_response(&response));
    }
    let Some(value) = json_body(&response) else {
        return Ok(unknown());
    };
    if value.is_null() {
        return Ok(available());
    }
    match value.get("id").and_then(Value::as_str) {
        Some(_) => Ok(taken()),
        None => Ok(unknown()),
    }
}

pub(crate) fn check_minds(
    _: &Source,
    transport: &dyn Transport,
    label: &str,
) -> Result<Outcome, Error> {
    let url = format!("https://www.minds.com/api/v1/channel/{label}");
    let response = get(transport, &url, DEFAULT_ACCEPT)?;
    if response.status != 200 {
        return Ok(unknown_response(&response));
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
pub(crate) fn check_snap(
    _: &Source,
    transport: &dyn Transport,
    label: &str,
) -> Result<Outcome, Error> {
    let mut headers = BTreeMap::new();
    headers.insert("Accept".to_string(), DEFAULT_ACCEPT.to_string());
    headers.insert("Snap-Device-Series".to_string(), "16".to_string());
    headers.insert("Accept-Language".to_string(), ACCEPT_LANGUAGE.to_string());
    headers.insert("User-Agent".to_string(), USER_AGENT.to_string());
    let response = transport.execute(HttpRequest {
        method: "GET".to_string(),
        url: format!("https://api.snapcraft.io/v2/snaps/info/{label}"),
        headers,
        body: None,
    })?;
    Ok(check_status(&response, Status::Taken).map_or_else(unknown, Outcome::new))
}

pub(crate) fn check_steam(
    _: &Source,
    transport: &dyn Transport,
    label: &str,
) -> Result<Outcome, Error> {
    let url = format!("https://steamcommunity.com/id/{label}?xml=1");
    let response = get(transport, &url, DEFAULT_ACCEPT)?;
    if response.status != 200 {
        return Ok(unknown_response(&response));
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

pub(crate) fn check_wordpress(
    _: &Source,
    transport: &dyn Transport,
    label: &str,
) -> Result<Outcome, Error> {
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

pub(crate) fn check_gitlab(
    _: &Source,
    transport: &dyn Transport,
    label: &str,
) -> Result<Outcome, Error> {
    let users_url = format!("https://gitlab.com/api/v4/users?username={label}");
    let users = get(transport, &users_url, DEFAULT_ACCEPT)?;
    if users.status != 200 {
        return Ok(unknown_response(&users));
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
            Ok(check_status(&group, Status::Taken)
                .map_or_else(|| unknown_response(&group), Outcome::new))
        }
    }
}

pub(crate) fn check_peertube(
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

pub(crate) fn check_tiktok(
    _: &Source,
    transport: &dyn Transport,
    label: &str,
) -> Result<Outcome, Error> {
    // The oEmbed endpoint is public and answers 400 for a name with no profile.
    let url =
        format!("https://www.tiktok.com/oembed?url=https%3A%2F%2Fwww.tiktok.com%2F%40{label}");
    let response = get(transport, &url, DEFAULT_ACCEPT)?;
    match response.status {
        200 => Ok(
            match json_body(&response)
                .as_ref()
                .and_then(|v| v.get("author_url"))
            {
                Some(_) => taken(),
                None => unknown(),
            },
        ),
        400 | 404 => Ok(available()),
        _ => Ok(unknown()),
    }
}

/// hnsscan answers 502 for the ICANN special-use names (probed 2026-11).
/// Nobody can register them, so they report unavailable rather than unknown.
pub(crate) const HNS_RESERVED: &[&str] = &["example", "invalid", "local", "localhost", "test"];

#[cfg(test)]
mod tests {
    use super::*;
    use crate::fixture::{Fixture, ReplayTransport};
    use crate::manifest::Manifest;
    use crate::provider::Status;

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
