export type DomainStatus = 'registered' | 'unregistered' | 'unknown';

export interface LookupResult {
	domain: string;
	status: DomainStatus;
}

const RDAP_BASE = 'https://rdap.verisign.com/com/v1/domain/';
const LABEL = /^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/;

export function normalizeDomain(input: string): string | null {
	let label = input.trim().toLowerCase();
	if (label.endsWith('.com')) label = label.slice(0, -4);
	label = label.replace(/^\.+|\.+$/g, '');
	if (!LABEL.test(label)) return null;
	return `${label}.com`;
}

export function statusFromHttp(code: number): DomainStatus {
	if (code === 200) return 'registered';
	if (code === 404) return 'unregistered';
	return 'unknown';
}

export function rdapUrl(domain: string): string {
	return `${RDAP_BASE}${domain}`;
}

export async function lookupDomain(domain: string, signal?: AbortSignal): Promise<LookupResult> {
	const res = await fetch(rdapUrl(domain), { signal });
	return { domain, status: statusFromHttp(res.status) };
}
