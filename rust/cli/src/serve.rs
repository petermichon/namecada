use std::collections::HashSet;
use std::error::Error;
use std::io::{self, Read, Write};
use std::path::PathBuf;
use std::sync::{mpsc, Arc};
use std::thread;
use std::time::{Duration, Instant};

use namecheck_core::{
    events, DoneEvent, Engine, Manifest, ManifestEvent, MemoryCache, ProviderResult, ResultEvent,
};
use serde::Serialize;
use serde_json::json;
use tiny_http::{Header, Method, Request, Response, Server, StatusCode};

use crate::providers;

const DEFAULT_ADDR: &str = "127.0.0.1:8080";

pub fn run(args: &[String]) -> Result<(), Box<dyn Error>> {
    let mut addr = DEFAULT_ADDR.to_string();
    let mut manifest_path: Option<PathBuf> = None;
    let mut budget: Option<Duration> = None;

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
            "--budget-ms" => {
                let value = args.get(i + 1).ok_or("--budget-ms requires milliseconds")?;
                let ms: u64 = value
                    .parse()
                    .map_err(|_| format!("invalid --budget-ms {value:?}"))?;
                budget = Some(Duration::from_millis(ms));
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
    let providers = providers::live(&manifest)?;
    let checked: Arc<HashSet<String>> = Arc::new(
        providers
            .iter()
            .map(|provider| provider.id().to_string())
            .collect(),
    );
    let manifest = Arc::new(manifest);
    let engine = Arc::new(Engine::new(providers).with_cache(Arc::new(MemoryCache::new())));
    let server = Server::http(&addr)
        .map_err(|err| -> Box<dyn Error> { format!("listen on {addr}: {err}").into() })?;
    eprintln!("listening on http://{addr}");

    for request in server.incoming_requests() {
        let engine = Arc::clone(&engine);
        let manifest = Arc::clone(&manifest);
        let checked = Arc::clone(&checked);
        thread::spawn(move || handle(request, &engine, &manifest, &checked, budget));
    }
    Ok(())
}

fn handle(
    request: Request,
    engine: &Engine,
    manifest: &Manifest,
    checked: &HashSet<String>,
    budget: Option<Duration>,
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
        .map(|value| value.trim().to_string())
        .filter(|value| !value.is_empty());
    let Some(label) = label else {
        respond_error(request, 400, "missing required query parameter \"name\"");
        return;
    };
    if !namecheck_core::valid_label(&label) {
        respond_error(request, 400, "invalid query parameter \"name\"");
        return;
    }

    let receiver = engine.scan(&label);
    stream_ndjson(request, manifest, checked, &label, receiver, budget);
}

/// stream_ndjson writes the response itself instead of going through tiny_http's
/// Response body. tiny_http sends chunked bodies through chunked_transfer's
/// Encoder, which buffers 8 KiB before writing: with one short line per provider
/// that meant the UI saw nothing until ~45 results had accumulated. Writing the
/// chunks here lets every event flush as it resolves.
fn stream_ndjson(
    request: Request,
    manifest: &Manifest,
    checked: &HashSet<String>,
    label: &str,
    receiver: mpsc::Receiver<ProviderResult>,
    budget: Option<Duration>,
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
    if let Err(err) = write_events(&mut writer, manifest, checked, label, receiver, budget) {
        eprintln!("response failed: {err}");
    }
}

fn write_events<W: Write>(
    writer: &mut W,
    manifest: &Manifest,
    checked: &HashSet<String>,
    label: &str,
    receiver: mpsc::Receiver<ProviderResult>,
    budget: Option<Duration>,
) -> io::Result<()> {
    write_event(
        writer,
        &ManifestEvent::new(label, events::source_events(manifest, checked, label)),
    )?;

    if let Some(budget) = budget {
        let deadline = Instant::now() + budget;
        loop {
            let remaining = deadline.saturating_duration_since(Instant::now());
            if remaining.is_zero() {
                break;
            }
            match receiver.recv_timeout(remaining) {
                Ok(result) => write_event(writer, &ResultEvent::new(&result))?,
                Err(mpsc::RecvTimeoutError::Timeout | mpsc::RecvTimeoutError::Disconnected) => {
                    break;
                }
            }
        }
    } else {
        for result in receiver {
            write_event(writer, &ResultEvent::new(&result))?;
        }
    }

    write_event(writer, &DoneEvent::default())?;
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

    fn manifest() -> Manifest {
        serde_json::from_value(serde_json::json!({
            "generatedAt": "2026-11-11T00:00:00Z",
            "label": "acema",
            "sources": [
                {"id": "npm", "platform": "npm", "platformName": "npm",
                 "name": "acema", "nameTemplate": "%s",
                 "url": "https://example.com/acema/npm",
                 "urlTemplate": "https://example.com/%s/npm",
                 "paid": false, "minLength": 0, "maxLength": 0, "keyed": false},
                {"id": "tiktok", "platform": "tiktok", "platformName": "TikTok",
                 "name": "acema", "nameTemplate": "%s",
                 "url": "https://example.com/acema/tiktok",
                 "urlTemplate": "https://example.com/%s/tiktok",
                 "paid": false, "minLength": 0, "maxLength": 0, "keyed": false}
            ]
        }))
        .unwrap()
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
        let manifest = manifest();
        let checked: HashSet<String> = ["npm".to_string()].into_iter().collect();
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
        write_events(&mut body, &manifest, &checked, "acema", rx, None).unwrap();
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
