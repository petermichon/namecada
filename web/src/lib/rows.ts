export type Status = 'available' | 'taken' | 'unknown' | 'invalid';

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

// toLabel normalises what the user typed. Whether a name can exist on a given
// source is the engine's job; the input itself accepts anything.
export function toLabel(input: string): string | null {
	const value = input.trim().toLowerCase();
	return value || null;
}

export interface Row {
	id: string;
	platform: string;
	platformName: string;
	name: string;
	url: string;
	status: Status | 'pending' | 'unsupported';
	detail?: string;
	paid: boolean;
}

export type FilterStatus = Status | 'unsupported';

// A filter is the set of statuses to keep. An empty set means no selection,
// which shows every row (the default), so `available + not checked` unions
// those two and hides the rest.
export type Filter = FilterStatus[];

export type Counts = {
	available: number;
	taken: number;
	unknown: number;
	invalid: number;
	unsupported: number;
};

// validate turns raw input into a label. Any non-empty text is allowed: the
// engine reports per-source invalidity in the results.
export function validate(input: string): { label: string | null; error: string } {
	const label = toLabel(input);
	return { label, error: label ? '' : 'Enter a name to check.' };
}

// The default view is the verdicts; deferred/keyed sources are opt-in.
export const DEFAULT_FILTER: Filter = ['available', 'taken', 'unknown'];
export const ALL_STATUSES: Filter = [...DEFAULT_FILTER, 'invalid', 'unsupported'];
export const DEFAULT_COST: CostFilter = ['free', 'paid'];

// Order rows alphabetically by their display name, independent of status, so
// the list stays stable as results stream in. Ties break on the provider URL so
// the order is deterministic regardless of arrival order — domains sort by TLD
// (.ai .app .com .dev .io .org) and Hex sorts packages/ before users/.
export function byPlatformName(a: Row, b: Row): number {
	if (a.platformName < b.platformName) return -1;
	if (a.platformName > b.platformName) return 1;
	if (a.url < b.url) return -1;
	if (a.url > b.url) return 1;
	return 0;
}

export function matchesStatus(row: Row, filter: Filter): boolean {
	if (filter.length === 0) return false;
	if (filter.includes(row.status as FilterStatus)) return true;
	// Pending rows stay visible while a verdict is selected, but not when the
	// only thing asked for is the not-checked rows.
	return row.status === 'pending' && filter.some((status) => status !== 'unsupported');
}

export function matchesCost(row: Row, cost: CostFilter): boolean {
	return cost.includes(row.paid ? 'paid' : 'free');
}

// tally counts the verdicts; pending rows are not a verdict.
export function tally(rows: Row[]): Counts {
	const counts: Counts = { available: 0, taken: 0, unknown: 0, invalid: 0, unsupported: 0 };
	for (const row of rows) {
		if (row.status !== 'pending') counts[row.status]++;
	}
	return counts;
}

export function settlePending(rows: Row[]): Row[] {
	return rows.map((row) =>
		row.status === 'pending' ? { ...row, status: 'unknown' as const, detail: 'unreachable' } : row
	);
}

// visible is what the list shows: both facets applied, in stable order.
export function visible(rows: Row[], filter: Filter, cost: CostFilter): Row[] {
	if (filter.length === 0) return [];
	return rows
		.filter((row) => matchesStatus(row, filter) && matchesCost(row, cost))
		.toSorted(byPlatformName);
}

// costTally honours the status filter, so a cost chip previews its own count.
export function costTally(rows: Row[], filter: Filter): { free: number; paid: number } {
	const counts = { free: 0, paid: 0 };
	for (const row of rows) {
		if (matchesStatus(row, filter)) counts[row.paid ? 'paid' : 'free']++;
	}
	return counts;
}

// announce is the screen-reader summary of the current selection.
export function announce(visible: Row[], filter: Filter, searching: boolean): string {
	if (searching) return '';
	if (filter.length === 0) return 'No statuses selected.';
	const counts = tally(visible);
	const summary = `${counts.available} available, ${counts.taken} unavailable, ${counts.unknown} unknown`;
	const invalid = counts.invalid ? `, ${counts.invalid} invalid` : '';
	const notChecked = counts.unsupported ? `, ${counts.unsupported} not checked` : '';
	return `${visible.length} shown. ${summary}${invalid}${notChecked}`;
}
