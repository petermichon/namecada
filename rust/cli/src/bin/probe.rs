//! probe reports which sources still have no provider: it fetches every
//! "not checked" row with a label that should exist and a control label that
//! should not, then sorts the pair into differs / challenge / blocked / keyed /
//! shell / dead. Promotion stays manual (spec table, fixture, docs).
//!
//! `--save` keeps a local history under `.probe/` and `--diff` prints what
//! changed since the previous run, so drift is visible without a scheduler.
//!
//! Run it with `npm run probe`.

use std::collections::BTreeMap;
use std::path::PathBuf;
use std::sync::atomic::{AtomicUsize, Ordering};
use std::sync::{Arc, Mutex};
use std::time::{Duration, SystemTime, UNIX_EPOCH};

use namecada_core::adapters::status::{ACCEPT_LANGUAGE, USER_AGENT};
use namecada_core::provider::HttpRequest;
use namecada_core::{live_providers, HttpTransport, Manifest, Transport};
use serde::{Deserialize, Serialize};
use url::Url;

const ACCEPT: &str = "text/html,application/xhtml+xml,application/json;q=0.9,*/*;q=0.8";
const BODY_CAP: usize = 120_000;

const BOT_MARKERS: &[&str] = &[
    "just a moment",
    "attention required",
    "client challenge",
    "cf-chl",
    "verifying you are human",
    "checking your browser",
    "security checkpoint",
    "access denied",
    "request blocked",
    "captcha",
    "are you a robot",
    "unusual traffic",
];

const SOFT_404: &[&str] = &[
    "page not found",
    "not found",
    "doesn't exist",
    "no longer available",
];

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
struct Row {
    id: String,
    platform_name: String,
    verdict: String,
    real: u16,
    fake: u16,
    url: String,
    title: String,
    #[serde(default)]
    note: String,
}

struct Fetched {
    status: u16,
    url: String,
    body: String,
    #[allow(dead_code)]
    error: Option<String>,
}

struct Args {
    label: String,
    fake: String,
    only: Option<Vec<String>>,
    jobs: usize,
    timeout: Duration,
    json: bool,
    save: bool,
    diff: bool,
}

fn main() {
    let args = match parse_args() {
        Ok(args) => args,
        Err(message) => {
            eprintln!("probe: {message}");
            std::process::exit(2);
        }
    };

    let manifest = match Manifest::embedded() {
        Ok(manifest) => manifest,
        Err(err) => {
            eprintln!("probe: cannot read the registry: {err}");
            std::process::exit(1);
        }
    };
    let checked = checked_ids(&manifest);
    let mut rows: Vec<_> = manifest
        .sources
        .iter()
        .filter(|source| !checked.contains(&source.id))
        .filter(|source| match &args.only {
            Some(only) => only.iter().any(|id| id == &source.id),
            None => true,
        })
        .collect();
    rows.sort_by(|a, b| a.id.cmp(&b.id));

    if rows.is_empty() {
        let scope = args
            .only
            .as_ref()
            .map(|only| format!(" matching {}", only.join(", ")))
            .unwrap_or_default();
        eprintln!("probe: no unchecked sources{scope}");
        std::process::exit(1);
    }

    let transport: Arc<dyn Transport> = Arc::new(HttpTransport::with_timeout(args.timeout));
    let results = run(&rows, &transport, &args);
    let mut results = results;

    if args.diff {
        print_changes(&results, args.only.is_none());
    }
    if args.save {
        match save_run(&results, &args) {
            Ok(path) => eprintln!("saved {}", path.display()),
            Err(err) => eprintln!("probe: cannot save the run: {err}"),
        }
    }
    if args.json {
        println!(
            "{}",
            serde_json::to_string_pretty(&results).expect("serialize results")
        );
        return;
    }

    results.sort_by_key(|row| (rank(&row.verdict), row.id.clone()));
    for row in &results {
        let note = if row.note.is_empty() {
            String::new()
        } else {
            format!("  ({})", row.note)
        };
        println!(
            "{:<11} {:<21} {:>3}/{:<3} {}{note}",
            row.verdict, row.id, row.real, row.fake, row.title
        );
    }

    let mut counts: BTreeMap<String, usize> = BTreeMap::new();
    for row in &results {
        let key = row.verdict.split(' ').next().unwrap_or("").to_string();
        *counts.entry(key).or_default() += 1;
    }
    let summary = counts
        .iter()
        .map(|(key, count)| format!("{count} {key}"))
        .collect::<Vec<_>>()
        .join(", ");
    println!("\n{} unchecked sources probed ({summary})", results.len());
    if counts.contains_key("differs") {
        println!("differs = candidate: confirm the mapping, then add a spec, a fixture and docs");
    }
}

