//! Fixture loading, replaying and recording for parity tests.

use std::collections::{BTreeMap, VecDeque};
use std::path::Path;
use std::sync::{Arc, Mutex};

use base64::engine::general_purpose::STANDARD as BASE64;
use base64::Engine;
use serde::{Deserialize, Serialize};

use crate::provider::{Error, HttpRequest, HttpResponse, Status, Transport};

/// BODY_CAP is the largest body a fixture stores, matching the exporter.
const BODY_CAP: usize = 256 * 1024;

/// Fixture is one captured provider result and its request transcript.
#[derive(Debug, Clone, Deserialize, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct Fixture {
    pub source: String,
    pub platform: String,
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub platform_name: String,
    pub label: String,
    pub name: String,
    pub url: String,
    #[serde(default)]
    pub paid: bool,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub status: Option<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub detail: Option<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
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

    /// save writes the fixture as the exporter did: two-space JSON, one
    /// newline at the end.
    pub fn save(&self, path: impl AsRef<Path>) -> Result<(), Error> {
        let mut json = serde_json::to_string_pretty(self)?;
        json.push('\n');
        std::fs::write(path, json)?;
        Ok(())
    }

    pub fn expected_status(&self) -> Option<Status> {
        self.status.as_deref().and_then(Status::parse)
    }
}

/// RecordedRequest is one request/response pair from a fixture.
#[derive(Debug, Clone, Deserialize, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct RecordedRequest {
    pub method: String,
    pub url: String,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub headers: Option<BTreeMap<String, String>>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub body: Option<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub body_encoding: Option<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub response: Option<RecordedResponse>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub error: Option<String>,
}

/// RecordedResponse is the captured part of an HTTP response.
#[derive(Debug, Clone, Deserialize, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct RecordedResponse {
    pub status: u16,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub headers: Option<BTreeMap<String, String>>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub body: Option<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub body_encoding: Option<String>,
    #[serde(default, skip_serializing_if = "is_false")]
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

/// RecordingTransport wraps another transport and records every exchange, so
/// a provider can be run once and its transcript written as a fixture.
pub struct RecordingTransport {
    inner: Arc<dyn Transport>,
    requests: Mutex<Vec<RecordedRequest>>,
}

impl RecordingTransport {
    pub fn new(inner: Arc<dyn Transport>) -> Self {
        Self {
            inner,
            requests: Mutex::new(Vec::new()),
        }
    }

    pub fn take(&self) -> Vec<RecordedRequest> {
        std::mem::take(&mut *self.requests.lock().expect("recording queue poisoned"))
    }
}

impl Transport for RecordingTransport {
    fn execute(&self, request: HttpRequest) -> Result<HttpResponse, Error> {
        let (body, body_encoding) = match request.body.as_deref() {
            Some(bytes) => {
                let (text, encoding) = encode_body(bytes);
                (Some(text), Some(encoding.to_string()))
            }
            None => (None, None),
        };
        let result = self.inner.execute(request.clone());
        let recorded = match &result {
            Ok(response) => RecordedRequest {
                method: request.method,
                url: request.url,
                headers: Some(request.headers),
                body,
                body_encoding,
                response: Some(record(response)),
                error: None,
            },
            Err(err) => RecordedRequest {
                method: request.method,
                url: request.url,
                headers: Some(request.headers),
                body,
                body_encoding,
                response: None,
                error: Some(err.to_string()),
            },
        };
        self.requests
            .lock()
            .expect("recording queue poisoned")
            .push(recorded);
        result
    }
}

fn record(response: &HttpResponse) -> RecordedResponse {
    let mut headers = BTreeMap::new();
    for (key, value) in &response.headers {
        let lower = key.to_ascii_lowercase();
        if lower == "content-type" || lower == "location" {
            headers.insert(key.clone(), value.clone());
        }
    }
    let truncated = response.body.len() > BODY_CAP;
    let body = &response.body[..response.body.len().min(BODY_CAP)];
    let (text, encoding) = encode_body(body);
    RecordedResponse {
        status: response.status,
        headers: Some(headers),
        body: Some(text),
        body_encoding: Some(encoding.to_string()),
        body_truncated: truncated,
    }
}

/// encode_body stores UTF-8 as text and anything else as base64.
fn encode_body(body: &[u8]) -> (String, &'static str) {
    match std::str::from_utf8(body) {
        Ok(text) => (text.to_string(), "utf8"),
        Err(_) => (BASE64.encode(body), "base64"),
    }
}

fn is_false(value: &bool) -> bool {
    !*value
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

#[cfg(test)]
mod tests {
    use super::*;

    struct Stub;

    impl Transport for Stub {
        fn execute(&self, request: HttpRequest) -> Result<HttpResponse, Error> {
            if request.url.contains("broken") {
                return Err(Error::Transport("boom".to_string()));
            }
            let mut headers = BTreeMap::new();
            headers.insert("location".to_string(), "/next".to_string());
            headers.insert("x-noise".to_string(), "drop me".to_string());
            Ok(HttpResponse {
                status: 429,
                headers,
                body: b"rate limited".to_vec(),
            })
        }
    }

    #[test]
    fn recording_captures_the_exchange() {
        let recorder = RecordingTransport::new(Arc::new(Stub));
        let request = |url: &str| HttpRequest {
            method: "GET".to_string(),
            url: url.to_string(),
            headers: BTreeMap::from([("Accept".to_string(), "text/html".to_string())]),
            body: None,
        };

        recorder
            .execute(request("https://example.test/ok"))
            .expect("ok");
        recorder
            .execute(request("https://example.test/broken"))
            .expect_err("boom");

        let recorded = recorder.take();
        assert_eq!(recorded.len(), 2);

        let ok = &recorded[0];
        assert_eq!(ok.method, "GET");
        assert_eq!(ok.url, "https://example.test/ok");
        assert_eq!(ok.headers.as_ref().unwrap()["Accept"], "text/html");
        let response = ok.response.as_ref().expect("response");
        assert_eq!(response.status, 429);
        assert_eq!(response.body.as_deref(), Some("rate limited"));
        assert_eq!(response.body_encoding.as_deref(), Some("utf8"));
        let headers = response.headers.as_ref().expect("headers");
        assert_eq!(headers["location"], "/next");
        assert!(
            !headers.contains_key("x-noise"),
            "only content-type/location"
        );
        assert!(!response.body_truncated);

        let broken = &recorded[1];
        assert!(broken.response.is_none());
        assert_eq!(broken.error.as_deref(), Some("transport: boom"));
        assert!(recorder.take().is_empty(), "take drains the queue");
    }
}
