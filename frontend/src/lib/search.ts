export type Status = 'available' | 'registered';
export type Mode = 'prefix' | 'contains';

export interface Row {
	name: string;
	status: Status;
}

export interface Dataset {
	names: string[];
	available: Uint8Array;
	availableTotal: number;
}

export interface SearchOptions {
	query: string;
	mode: Mode;
	onlyAvailable: boolean;
	limit?: number;
}

export interface SearchResult {
	total: number;
	rows: Row[];
	truncated: boolean;
}

export const DEFAULT_LIMIT = 500;

export function parseDataset(namesRaw: string, availRaw: string): Dataset {
	const names = namesRaw.split('\n');
	const available = new Uint8Array(availRaw.length);
	let availableTotal = 0;
	for (let i = 0; i < availRaw.length; i++) {
		const a = availRaw.charCodeAt(i) === 49 ? 1 : 0;
		available[i] = a;
		availableTotal += a;
	}
	return { names, available, availableTotal };
}

export function lowerBound(names: string[], x: string): number {
	let lo = 0;
	let hi = names.length;
	while (lo < hi) {
		const mid = (lo + hi) >> 1;
		if (names[mid] < x) lo = mid + 1;
		else hi = mid;
	}
	return lo;
}

export function search(ds: Dataset, opts: SearchOptions): SearchResult {
	const limit = opts.limit ?? DEFAULT_LIMIT;
	const q = opts.query.trim().toLowerCase();
	if (q.length === 0) return { total: 0, rows: [], truncated: false };

	const rows: Row[] = [];
	let total = 0;
	const push = (i: number) => {
		total++;
		if (rows.length < limit) {
			rows.push({ name: ds.names[i], status: ds.available[i] ? 'available' : 'registered' });
		}
	};

	if (opts.mode === 'prefix') {
		for (let i = lowerBound(ds.names, q); i < ds.names.length; i++) {
			if (!ds.names[i].startsWith(q)) break;
			if (opts.onlyAvailable && ds.available[i] === 0) continue;
			push(i);
		}
	} else {
		for (let i = 0; i < ds.names.length; i++) {
			if (!ds.names[i].includes(q)) continue;
			if (opts.onlyAvailable && ds.available[i] === 0) continue;
			push(i);
		}
	}

	return { total, rows, truncated: total > rows.length };
}
