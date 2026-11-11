export type Status = 'available' | 'taken' | 'unknown';

export type CostGroup = 'free' | 'paid';
export type CostFilter = CostGroup[];

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
