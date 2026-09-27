#!/usr/bin/env python3
"""Resumable .com availability scanner using Verisign RDAP.

Usage:
  python3 domscan.py --letters aceimnorsuvwxz --workers 8 --out out.tsv
  python3 domscan.py --letters aceimnorsuvwxz --limit 2000 --out trial.tsv

The default letter set (aceimnorsuvwxz, 14 letters -> 14^5 = 537824 names)
is the one used to produce data/full.tsv.
"""
import argparse
import itertools
import os
import threading
import time
import urllib.error
import urllib.request
from concurrent.futures import ThreadPoolExecutor, wait, FIRST_COMPLETED

UA = "domscan/1.0 (+domain availability research)"


class RateLimiter:
    def __init__(self, rate):
        self.rate = rate
        self.lock = threading.Lock()
        self.next_at = 0.0

    def wait(self):
        with self.lock:
            now = time.monotonic()
            if self.next_at < now:
                self.next_at = now
            self.next_at += 1.0 / self.rate
            delay = self.next_at - now
        if delay > 0:
            time.sleep(delay)


class Backoff:
    def __init__(self):
        self.lock = threading.Lock()
        self.until = 0.0

    def pause(self, seconds):
        with self.lock:
            self.until = max(self.until, time.monotonic() + seconds)

    def wait(self):
        with self.lock:
            delay = self.until - time.monotonic()
        if delay > 0:
            time.sleep(delay)


def fetch(name, opener, timeout):
    url = f"https://rdap.verisign.com/com/v1/domain/{name}.com"
    req = urllib.request.Request(url, headers={"User-Agent": UA})
    try:
        with opener.open(req, timeout=timeout) as r:
            return r.status
    except urllib.error.HTTPError as e:
        return e.code
    except Exception:
        return None


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--letters", default="aceimnorsuvwxz")
    ap.add_argument("--out", default="scan.tsv")
    ap.add_argument("--workers", type=int, default=8)
    ap.add_argument("--rate", type=float, default=30.0, help="global max req/s")
    ap.add_argument("--limit", type=int, default=0, help="stop after N names (0=all)")
    ap.add_argument("--timeout", type=float, default=15.0)
    ap.add_argument("--retries", type=int, default=4)
    args = ap.parse_args()

    letters = list(args.letters)
    total = len(letters) ** 5

    done = set()
    if os.path.exists(args.out):
        with open(args.out) as f:
            for line in f:
                if line.strip():
                    done.add(line.split("\t", 1)[0])
    print(f"[scan] letters={''.join(letters)} total={total} already_done={len(done)}", flush=True)

    opener = urllib.request.build_opener()
    rl = RateLimiter(args.rate)
    bo = Backoff()

    lock = threading.Lock()
    out = open(args.out, "a", buffering=1)
    counts = {"reg": 0, "avail": 0, "err": 0}
    started = time.monotonic()
    processed = 0

    def work(name):
        code = None
        for attempt in range(args.retries):
            bo.wait()
            rl.wait()
            code = fetch(name, opener, args.timeout)
            if code in (200, 404):
                return name, code
            if code == 429:
                bo.pause(min(300, 15 * (2 ** attempt)))
                continue
            if code is None or code >= 500:
                time.sleep(1.5 * (attempt + 1))
                continue
            time.sleep(1.0 * (attempt + 1))
        return name, code

    names = ("".join(t) for t in itertools.product(letters, repeat=5) if "".join(t) not in done)
    if args.limit:
        names = itertools.islice(names, args.limit)

    ex = ThreadPoolExecutor(max_workers=args.workers)
    active = set()

    def drain_block(block=True):
        nonlocal processed
        ready, _ = wait(active, return_when=FIRST_COMPLETED, timeout=None if block else 0)
        for fut in ready:
            active.discard(fut)
            nm, code = fut.result()
            processed += 1
            if code == 200:
                status = "registered"; counts["reg"] += 1
            elif code == 404:
                status = "available"; counts["avail"] += 1
            else:
                status = f"error:{code}"; counts["err"] += 1
            with lock:
                out.write(f"{nm}\t{status}\n")
            if processed % 1000 == 0:
                el = time.monotonic() - started
                print(f"[scan] {processed} | avail={counts['avail']} reg={counts['reg']} err={counts['err']} | {processed/el:.1f}/s", flush=True)

    for name in names:
        active.add(ex.submit(work, name))
        if len(active) >= args.workers * 4:
            drain_block(True)
    while active:
        drain_block(True)

    ex.shutdown(wait=True)
    out.close()
    el = time.monotonic() - started
    print(f"[scan] DONE processed={processed} avail={counts['avail']} reg={counts['reg']} err={counts['err']} in {el:.1f}s", flush=True)


if __name__ == "__main__":
    main()
