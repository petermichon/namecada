#!/usr/bin/env python3
"""Probe the sources that no provider checks yet ("not checked" rows).

Run it on demand when you want to know whether any of them has become
checkable: it fetches each unchecked source with a real label and a control
label, then sorts the results into differs / blocked / shell / dead.
Promoting a source is still a manual change (spec table, fixture, docs).

--save appends the run to .probe/<timestamp>.json and --diff prints what
changed since the previous saved run, so drift is visible without a scheduler.

    npm run probe
    npm run probe -- --label acema
    npm run probe -- --only opensea,figma --json

Exit status is 0 for a completed report, 1 when the registry cannot be read.
"""

from __future__ import annotations

import argparse
import datetime
import http.cookiejar
import json
import re
import sys
import urllib.error
import urllib.request
from concurrent.futures import ThreadPoolExecutor
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
HISTORY = ROOT / ".probe"
UA = (
    "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 "
    "(KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36"
)
BOT_MARKERS = (
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
)
SOFT_404 = ("page not found", "not found", "doesn't exist", "no longer available")


def provider_ids() -> set[str]:
    """Read the compiled provider registry: status, custom and DNS tables."""
    status = (ROOT / "rust/core/src/status.rs").read_text()
    ids = set(re.findall(r'spec(?:_with)?\(\s*"([A-Za-z0-9._-]+)"', status))

    custom = (ROOT / "rust/core/src/custom.rs").read_text()
    specs = custom[custom.index("static CUSTOM_SPECS") :]
    ids |= set(re.findall(r'id:\s*"([A-Za-z0-9._-]+)"', specs))

    dns = (ROOT / "rust/core/src/dns.rs").read_text()
    hosts = dns[dns.index("pub fn dns_hosts") :]
    ids |= set(re.findall(r'\("([A-Za-z0-9._-]+)",\s*"[^"]+"\)', hosts))
    return ids


def load_unchecked(only: set[str] | None) -> list[dict]:
    manifest = json.loads((ROOT / "fixtures/sources.json").read_text())
    checked = provider_ids()
    rows = [s for s in manifest["sources"] if s["id"] not in checked]
    if only:
        rows = [s for s in rows if s["id"] in only]
    return rows


def fetch(url: str, timeout: float) -> dict:
    jar = http.cookiejar.CookieJar()
    opener = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(jar))
    request = urllib.request.Request(
        url,
        headers={
            "User-Agent": UA,
            "Accept": "text/html,application/xhtml+xml,application/json;q=0.9,*/*;q=0.8",
            "Accept-Language": "en-US,en;q=0.9",
        },
    )
    try:
        with opener.open(request, timeout=timeout) as response:
            body = response.read(120_000).decode("utf-8", "replace")
            return {
                "status": response.status,
                "url": response.geturl(),
                "body": body,
            }
    except urllib.error.HTTPError as err:
        body = err.read(120_000).decode("utf-8", "replace")
        return {"status": err.code, "url": err.geturl() or url, "body": body}
    except Exception as err:  # timeouts, TLS, DNS, connection resets
        return {"status": 0, "url": url, "body": "", "error": str(err)}


def title(body: str) -> str:
    match = re.search(r"<title[^>]*>(.*?)</title>", body, re.S | re.I)
    return re.sub(r"\s+", " ", match.group(1)).strip()[:70] if match else ""


def classify(real: dict, fake: dict, label: str, control: str, keyed: bool) -> str:
    if real["status"] == 0 or fake["status"] == 0:
        return "dead"
    normalized = lambda url: url.replace(label, "{name}").replace(control, "{name}")
    if real["status"] != fake["status"] or normalized(real["url"]) != normalized(fake["url"]):
        return "differs"
    status = real["status"]
    if status in (200, 202):
        return "shell"
    body = (real["body"] + fake["body"]).lower()
    bots = any(marker in body for marker in BOT_MARKERS)
    if bots and status in (400, 403, 429, 999):
        return "keyed" if keyed else "challenge"
    if status in (401, 402, 403, 429, 999):
        return "keyed" if keyed else "blocked"
    return f"same {status}"


