export type Status = 'available' | 'taken' | 'unknown';

// Facet tags live with each source in fixtures/sources.json. The order here is
// the order the report's category chips render in.
export const CATEGORIES = [
	{ id: 'domains', label: 'Domains' },
	{ id: 'packages', label: 'Packages' },
	{ id: 'social', label: 'Social' },
	{ id: 'code', label: 'Code' },
	{ id: 'creator', label: 'Creator' },
	{ id: 'gaming', label: 'Gaming' },
	{ id: 'hosting', label: 'Hosting' },
	{ id: 'commerce', label: 'Commerce' },
	{ id: 'company', label: 'Company' },
	{ id: 'web3', label: 'Web3' }
] as const;

export type CategoryId = (typeof CATEGORIES)[number]['id'];
export type CategoryFilter = 'all' | CategoryId;
export type CostFilter = 'all' | 'free' | 'paid';

export interface ProviderResult {
	id: string;
	platform: string;
	platformName: string;
	name: string;
	url: string;
	status: Status;
	detail?: string;
	paid: boolean;
}

const LABEL = /^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/;

export function toLabel(input: string): string | null {
	const value = input.trim().toLowerCase();
	return LABEL.test(value) ? value : null;
}

// suggestLabel extracts a bare label from a rejected value so the UI can offer
// a one-click correction. It only removes unambiguous wrappers (a protocol, a
// leading www. or @, a domain suffix or path); anything else stays rejected.
export function suggestLabel(input: string): string | null {
	const original = input.trim().toLowerCase();
	if (!original) return null;

	let value = original.replace(/^https?:\/\//, '').replace(/^www\./, '');
	if (value.startsWith('@')) value = value.slice(1);

	const [first] = value.split(/[./?#]/);
	const candidate = first.trim();
	if (!candidate || candidate === original) return null;
	return toLabel(candidate) ? candidate : null;
}
