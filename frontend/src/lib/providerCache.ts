import type { Status } from './providers';

export interface CachedResult {
	status: Status;
	detail?: string;
	at: number;
}

const KEY = 'domain-browser:provider-cache:v1';
export const FRESH = 60 * 60 * 1000;

type Store = Record<string, CachedResult>;

export function cacheKey(providerId: string, label: string): string {
	return `${providerId}:${label}`;
}

export function isFresh(
	entry: CachedResult,
	now: number = Date.now(),
	ttl: number = FRESH
): boolean {
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

export function getCached(providerId: string, label: string): CachedResult | null {
	return read()[cacheKey(providerId, label)] ?? null;
}

export function setCached(providerId: string, label: string, status: Status, detail?: string) {
	const store = read();
	store[cacheKey(providerId, label)] = { status, detail, at: Date.now() };
	write(store);
}

export function clearProviderCache() {
	write({});
}
