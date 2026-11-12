//! Core contract shared by every source: the provider trait, the transport
//! seam, and the result types.

use std::collections::BTreeMap;
use std::fmt;

use serde::{Deserialize, Serialize};

/// Status is the outcome of one availability check.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum Status {
    Available,
    Taken,
    Unknown,
}

impl Status {
    pub fn as_str(self) -> &'static str {
        match self {
            Status::Available => "available",
            Status::Taken => "taken",
            Status::Unknown => "unknown",
        }
    }

    pub fn parse(value: &str) -> Option<Self> {
        match value {
            "available" => Some(Status::Available),
            "taken" => Some(Status::Taken),
            "unknown" => Some(Status::Unknown),
            _ => None,
        }
    }
}

/// Outcome is what a provider reports for a label.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Outcome {
    pub status: Status,
    /// Optional explanation streamed alongside the status.
    pub detail: String,
    /// Overrides the provider's default public URL when non-empty.
    pub url: Option<String>,
}

impl Outcome {
    pub fn new(status: Status) -> Self {
        Self {
            status,
            detail: String::new(),
            url: None,
        }
    }

    /// with_detail explains a non-definitive verdict (for example why a check
    /// ended in `unknown`), so the UI can say more than "unknown".
    pub fn with_detail(self, detail: impl Into<String>) -> Self {
        Self {
            detail: detail.into(),
            ..self
        }
    }
}

/// HttpRequest is one outbound request. The transport owns headers and body so
/// providers stay independent of the HTTP client.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct HttpRequest {
    pub method: String,
    pub url: String,
    pub headers: BTreeMap<String, String>,
    pub body: Option<Vec<u8>>,
}

/// HttpResponse is the part of a response a provider may inspect.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct HttpResponse {
    pub status: u16,
    pub headers: BTreeMap<String, String>,
    pub body: Vec<u8>,
}

/// Transport executes requests. Production uses HTTP; parity tests replay the
/// captured fixtures.
pub trait Transport: Send + Sync {
    fn execute(&self, request: HttpRequest) -> Result<HttpResponse, Error>;
}

/// Provider checks whether a label is taken on one source.
pub trait Provider: Send + Sync {
    fn id(&self) -> &str;
    fn platform(&self) -> &str;
    fn platform_name(&self) -> String {
        self.platform().to_string()
    }
    fn name(&self, label: &str) -> String;
    fn url(&self, label: &str) -> String;
    fn paid(&self) -> bool;

    /// normalize_label maps the user's input to the form this source expects.
    /// Case-insensitive sources fold to lowercase; a source whose namespace is
    /// case-sensitive keeps the raw label.
    fn normalize_label(&self, label: &str) -> String {
        label.to_ascii_lowercase()
    }

    fn min_length(&self) -> usize {
        0
    }

    fn max_length(&self) -> Option<usize> {
        None
    }

    fn check(&self, label: &str) -> Result<Outcome, Error>;
}

/// Error covers manifest/fixture IO, replay mismatches, and transport failures.
#[derive(Debug)]
pub enum Error {
    Io(std::io::Error),
    Json(serde_json::Error),
    Base64(base64::DecodeError),
    Url(url::ParseError),
    NotFound(String),
    Invalid(String),
    Transport(String),
    RequestMismatch {
        expected_method: String,
        expected_url: String,
        actual_method: String,
        actual_url: String,
    },
    UnexpectedRequest(String),
    MissingResponse(String),
}

impl fmt::Display for Error {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Error::Io(err) => write!(f, "io: {err}"),
            Error::Json(err) => write!(f, "json: {err}"),
            Error::Base64(err) => write!(f, "base64: {err}"),
            Error::Url(err) => write!(f, "url: {err}"),
            Error::NotFound(what) => write!(f, "not found: {what}"),
            Error::Invalid(what) => write!(f, "invalid: {what}"),
            Error::Transport(err) => write!(f, "transport: {err}"),
            Error::RequestMismatch {
                expected_method,
                expected_url,
                actual_method,
                actual_url,
            } => write!(
                f,
                "request mismatch: expected {expected_method} {expected_url}, got {actual_method} {actual_url}"
            ),
            Error::UnexpectedRequest(url) => write!(f, "unexpected request: {url}"),
            Error::MissingResponse(url) => write!(f, "response missing for {url}"),
        }
    }
}

impl std::error::Error for Error {}

impl From<std::io::Error> for Error {
    fn from(value: std::io::Error) -> Self {
        Error::Io(value)
    }
}

impl From<serde_json::Error> for Error {
    fn from(value: serde_json::Error) -> Self {
        Error::Json(value)
    }
}

impl From<base64::DecodeError> for Error {
    fn from(value: base64::DecodeError) -> Self {
        Error::Base64(value)
    }
}

impl From<url::ParseError> for Error {
    fn from(value: url::ParseError) -> Self {
        Error::Url(value)
    }
}
