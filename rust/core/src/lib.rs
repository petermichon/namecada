//! Parity-first engine core for the Domain Browser rewrite.
//!
//! The Go implementation remains the oracle while providers move here. The
//! replay transport lets every check run against the captured fixtures without
//! touching the network.

pub mod custom;
pub mod dns;
pub mod engine;
pub mod fixture;
pub mod http;
pub mod manifest;
pub mod provider;
pub mod status;

pub use custom::{build_custom_provider, custom_specs, CustomProvider, CustomSpec};
pub use dns::{build_dns_provider, dns_hosts, DnsError, DnsProvider, Resolver, SystemResolver};
pub use engine::{Engine, ProviderResult, DEFAULT_CONCURRENCY};
pub use fixture::{Fixture, ReplayTransport};
pub use http::HttpTransport;
pub use manifest::{Manifest, Source};
pub use provider::{Error, HttpRequest, HttpResponse, Outcome, Provider, Status, Transport};
pub use status::{
    build_status_provider, status_specs, CodeMap, StatusProvider, StatusSpec, DEFAULT_CODES,
    USER_AGENT,
};
