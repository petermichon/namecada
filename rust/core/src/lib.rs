//! Parity-first engine core.
//!
//! The captured fixtures are the oracle: every provider can replay them without
//! touching the network.

pub mod adapters;
pub mod engine;
pub mod events;
pub mod fixture;
pub mod http;
pub mod label;
pub mod manifest;
pub mod provider;
pub mod registry;

pub use adapters::dns::{build_dns_provider, dns_hosts, DnsError, Resolver, SystemResolver};
pub use adapters::status::{build_status_provider, status_specs};
pub use adapters::{build_custom_provider, custom_specs};
pub use engine::{Engine, MemoryCache, ProviderResult, DEFAULT_CONCURRENCY};
pub use events::{DoneEvent, ManifestEvent, ResultEvent, SourceEvent};
pub use fixture::{Fixture, RecordingTransport, ReplayTransport};
pub use http::HttpTransport;
pub use label::valid_label;
pub use manifest::Manifest;
pub use provider::{Error, Outcome, Provider, Status, Transport};
pub use registry::{live as live_providers, stats as registry_stats, RegistryStats};