fn parse_args() -> Result<Args, String> {
    let mut args = Args {
        label: "google".to_string(),
        fake: "zzqxnvc".to_string(),
        only: None,
        jobs: 4,
        timeout: Duration::from_secs(20),
        json: false,
        save: false,
        diff: false,
    };
    let mut rest = std::env::args().skip(1);
    while let Some(flag) = rest.next() {
        let mut value = || rest.next().ok_or_else(|| format!("{flag} needs a value"));
        match flag.as_str() {
            "--label" => args.label = value()?,
            "--fake" => args.fake = value()?,
            "--only" => {
                args.only = Some(
                    value()?
                        .split(',')
                        .map(str::trim)
                        .filter(|id| !id.is_empty())
                        .map(str::to_string)
                        .collect(),
                )
            }
            "--jobs" => args.jobs = value()?.parse().map_err(|_| "--jobs wants a number")?,
            "--timeout" => {
                let seconds: f64 = value()?.parse().map_err(|_| "--timeout wants seconds")?;
                args.timeout = Duration::from_secs_f64(seconds);
            }
            "--json" => args.json = true,
            "--save" => args.save = true,
            "--diff" => args.diff = true,
            "-h" | "--help" => {
                print!("{HELP}");
                std::process::exit(0);
            }
            other => return Err(format!("unknown argument {other}; try --help")),
        }
    }
    args.jobs = args.jobs.max(1);
    Ok(args)
}

fn checked_ids(manifest: &Manifest) -> Vec<String> {
    live_providers(manifest)
        .map(|providers| providers.iter().map(|p| p.id().to_string()).collect())
        .unwrap_or_default()
}

fn run(
    rows: &[&namecada_core::manifest::Source],
    transport: &Arc<dyn Transport>,
    args: &Args,
) -> Vec<Row> {
    let index = AtomicUsize::new(0);
    let results = Mutex::new(Vec::with_capacity(rows.len()));
    std::thread::scope(|scope| {
        for _ in 0..args.jobs.min(rows.len()) {
            scope.spawn(|| loop {
                let position = index.fetch_add(1, Ordering::Relaxed);
                let Some(source) = rows.get(position) else {
                    break;
                };
                let row = probe(source, transport.as_ref(), &args.label, &args.fake);
                results.lock().expect("results mutex").push(row);
            });
        }
    });
    results.into_inner().expect("results mutex")
}

fn probe(
    source: &namecada_core::manifest::Source,
    transport: &dyn Transport,
    label: &str,
    control: &str,
) -> Row {
    let url = source.url_template.replace("%s", label);
    let real = fetch(transport, &url);
    let fake = fetch(transport, &source.url_template.replace("%s", control));
    let verdict = classify(&real, &fake, label, control, source.keyed);
    let note = if verdict == "differs" {
        let body = real.body.to_lowercase();
        if SOFT_404.iter().any(|marker| body.contains(marker)) {
            "soft-404 body".to_string()
        } else {
            String::new()
        }
    } else {
        String::new()
    };
    Row {
        id: source.id.clone(),
        platform_name: source.platform_name.clone(),
        verdict,
        real: real.status,
        fake: fake.status,
        url,
        title: title(&real.body),
        note,
    }
}

fn fetch(transport: &dyn Transport, url: &str) -> Fetched {
    let mut current = url.to_string();
    for _ in 0..=10 {
        let mut headers = BTreeMap::new();
        headers.insert("Accept".to_string(), ACCEPT.to_string());
        headers.insert("Accept-Language".to_string(), ACCEPT_LANGUAGE.to_string());
        headers.insert("User-Agent".to_string(), USER_AGENT.to_string());
        let request = HttpRequest {
            method: "GET".to_string(),
            url: current.clone(),
            headers,
            body: None,
        };
        let response = match transport.execute(request) {
            Ok(response) => response,
            Err(err) => {
                return Fetched {
                    status: 0,
                    url: current,
                    body: String::new(),
                    error: Some(err.to_string()),
                }
            }
        };
        if matches!(response.status, 301 | 302 | 303 | 307 | 308) {
            if let Some(location) = response.headers.get("location") {
                if let Ok(base) = Url::parse(&current) {
                    if let Ok(next) = base.join(location) {
                        current = next.to_string();
                        continue;
                    }
                }
            }
        }
        let body = &response.body[..response.body.len().min(BODY_CAP)];
        return Fetched {
            status: response.status,
            url: current,
            body: String::from_utf8_lossy(body).into_owned(),
            error: None,
        };
    }
    Fetched {
        status: 0,
        url: current,
        body: String::new(),
        error: Some("too many redirects".to_string()),
    }
}

fn title(body: &str) -> String {
    let lower = body.to_lowercase();
    let Some(start) = lower.find("<title") else {
        return String::new();
    };
    let Some(open) = body[start..].find('>') else {
        return String::new();
    };
    let from = start + open + 1;
    let Some(end) = lower[from..].find("</title>") else {
        return String::new();
    };
    body[from..from + end]
        .split_whitespace()
        .collect::<Vec<_>>()
        .join(" ")
        .chars()
        .take(70)
        .collect()
}

