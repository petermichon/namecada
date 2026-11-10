use std::error::Error;
use std::io::{self, Read};
use std::path::PathBuf;
use std::sync::{mpsc, Arc};
use std::thread;

use domain_browser_core::{Cache, Engine, Manifest, MemoryCache, Provider, ProviderResult};
use serde_json::json;
use tiny_http::{Header, Method, Request, Response, Server, StatusCode};

use crate::providers;

const DEFAULT_ADDR: &str = "127.0.0.1:8080";
const DEFAULT_ORIGIN: &str = "http://localhost:5173";

pub fn run(args: &[String]) -> Result<(), Box<dyn Error>> {
    let mut addr = DEFAULT_ADDR.to_string();
    let mut origin = DEFAULT_ORIGIN.to_string();
    let mut manifest_path = PathBuf::from("fixtures/sources.json");

    let mut i = 0;
    while i < args.len() {
        match args[i].as_str() {
            "--addr" => {
                addr = args.get(i + 1).ok_or("--addr requires an address")?.clone();
                i += 2;
            }
            "--origin" => {
                origin = args.get(i + 1).ok_or("--origin requires a URL")?.clone();
                i += 2;
            }
            "--manifest" => {
                manifest_path = PathBuf::from(args.get(i + 1).ok_or("--manifest requires a path")?);
                i += 2;
            }
            flag => return Err(format!("unknown flag {flag:?}").into()),
        }
    }

    let manifest = Manifest::load(manifest_path)?;
    let providers: Arc<Vec<Arc<dyn Provider>>> = Arc::new(providers::live(&manifest)?);
    let cache: Arc<dyn Cache> = Arc::new(MemoryCache::new());
    let server = Server::http(&addr)
        .map_err(|err| -> Box<dyn Error> { format!("listen on {addr}: {err}").into() })?;
    eprintln!("listening on http://{addr}");

    for request in server.incoming_requests() {
        let providers = Arc::clone(&providers);
        let cache = Arc::clone(&cache);
        let origin = origin.clone();
        thread::spawn(move || handle(request, providers.as_slice(), cache, &origin));
    }
    Ok(())
}

fn handle(request: Request, providers: &[Arc<dyn Provider>], cache: Arc<dyn Cache>, origin: &str) {
    let cors = allowed_origin(origin, &request);
    if request.method() == &Method::Options {
        let response = Response::new_empty(StatusCode(204))
            .with_header(header("Access-Control-Allow-Origin", &cors))
            .with_header(header("Access-Control-Allow-Methods", "GET, OPTIONS"))
            .with_header(header("Access-Control-Allow-Headers", "Content-Type"))
            .with_header(header("Vary", "Origin"));
        respond(request, response);
        return;
    }
    if request.method() != &Method::Get {
        respond_error(request, 405, "method not allowed", &cors);
        return;
    }
    if request.url().split('?').next() != Some("/check") {
        respond_error(request, 404, "not found", &cors);
        return;
    }

    let label = query_param(request.url(), "name")
        .map(|value| value.trim().to_ascii_lowercase())
        .filter(|value| !value.is_empty());
    let Some(label) = label else {
        respond_error(
            request,
            400,
            "missing required query parameter \"name\"",
            &cors,
        );
        return;
    };
    if !crate::valid_label(&label) {
        respond_error(request, 400, "invalid query parameter \"name\"", &cors);
        return;
    }

    let engine = Engine::new(providers.to_vec()).with_cache(cache);
    let receiver = engine.scan(&label);
    let response = Response::new(
        StatusCode(200),
        vec![
            header("Content-Type", "application/x-ndjson"),
            header("Cache-Control", "no-store"),
            header("Access-Control-Allow-Origin", &cors),
            header("Vary", "Origin"),
        ],
        NdjsonStream::new(receiver),
        None,
        None,
    );
    respond(request, response);
}

fn query_param(url: &str, key: &str) -> Option<String> {
    let parsed = url::Url::parse(&format!("http://localhost{url}")).ok()?;
    parsed
        .query_pairs()
        .find(|(name, _)| name == key)
        .map(|(_, value)| value.into_owned())
}

fn allowed_origin(configured: &str, request: &Request) -> String {
    for header in request.headers() {
        if header.field.equiv("Origin") {
            let origin = header.value.as_str();
            if is_loopback_origin(origin) {
                return origin.to_string();
            }
        }
    }
    configured.to_string()
}

fn is_loopback_origin(origin: &str) -> bool {
    let Ok(url) = url::Url::parse(origin) else {
        return false;
    };
    if url.scheme() != "http" || url.port().is_none() {
        return false;
    }
    match url.host() {
        Some(url::Host::Domain("localhost")) => true,
        Some(url::Host::Ipv4(ip)) => ip.is_loopback(),
        Some(url::Host::Ipv6(ip)) => ip.is_loopback(),
        _ => false,
    }
}

fn respond_error(request: Request, status: u16, message: &str, origin: &str) {
    let body = json!({ "error": message }).to_string();
    let response = Response::from_string(body)
        .with_status_code(StatusCode(status))
        .with_header(header("Content-Type", "application/json"))
        .with_header(header("Access-Control-Allow-Origin", origin))
        .with_header(header("Vary", "Origin"));
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

/// NdjsonStream turns the engine receiver into a chunked response body.
struct NdjsonStream {
    receiver: mpsc::Receiver<ProviderResult>,
    buffer: Vec<u8>,
    offset: usize,
}

impl NdjsonStream {
    fn new(receiver: mpsc::Receiver<ProviderResult>) -> Self {
        Self {
            receiver,
            buffer: Vec::new(),
            offset: 0,
        }
    }
}

impl Read for NdjsonStream {
    fn read(&mut self, out: &mut [u8]) -> io::Result<usize> {
        loop {
            if self.offset < self.buffer.len() {
                let available = &self.buffer[self.offset..];
                let count = available.len().min(out.len());
                out[..count].copy_from_slice(&available[..count]);
                self.offset += count;
                return Ok(count);
            }
            match self.receiver.recv() {
                Ok(result) => {
                    let mut line = serde_json::to_vec(&result).map_err(io::Error::other)?;
                    line.push(b'\n');
                    self.buffer = line;
                    self.offset = 0;
                }
                Err(_) => return Ok(0),
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use domain_browser_core::Status;

    #[test]
    fn parses_query_parameter() {
        assert_eq!(
            query_param("/check?name=acema", "name").as_deref(),
            Some("acema")
        );
        assert_eq!(
            query_param("/check?name=foo%20bar", "name").as_deref(),
            Some("foo bar")
        );
        assert_eq!(query_param("/check", "name"), None);
    }

    #[test]
    fn recognizes_loopback_origins() {
        assert!(is_loopback_origin("http://localhost:5173"));
        assert!(is_loopback_origin("http://127.0.0.1:5174"));
        assert!(is_loopback_origin("http://[::1]:5175"));
        assert!(!is_loopback_origin("https://localhost:5173"));
        assert!(!is_loopback_origin("http://localhost"));
        assert!(!is_loopback_origin("https://example.com"));
    }

    #[test]
    fn ndjson_stream_serializes_results() {
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

        let mut stream = NdjsonStream::new(rx);
        let mut body = String::new();
        stream.read_to_string(&mut body).unwrap();
        assert!(body.ends_with('\n'));
        assert!(body.contains("\"status\":\"available\""));
        assert!(body.contains("\"id\":\"npm\""));
    }
}
