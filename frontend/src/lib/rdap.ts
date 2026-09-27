import { resolveBase } from './bootstrap';

export type DomainStatus = 'registered' | 'unregistered' | 'unknown';

export interface LookupResult {
	domain: string;
	status: DomainStatus;
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

export function normalizeDomain(input: string): string | null {
	const value = input
		.trim()
		.toLowerCase()
		.replace(/^\.+|\.+$/g, '');
	if (!value || !value.includes('.')) return null;
	if (value.length > MAX_DOMAIN) return null;
	const labels = value.split('.');
	if (labels.length < 2) return null;
	if (!labels.every((label) => LABEL.test(label))) return null;
	return value;
}

export function statusFromHttp(code: number): DomainStatus {
	if (code === 200) return 'registered';
	if (code === 404) return 'unregistered';
	return 'unknown';
}

export function domainUrl(base: string, domain: string): string {
	return `${base}domain/${domain}`;
}

export async function lookupDomain(domain: string, signal?: AbortSignal): Promise<LookupResult> {
	const tld = tldOf(domain);
	const base = await resolveBase(tld);
	if (!base) throw new UnsupportedTldError(tld);
	const res = await fetch(domainUrl(base, domain), { signal });
	res.body?.cancel().catch(() => {});
	return { domain, status: statusFromHttp(res.status) };
}
