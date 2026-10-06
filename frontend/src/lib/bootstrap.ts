export const BOOTSTRAP_URL = 'https://data.iana.org/rdap/dns.json';

const KEY = 'domain-browser:bootstrap';
const TTL = 7 * 24 * 60 * 60 * 1000;

interface Cached {
	at: number;
	map: Record<string, string>;
}

const FALLBACK = new Map<string, string>([
	['com', 'https://rdap.verisign.com/com/v1/'],
	['ai', 'https://rdap.identitydigital.services/rdap/'],
	['io', 'https://rdap.identitydigital.services/rdap/']
]);

export function parseBootstrap(data: unknown): Map<string, string> {
	const map = new Map<string, string>();
	const services = (data as { services?: unknown })?.services;
	if (!Array.isArray(services)) return map;
	for (const service of services) {
		if (!Array.isArray(service) || service.length < 2) continue;
		const [tlds, urls] = service as [unknown, unknown];
		if (!Array.isArray(tlds) || !Array.isArray(urls) || urls.length === 0) continue;
		const base = `${String(urls[0]).replace(/\/+$/, '')}/`;
		for (const tld of tlds) {
			if (typeof tld === 'string') map.set(tld.toLowerCase(), base);
		}
	}
	return map;
}

function readCache(): Map<string, string> | null {
	if (typeof localStorage === 'undefined') return null;
	try {
		const raw = JSON.parse(localStorage.getItem(KEY) ?? 'null') as Cached | null;
		if (!raw || typeof raw.at !== 'number' || Date.now() - raw.at >= TTL) return null;
		return new Map(Object.entries(raw.map));
	} catch {
		return null;
	}
}

function writeCache(map: Map<string, string>) {
	try {
		const cached: Cached = { at: Date.now(), map: Object.fromEntries(map) };
		localStorage.setItem(KEY, JSON.stringify(cached));
	} catch {
		// storage unavailable or over quota: caching is best-effort
	}
}

function merge(base: Map<string, string>, extra: Map<string, string>): Map<string, string> {
	return new Map([...base, ...extra]);
}

async function load(): Promise<Map<string, string>> {
	const cached = readCache();
	if (cached) return merge(FALLBACK, cached);
	const res = await fetch(BOOTSTRAP_URL, { referrerPolicy: 'no-referrer', credentials: 'omit' });
	if (!res.ok) throw new Error(`bootstrap: HTTP ${res.status}`);
	const map = parseBootstrap(await res.json());
	writeCache(map);
	return merge(FALLBACK, map);
}

let inflight: Promise<Map<string, string>> | null = null;

export function getBootstrap(): Promise<Map<string, string>> {
	if (!inflight) {
		inflight = load().catch(() => FALLBACK);
	}
	return inflight;
}

export async function resolveBase(tld: string): Promise<string | null> {
	const map = await getBootstrap();
	return map.get(tld.toLowerCase()) ?? null;
}
