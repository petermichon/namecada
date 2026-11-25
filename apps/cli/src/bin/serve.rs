use std::collections::HashSet;
use std::error::Error;
use std::io::{self, Read, Write};
use std::sync::Arc;
use std::thread;

use namecada_core::{
    events, live_providers, DoneEvent, Engine, Jobs, Manifest, ManifestEvent, MemoryCache,
    PendingEvent, ResultEvent, ScanJob,
};
use serde::Serialize;
use serde_json::json;
use tiny_http::{Header, Method, Request, Response, Server, StatusCode};

use namecada::{load_manifest, run_or_exit, Flags};

const DEFAULT_ADDR: &str = "127.0.0.1:8080";

fn main() {
    run_or_exit(run());
}

fn run() -> Result<(), Box<dyn Error>> {
    let args: Vec<String> = std::env::args().skip(1).collect();
    let flags = Flags::parse(&args, &[("addr", true), ("manifest", true)])?;
    let addr = flags.value("addr").unwrap_or(DEFAULT_ADDR).to_string();
    let manifest = load_manifest(flags.value("manifest"))?;
    eprintln!(
        "registry: {}",
        flags.value("manifest").unwrap_or("embedded")
    );
    let providers = live_providers(&manifest)?;
    let checked: Arc<HashSet<String>> = Arc::new(
        providers
            .iter()
            .map(|provider| provider.id().to_string())
            .collect(),
    );
    let manifest = Arc::new(manifest);
    let engine = Arc::new(Engine::new(providers).with_cache(Arc::new(MemoryCache::new())));
    let jobs: Arc<Jobs> = Arc::new(Jobs::new());
    let server = Server::http(&addr)
        .map_err(|err| -> Box<dyn Error> { format!("listen on {addr}: {err}").into() })?;
    eprintln!("listening on http://{addr}");

    for request in server.incoming_requests() {
        let engine = Arc::clone(&engine);
        let manifest = Arc::clone(&manifest);
        let checked = Arc::clone(&checked);
        let jobs = Arc::clone(&jobs);
        thread::spawn(move || handle(request, &engine, &jobs, &manifest, &checked));
    }
    Ok(())
}

fn handle(
    request: Request,
    engine: &Arc<Engine>,
    jobs: &Jobs,
    manifest: &Manifest,
    checked: &HashSet<String>,
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
    if label.trim().is_empty() {
        respond_error(request, 400, "invalid query parameter \"name\"");
        return;
    }

    // A request either starts a scan or collects the next page of one. The job
    // itself is not tied to this connection: it keeps running after the
    // response closes.
    let scan = query_param(request.url(), "scan").filter(|value| !value.is_empty());
    let cursor = query_param(request.url(), "cursor")
        .and_then(|value| value.parse::<usize>().ok())
        .unwrap_or(0);
    jobs.reap();
    let job = scan
        .and_then(|id| jobs.get(&id))
        .filter(|job| job.label() == label)
        .unwrap_or_else(|| jobs.start(Arc::clone(engine), &label));
    stream_ndjson(request, manifest, checked, &label, &job, cursor);
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
    job: &ScanJob,
    cursor: usize,
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
    if let Err(err) = write_events(&mut writer, manifest, checked, label, job, cursor) {
        eprintln!("response failed: {err}");
    }
}

fn write_events<W: Write>(
    writer: &mut W,
    manifest: &Manifest,
    checked: &HashSet<String>,
    label: &str,
    job: &ScanJob,
    cursor: usize,
) -> io::Result<()> {
    write_event(
        writer,
        &ManifestEvent::new(label, events::source_events(manifest, checked, label)),
    )?;

    // Stream what the job already has, then keep the connection open for a
    // short budget so a small scan still finishes in one response. Once the
    // budget is spent the response ends with a cursor and the client asks for
    // the next page; the scan itself never stops.
    let started = std::time::Instant::now();
    let budget = std::time::Duration::from_secs(5);
    let mut next = cursor;
    loop {
        let results = job.results_from(next);
        for result in &results {
            write_event(writer, &ResultEvent::new(result))?;
        }
        next += results.len();
        if job.is_done() || started.elapsed() >= budget {
            break;
        }
        std::thread::sleep(std::time::Duration::from_millis(25));
    }

    if job.is_done() {
        write_event(writer, &DoneEvent::default())?;
    } else {
        write_event(writer, &PendingEvent::new(job.id(), next))?;
    }
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
    let query = url.split_once('?')?.1;
    query
        .split('&')
        .filter_map(|pair| pair.split_once('='))
        .find(|(name, _)| *name == key)
        .map(|(_, value)| percent_decode(value))
}

fn percent_decode(value: &str) -> String {
    let bytes = value.as_bytes();
    let mut out = Vec::with_capacity(bytes.len());
    let mut i = 0;
    while i < bytes.len() {
        if bytes[i] == b'%' && i + 2 < bytes.len() {
            if let Ok(byte) = u8::from_str_radix(&value[i + 1..i + 3], 16) {
                out.push(byte);
                i += 3;
                continue;
            }
        }
        out.push(if bytes[i] == b'+' { b' ' } else { bytes[i] });
        i += 1;
    }
    String::from_utf8_lossy(&out).into_owned()
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
    use namecada_core::provider::{HttpRequest, HttpResponse};
    use namecada_core::{build_status_provider, Error as CoreError, Transport};

    fn manifest() -> Manifest {
        serde_json::from_value(serde_json::json!({
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

    struct Ok200;

    impl Transport for Ok200 {
        fn execute(&self, _request: HttpRequest) -> Result<HttpResponse, CoreError> {
            Ok(HttpResponse {
                status: 200,
                headers: Default::default(),
                body: Vec::new(),
            })
        }
    }

    #[test]
    fn stream_starts_with_the_manifest_and_ends_with_done() {
        let manifest = manifest();
        let checked: HashSet<String> = ["npm".to_string()].into_iter().collect();
        let provider = build_status_provider(&manifest, "npm", Arc::new(Ok200)).unwrap();
        let engine = Arc::new(Engine::new(vec![Arc::new(provider)]));
        let job = ScanJob::start(engine, "acema".to_string());
        while !job.is_done() {
            thread::sleep(std::time::Duration::from_millis(5));
        }

        let mut body = Vec::new();
        write_events(&mut body, &manifest, &checked, "acema", &job, 0).unwrap();
        let body = String::from_utf8(body).unwrap();

        assert!(body.contains("\"type\":\"manifest\""));
        assert!(body.contains("\"sources\":[{"));
        assert!(body.contains("\"url\":\"https://example.com/acema/npm\""));
        assert!(body.contains("\"checked\":false"));
        assert!(body.contains("\"type\":\"result\""));
        assert!(body.contains("\"status\":\"taken\""));
        assert!(body.contains("\"type\":\"done\""));
        assert!(body.ends_with("0\r\n\r\n"));
    }
}
