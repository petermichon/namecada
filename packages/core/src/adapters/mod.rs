//! Provider adapters, one module per source shape.
//!
//! `shared` holds the request helpers, `spec` the id → check table.

mod api;
pub mod dns;
mod domains;
mod page;
mod shared;
mod spec;
pub mod status;

pub use spec::{build_custom_provider, custom_specs};