fn classify(real: &Fetched, fake: &Fetched, label: &str, control: &str, keyed: bool) -> String {
    if real.status == 0 || fake.status == 0 {
        return "dead".to_string();
    }
    let normalize = |url: &str| url.replace(label, "{name}").replace(control, "{name}");
    if real.status != fake.status || normalize(&real.url) != normalize(&fake.url) {
        return "differs".to_string();
    }
    let status = real.status;
    if status == 200 || status == 202 {
        return "shell".to_string();
    }
    let body = format!("{}{}", real.body, fake.body).to_lowercase();
    let bot = BOT_MARKERS.iter().any(|marker| body.contains(marker));
    if bot && matches!(status, 400 | 403 | 429 | 999) {
        return if keyed { "keyed" } else { "challenge" }.to_string();
    }
    if matches!(status, 401 | 402 | 403 | 429 | 999) {
        return if keyed { "keyed" } else { "blocked" }.to_string();
    }
    format!("same {status}")
}

fn rank(verdict: &str) -> u8 {
    match verdict {
        "differs" => 0,
        "dead" => 1,
        "keyed" => 2,
        "challenge" => 3,
        "blocked" => 4,
        "shell" => 5,
        _ => 4,
    }
}

fn history_dir() -> PathBuf {
    PathBuf::from(".probe")
}

fn save_run(results: &[Row], args: &Args) -> std::io::Result<PathBuf> {
    let dir = history_dir();
    std::fs::create_dir_all(&dir)?;
    let stamp = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map(|elapsed| elapsed.as_secs())
        .unwrap_or(0);
    let path = dir.join(format!("{stamp:012}.json"));
    let payload = serde_json::json!({
        "timestamp": stamp,
        "label": args.label,
        "fake": args.fake,
        "sources": results.iter().map(|row| serde_json::json!({
            "id": row.id,
            "verdict": row.verdict,
            "real": row.real,
            "fake": row.fake,
            "title": row.title,
        })).collect::<Vec<_>>(),
    });
    std::fs::write(&path, serde_json::to_string_pretty(&payload)? + "\n")?;
    Ok(path)
}

fn print_changes(results: &[Row], full: bool) {
    let dir = history_dir();
    let mut runs: Vec<PathBuf> = std::fs::read_dir(&dir)
        .map(|entries| {
            entries
                .filter_map(|entry| entry.ok())
                .map(|entry| entry.path())
                .filter(|path| path.extension().is_some_and(|ext| ext == "json"))
                .collect()
        })
        .unwrap_or_default();
    runs.sort();
    let Some(previous_path) = runs.last() else {
        eprintln!("no previous run saved; nothing to compare");
        return;
    };
    let previous: ProbeRun = match std::fs::read_to_string(previous_path)
        .ok()
        .and_then(|text| serde_json::from_str(&text).ok())
    {
        Some(previous) => previous,
        None => {
            eprintln!("probe: cannot read {}", previous_path.display());
            return;
        }
    };

    let before: BTreeMap<&str, &SavedRow> = previous
        .sources
        .iter()
        .map(|row| (row.id.as_str(), row))
        .collect();
    let after: BTreeMap<&str, &Row> = results.iter().map(|row| (row.id.as_str(), row)).collect();

    let mut changes = Vec::new();
    for (id, row) in &after {
        match before.get(id) {
            None => changes.push(format!("  {id:<21} new            {}", row.verdict)),
            Some(old) if old.verdict != row.verdict => {
                changes.push(format!("  {id:<21} {:<14} -> {}", old.verdict, row.verdict))
            }
            Some(_) => {}
        }
    }
    if full {
        for (id, old) in &before {
            if !after.contains_key(id) {
                changes.push(format!(
                    "  {id:<21} {:<14} -> no longer unchecked",
                    old.verdict
                ));
            }
        }
    }

    if changes.is_empty() {
        println!("no changes since {}", previous_path.display());
    } else {
        println!(
            "changed since {} ({:?}):",
            previous_path.display(),
            previous.timestamp
        );
        println!("{}", changes.join("\n"));
    }
}

#[derive(Deserialize)]
struct ProbeRun {
    #[serde(default)]
    timestamp: u64,
    #[serde(default)]
    sources: Vec<SavedRow>,
}

#[derive(Deserialize)]
struct SavedRow {
    id: String,
    verdict: String,
}

const HELP: &str = "\
Probe the sources that no provider checks yet.

Usage: probe [options]

  --label <name>     label that should exist (default: google)
  --fake <name>      control label that should not exist (default: zzqxnvc)
  --only <ids>       comma-separated source ids
  --jobs <n>         parallel probes (default: 4)
  --timeout <secs>   seconds per request (default: 20)
  --json             print JSON instead of a table
  --save             write the run to .probe/<timestamp>.json
  --diff             show what changed since the previous saved run
  -h, --help         print this help
";
