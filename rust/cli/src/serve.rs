use std::collections::HashSet;
use std::error::Error;
use std::io::{self, Read, Write};
use std::path::PathBuf;
use std::sync::{mpsc, Arc};
use std::thread;

use namecheck_core::{Cache, Engine, Manifest, MemoryCache, Provider, ProviderResult};
use serde::Serialize;
use serde_json::json;
use tiny_http::{Header, Method, Request, Response, Server, StatusCode};

use crate::providers;

const DEFAULT_ADDR: &str = "127.0.0.1:8080";
const API_VERSION: u8 = 1;

/// SourceEntry is one row of the manifest streamed before a scan. It comes from
/// the registry, not the provider set, so deferred and keyed sources are listed
/// too — the client renders them as "not checked" instead of guessing.
struct SourceEntry {
    id: String,
    platform_name: String,
    url_template: String,
    paid: bool,
    checked: bool,
}

pub fn run(args: &[String]) -> Result<(), Box<dyn Error>> {
    let mut addr = DEFAULT_ADDR.to_string();
    let mut manifest_path: Option<PathBuf> = None;

    let mut i = 0;
    while i < args.len() {
        match args[i].as_str() {
            "--addr" => {
                addr = args.get(i + 1).ok_or("--addr requires an address")?.clone();
                i += 2;
            }
            "--manifest" => {
                manifest_path = Some(PathBuf::from(
                    args.get(i + 1).ok_or("--manifest requires a path")?,
                ));
                i += 2;
            }
            flag => return Err(format!("unknown flag {flag:?}").into()),
        }
    }

    let manifest = match &manifest_path {
        Some(path) => Manifest::load(path)?,
        None => Manifest::embedded()?,
    };
    manifest.validate()?;
    eprintln!(
        "registry: {}",
        manifest_path
            .as_ref()
            .map(|path| path.display().to_string())
            .unwrap_or_else(|| "embedded".to_string())
    );
    let providers: Arc<Vec<Arc<dyn Provider>>> = Arc::new(providers::live(&manifest)?);
    let entries: Arc<Vec<SourceEntry>> = Arc::new(source_entries(&manifest, &providers));
    let cache: Arc<dyn Cache> = Arc::new(MemoryCache::new());
    let server = Server::http(&addr)
        .map_err(|err| -> Box<dyn Error> { format!("listen on {addr}: {err}").into() })?;
    eprintln!("listening on http://{addr}");

    for request in server.incoming_requests() {
        let providers = Arc::clone(&providers);
        let entries = Arc::clone(&entries);
        let cache = Arc::clone(&cache);
        thread::spawn(move || handle(request, providers.as_slice(), &entries, cache));
    }
    Ok(())
}

/// source_entries joins the registry with the providers that were actually
/// built. `checked` is derived, never stored, so it cannot drift.
fn source_entries(manifest: &Manifest, providers: &[Arc<dyn Provider>]) -> Vec<SourceEntry> {
    let checked: HashSet<&str> = providers.iter().map(|provider| provider.id()).collect();
    manifest
        .sources
        .iter()
        .map(|source| SourceEntry {
            id: source.id.clone(),
            platform_name: source.platform_name.clone(),
            url_template: source.url_template.clone(),
            paid: source.paid,
            checked: checked.contains(source.id.as_str()),
        })
        .collect()
}

fn handle(
    request: Request,
    providers: &[Arc<dyn Provider>],
    entries: &[SourceEntry],
    cache: Arc<dyn Cache>,
) {
    if request.method() != &Method::Get {
        respond_error(request, 405, "method not allowed");
        return;
    }

    if request.url().split('?').next() != Some("/v1/check") {
        respond_error(request, 404, "not found");
        return;
    }

    let label = query_param(request.url(), "name")
        .map(|value| value.trim().to_ascii_lowercase())
        .filter(|value| !value.is_empty());
    let Some(label) = label else {
        respond_error(request, 400, "missing required query parameter \"name\"");
        return;
    };
    if !crate::valid_label(&label) {
        respond_error(request, 400, "invalid query parameter \"name\"");
        return;
    }

    let engine = Engine::new(providers.to_vec()).with_cache(cache);
    let receiver = engine.scan(&label);
    stream_ndjson(request, entries, &label, receiver);
}

/// stream_ndjson writes the response itself instead of going through tiny_http's
/// Response body. tiny_http sends chunked bodies through chunked_transfer's
/// Encoder, which buffers 8 KiB before writing: with one short line per provider
/// that meant the UI saw nothing until ~45 results had accumulated. Writing the
/// chunks here lets every event flush as it resolves.
fn stream_ndjson(
    request: Request,
    entries: &[SourceEntry],
    label: &str,
    receiver: mpsc::Receiver<ProviderResult>,
) {
    let mut writer = request.into_writer();
    let head = "HTTP/1.1 200 OK\r\n\
                Content-Type: application/x-ndjson\r\n\
                Cache-Control: no-store\r\n\
                Transfer-Encoding: chunked\r\n\
                Connection: close\r\n\r\n";
    if writer.write_all(head.as_bytes()).is_err() {
        return;
    }
    if let Err(err) = write_events(&mut writer, entries, label, receiver) {
        eprintln!("response failed: {err}");
    }
}

