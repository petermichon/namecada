//! Fixture loading and a replay transport for parity tests.

use std::collections::{BTreeMap, VecDeque};
use std::path::Path;
use std::sync::Mutex;

use base64::engine::general_purpose::STANDARD as BASE64;
use base64::Engine;
use serde::Deserialize;

use crate::provider::{Error, HttpRequest, HttpResponse, Status, Transport};

/// Fixture is one captured provider result and its request transcript.
#[derive(Debug, Clone, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Fixture {
    pub source: String,
    pub platform: String,
    #[serde(default)]
    pub platform_name: String,
    pub label: String,
    pub name: String,
    pub url: String,
    #[serde(default)]
    pub paid: bool,
    #[serde(default)]
    pub status: Option<String>,
    #[serde(default)]
    pub detail: Option<String>,
    #[serde(default)]
    pub error: Option<String>,
    #[serde(default)]
    pub duration_ms: u64,
    #[serde(default)]
    pub requests: Vec<RecordedRequest>,
}

impl Fixture {
    pub fn load(path: impl AsRef<Path>) -> Result<Self, Error> {
        let data = std::fs::read(path)?;
        Ok(serde_json::from_slice(&data)?)
    }

    pub fn expected_status(&self) -> Option<Status> {
        self.status.as_deref().and_then(Status::parse)
    }
}

/// RecordedRequest is one request/response pair from a fixture.
#[derive(Debug, Clone, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct RecordedRequest {
    pub method: String,
    pub url: String,
    #[serde(default)]
    pub headers: Option<BTreeMap<String, String>>,
    #[serde(default)]
    pub body: Option<String>,
    #[serde(default)]
    pub body_encoding: Option<String>,
    #[serde(default)]
    pub response: Option<RecordedResponse>,
    #[serde(default)]
    pub error: Option<String>,
}

/// RecordedResponse is the captured part of an HTTP response.
#[derive(Debug, Clone, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct RecordedResponse {
    pub status: u16,
    #[serde(default)]
    pub headers: Option<BTreeMap<String, String>>,
    #[serde(default)]
    pub body: Option<String>,
    #[serde(default)]
    pub body_encoding: Option<String>,
    #[serde(default)]
    pub body_truncated: bool,
}

/// ReplayTransport serves a fixture's requests in order and fails on a
/// mismatch, so a provider cannot silently drift from the captured flow.
pub struct ReplayTransport {
    source: String,
    requests: Mutex<VecDeque<RecordedRequest>>,
}

impl ReplayTransport {
    pub fn new(fixture: &Fixture) -> Self {
        Self {
            source: fixture.source.clone(),
            requests: Mutex::new(fixture.requests.iter().cloned().collect()),
        }
    }

    pub fn source(&self) -> &str {
        &self.source
    }

    pub fn remaining(&self) -> usize {
        self.requests.lock().expect("replay queue poisoned").len()
    }
}

impl Transport for ReplayTransport {
    fn execute(&self, request: HttpRequest) -> Result<HttpResponse, Error> {
        let recorded = {
            let mut queue = self.requests.lock().expect("replay queue poisoned");
            queue.pop_front()
        };
        let Some(recorded) = recorded else {
            return Err(Error::UnexpectedRequest(request.url));
        };

        if recorded.method != request.method || recorded.url != request.url {
            return Err(Error::RequestMismatch {
                expected_method: recorded.method,
                expected_url: recorded.url,
                actual_method: request.method,
                actual_url: request.url,
            });
        }

        if let Some(err) = recorded.error {
            return Err(Error::Transport(err));
        }
        let Some(response) = recorded.response else {
            return Err(Error::MissingResponse(recorded.url));
        };

        let mut headers = BTreeMap::new();
        for (key, value) in response.headers.unwrap_or_default() {
            headers.insert(key.to_ascii_lowercase(), value);
        }
        let body = decode_body(
            response.body.as_deref().unwrap_or(""),
            response.body_encoding.as_deref(),
        )?;

        Ok(HttpResponse {
            status: response.status,
            headers,
            body,
        })
    }
}

fn decode_body(body: &str, encoding: Option<&str>) -> Result<Vec<u8>, Error> {
    if body.is_empty() {
        return Ok(Vec::new());
    }
    match encoding.unwrap_or("utf8") {
        "utf8" => Ok(body.as_bytes().to_vec()),
        "base64" => Ok(BASE64.decode(body)?),
        other => Err(Error::Invalid(format!("unknown body encoding {other:?}"))),
    }
}
