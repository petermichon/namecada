import type { DomainStatus } from './rdap';

interface Entry {
	status: DomainStatus;
	at: number;
}

const KEY = 'domain-browser:cache';
const FRESH = 60 * 60 * 1000;

type Store = Record<string, Entry>;

export function isFresh(entry: Entry, now: number = Date.now(), ttl: number = FRESH): boolean {
	return now - entry.at < ttl;
}

function read(): Store {
	if (typeof localStorage === 'undefined') return {};
	try {
		const raw = JSON.parse(localStorage.getItem(KEY) ?? '{}');
		return raw && typeof raw === 'object' ? (raw as Store) : {};
	} catch {
		return {};
	}
}

function write(store: Store) {
	try {
		localStorage.setItem(KEY, JSON.stringify(store));
	} catch {
		// storage unavailable or over quota: caching is best-effort
	}
}

export function getCached(domain: string): Entry | null {
	return read()[domain] ?? null;
}

export function setCached(domain: string, status: DomainStatus) {
	const store = read();
	store[domain] = { status, at: Date.now() };
	write(store);
}

export function clearCache() {
	write({});
}
