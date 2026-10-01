import { parse as parseTld } from 'tldts';
import { resolveBase } from './bootstrap';

export type DomainStatus = 'registered' | 'unregistered' | 'unknown';

export interface LookupResult {
	domain: string;
	status: DomainStatus;
	registeredOn?: string;
	raw?: unknown;
}

const LABEL = /^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/;
const MAX_DOMAIN = 253;

export class UnsupportedTldError extends Error {
	constructor(public readonly tld: string) {
		super(`No RDAP endpoint for .${tld}`);
		this.name = 'UnsupportedTldError';
	}
}

export function tldOf(domain: string): string {
	return domain.slice(domain.lastIndexOf('.') + 1);
}

export interface ParsedDomain {
	domain: string;
	subdomain: string;
}

export function parseDomain(input: string): ParsedDomain | null {
	const value = input
		.trim()
		.toLowerCase()
		.replace(/^\.+|\.+$/g, '');
	if (!value || !value.includes('.')) return null;
	if (value.length > MAX_DOMAIN) return null;
	const labels = value.split('.');
	if (!labels.every((label) => LABEL.test(label))) return null;
	const parsed = parseTld(value, { allowPrivateDomains: false });
	if (!parsed.domain) return null;
	return { domain: parsed.domain, subdomain: parsed.subdomain ?? '' };
}

export function normalizeDomain(input: string): string | null {
	return parseDomain(input)?.domain ?? null;
}

export function statusFromHttp(code: number): DomainStatus {
	if (code === 200) return 'registered';
	if (code === 404) return 'unregistered';
	return 'unknown';
}

export function domainUrl(base: string, domain: string): string {
	return `${base}domain/${domain}`;
}

export function registrationDate(data: unknown): string | null {
	const events = (data as { events?: unknown })?.events;
	if (!Array.isArray(events)) return null;
	for (const entry of events) {
		const event = entry as { eventAction?: unknown; eventDate?: unknown };
		if (
			typeof event?.eventAction === 'string' &&
			event.eventAction.toLowerCase() === 'registration' &&
			typeof event.eventDate === 'string'
		) {
			return event.eventDate;
		}
	}
	return null;
}

export async function lookupDomain(domain: string, signal?: AbortSignal): Promise<LookupResult> {
	const tld = tldOf(domain);
	const base = await resolveBase(tld);
	if (!base) throw new UnsupportedTldError(tld);
	const res = await fetch(domainUrl(base, domain), {
		signal,
		referrerPolicy: 'no-referrer',
		credentials: 'omit'
	});
	if (res.status === 200) {
		const data = await res.json().catch(() => undefined);
		const registeredOn = registrationDate(data) ?? undefined;
		return { domain, status: 'registered', registeredOn, raw: data };
	}
	res.body?.cancel().catch(() => {});
	return { domain, status: statusFromHttp(res.status) };
}
