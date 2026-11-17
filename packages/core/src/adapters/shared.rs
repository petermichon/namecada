//! Request helpers shared by every adapter family.

use std::collections::BTreeMap;

use serde_json::Value;

use crate::provider::{Error, HttpRequest, HttpResponse, Outcome, Status, Transport};

pub(crate) use super::status::{ACCEPT_LANGUAGE, DEFAULT_ACCEPT, USER_AGENT};

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
    headers.insert("Accept-Language".to_string(), ACCEPT_LANGUAGE.to_string());
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
    headers.insert("Accept-Language".to_string(), ACCEPT_LANGUAGE.to_string());
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
    pub(crate) response: HttpResponse,
    pub(crate) url: String,
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