def probe(source: dict, label: str, control: str, timeout: float) -> dict:
    url = source["urlTemplate"].replace("%s", label)
    real = fetch(url, timeout)
    fake = fetch(source["urlTemplate"].replace("%s", control), timeout)
    verdict = classify(real, fake, label, control, source.get("keyed", False))
    note = ""
    if verdict == "differs":
        body = real["body"].lower()
        note = "soft-404 body" if any(m in body for m in SOFT_404) else ""
    return {
        "id": source["id"],
        "platformName": source["platformName"],
        "verdict": verdict,
        "real": real["status"],
        "fake": fake["status"],
        "url": url,
        "title": title(real["body"]),
        "note": note,
    }


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--label", default="google", help="label that should exist (default: google)")
    parser.add_argument("--fake", default="zzqxnvc", help="control label that should not exist")
    parser.add_argument("--only", help="comma-separated source ids")
    parser.add_argument("--jobs", type=int, default=4, help="parallel probes (default: 4)")
    parser.add_argument("--timeout", type=float, default=20.0, help="seconds per request (default: 20)")
    parser.add_argument("--json", action="store_true", help="print JSON instead of a table")
    parser.add_argument("--save", action="store_true", help=f"write the run to {HISTORY.name}/<timestamp>.json")
    parser.add_argument("--diff", action="store_true", help="show what changed since the previous saved run")
    args = parser.parse_args()

    only = {s.strip() for s in args.only.split(",")} if args.only else None
    try:
        rows = load_unchecked(only)
    except Exception as err:
        print(f"probe: cannot read the registry: {err}", file=sys.stderr)
        return 1

    if not rows:
        scope = f" matching {', '.join(sorted(only))}" if only else ""
        print(f"probe: no unchecked sources{scope}", file=sys.stderr)
        return 1

    with ThreadPoolExecutor(max_workers=max(1, args.jobs)) as pool:
        results = list(
            pool.map(lambda row: probe(row, args.label, args.fake, args.timeout), rows)
        )

    if args.diff:
        print_changes(results, full=only is None)

    if args.save:
        saved = save_run(results, args.label, args.fake)
        print(f"saved {saved}", file=sys.stderr)

    if args.json:
        print(json.dumps(results, indent=2))
        return 0

    order = {"differs": 0, "dead": 1, "keyed": 2, "challenge": 3, "blocked": 4, "shell": 5}
    results.sort(key=lambda r: (order.get(r["verdict"], 4), r["id"]))
    for row in results:
        note = f"  ({row['note']})" if row["note"] else ""
        print(
            f"{row['verdict']:<11} {row['id']:<21} {row['real']:>3}/{row['fake']:<3} "
            f"{row['title']}{note}"
        )

    counts = {}
    for row in results:
        key = row["verdict"].split()[0]
        counts[key] = counts.get(key, 0) + 1
    summary = ", ".join(f"{n} {k}" for k, n in sorted(counts.items()))
    print(f"\n{len(results)} unchecked sources probed ({summary})")
    if counts.get("differs"):
        print("differs = candidate: confirm the mapping, then add a spec, a fixture and docs")
    return 0



def latest_run() -> Path | None:
    if not HISTORY.is_dir():
        return None
    runs = sorted(HISTORY.glob("*.json"))
    return runs[-1] if runs else None


def save_run(results: list[dict], label: str, fake: str) -> Path:
    HISTORY.mkdir(exist_ok=True)
    stamp = datetime.datetime.now(datetime.timezone.utc).strftime("%Y-%m-%dT%H%M%SZ")
    path = HISTORY / f"{stamp}.json"
    payload = {
        "timestamp": stamp,
        "label": label,
        "fake": fake,
        "sources": [
            {key: row[key] for key in ("id", "verdict", "real", "fake", "title")}
            for row in results
        ],
    }
    path.write_text(json.dumps(payload, indent=1) + "\n")
    return path


def print_changes(results: list[dict], full: bool) -> None:
    previous_path = latest_run()
    if previous_path is None:
        print("no previous run saved; nothing to compare", file=sys.stderr)
        return
    try:
        previous = json.loads(previous_path.read_text())
    except (OSError, ValueError) as err:
        print(f"probe: cannot read {previous_path}: {err}", file=sys.stderr)
        return

    before = {row["id"]: row for row in previous.get("sources", [])}
    after = {row["id"]: row for row in results}
    changes = []
    for source_id, row in sorted(after.items()):
        old = before.get(source_id)
        if old is None:
            changes.append(f"  {source_id:<21} new            {row['verdict']}")
        elif old["verdict"] != row["verdict"]:
            changes.append(f"  {source_id:<21} {old['verdict']:<14} -> {row['verdict']}")
    if full:
        for source_id, row in sorted(before.items()):
            if source_id not in after:
                changes.append(f"  {source_id:<21} {row['verdict']:<14} -> no longer unchecked")

    if changes:
        print(f"changed since {previous_path.name} ({previous.get('timestamp', '?')}):")
        print("\n".join(changes))
    else:
        print(f"no changes since {previous_path.name}")

if __name__ == "__main__":
    sys.exit(main())