#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
struct ManifestEvent<'a> {
    #[serde(rename = "type")]
    kind: &'static str,
    version: u8,
    label: &'a str,
    sources: Vec<SourceEvent<'a>>,
}

#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
struct SourceEvent<'a> {
    id: &'a str,
    platform_name: &'a str,
    url: String,
    paid: bool,
    checked: bool,
}

#[derive(Serialize)]
struct ResultEvent<'a> {
    #[serde(rename = "type")]
    kind: &'static str,
    #[serde(flatten)]
    result: &'a ProviderResult,
}

#[derive(Serialize)]
struct DoneEvent {
    #[serde(rename = "type")]
    kind: &'static str,
}

fn write_events<W: Write>(
    writer: &mut W,
    entries: &[SourceEntry],
    label: &str,
    receiver: mpsc::Receiver<ProviderResult>,
) -> io::Result<()> {
    let sources = entries
        .iter()
        .map(|entry| SourceEvent {
            id: &entry.id,
            platform_name: &entry.platform_name,
            url: entry.url_template.replace("%s", label),
            paid: entry.paid,
            checked: entry.checked,
        })
        .collect();
    write_event(
        writer,
        &ManifestEvent {
            kind: "manifest",
            version: API_VERSION,
            label,
            sources,
        },
    )?;

    for result in receiver {
        write_event(
            writer,
            &ResultEvent {
                kind: "result",
                result: &result,
            },
        )?;
    }

    write_event(writer, &DoneEvent { kind: "done" })?;
    writer.write_all(b"0\r\n\r\n")?;
    writer.flush()
}

/// write_event writes one NDJSON line as a single HTTP chunk and flushes it.
fn write_event<W: Write, T: Serialize>(writer: &mut W, event: &T) -> io::Result<()> {
    let mut line = serde_json::to_vec(event).map_err(io::Error::other)?;
    line.push(b'\n');
    write!(writer, "{:X}\r\n", line.len())?;
    writer.write_all(&line)?;
    writer.write_all(b"\r\n")?;
    writer.flush()
}

fn query_param(url: &str, key: &str) -> Option<String> {
    let parsed = url::Url::parse(&format!("http://localhost{url}")).ok()?;
    parsed
        .query_pairs()
        .find(|(name, _)| name == key)
        .map(|(_, value)| value.into_owned())
}

fn respond_error(request: Request, status: u16, message: &str) {
    let body = json!({ "error": message }).to_string();
    let response = Response::from_string(body)
        .with_status_code(StatusCode(status))
        .with_header(header("Content-Type", "application/json"));
    respond(request, response);
}

fn respond<R: Read>(request: Request, response: Response<R>) {
    if let Err(err) = request.respond(response) {
        eprintln!("response failed: {err}");
    }
}

fn header(name: &str, value: &str) -> Header {
    Header::from_bytes(name.as_bytes(), value.as_bytes()).expect("valid header")
}

#[cfg(test)]
mod tests {
    use super::*;
    use namecheck_core::Status;

    fn entry(id: &str, checked: bool) -> SourceEntry {
        SourceEntry {
            id: id.to_string(),
            platform_name: id.to_string(),
            url_template: format!("https://example.com/%s/{id}"),
            paid: false,
            checked,
        }
    }

    #[test]
    fn parses_query_parameter() {
        assert_eq!(
            query_param("/v1/check?name=acema", "name").as_deref(),
            Some("acema")
        );
        assert_eq!(
            query_param("/v1/check?name=foo%20bar", "name").as_deref(),
            Some("foo bar")
        );
        assert_eq!(query_param("/v1/check", "name"), None);
    }

    #[test]
    fn stream_starts_with_the_manifest_and_ends_with_done() {
        let entries = vec![entry("npm", true), entry("tiktok", false)];
        let (tx, rx) = mpsc::channel();
        tx.send(ProviderResult {
            id: "npm".to_string(),
            platform: "npm".to_string(),
            platform_name: "npm".to_string(),
            name: "acema".to_string(),
            url: "https://www.npmjs.com/package/acema".to_string(),
            status: Status::Available,
            detail: String::new(),
            paid: false,
        })
        .unwrap();
        drop(tx);

        let mut body = Vec::new();
        write_events(&mut body, &entries, "acema", rx).unwrap();
        let body = String::from_utf8(body).unwrap();

        assert!(body.contains("\"type\":\"manifest\""));
        assert!(body.contains("\"sources\":[{"));
        assert!(body.contains("\"url\":\"https://example.com/acema/npm\""));
        assert!(body.contains("\"checked\":false"));
        assert!(body.contains("\"type\":\"result\""));
        assert!(body.contains("\"status\":\"available\""));
        assert!(body.contains("\"type\":\"done\""));
        assert!(body.ends_with("0\r\n\r\n"));
    }
}
