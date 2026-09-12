export type Status = 'available' | 'registered';
export type Mode = 'prefix' | 'contains' | 'postfix';
export type LetterMode = 'include' | 'exclude';

export interface Row {
	name: string;
	status: Status;
}

export interface Dataset {
	names: string[];
	available: Uint8Array;
	availableTotal: number;
	letterMask: Uint32Array;
}

export interface SearchOptions {
	query: string;
	mode: Mode;
	onlyAvailable: boolean;
	letters?: string[];
	letterMode?: LetterMode;
}

const LETTER_BIT: Record<string, number> = {};
for (let i = 0; i < 26; i++) LETTER_BIT[String.fromCharCode(97 + i)] = 1 << i;

export function maskOf(letters: string[]): number {
	let mask = 0;
	for (const l of letters) mask |= LETTER_BIT[l] ?? 0;
	return mask;
}

export function maskMatches(nameMask: number, required: number, mode: LetterMode): boolean {
	if (required === 0) return true;
	return mode === 'exclude' ? (nameMask & required) === 0 : (nameMask & required) === required;
}

export function parseDataset(namesRaw: string, availRaw: string): Dataset {
	const names = namesRaw.split('\n');
	const available = new Uint8Array(availRaw.length);
	const letterMask = new Uint32Array(names.length);
	let availableTotal = 0;
	for (let i = 0; i < availRaw.length; i++) {
		const a = availRaw.charCodeAt(i) === 49 ? 1 : 0;
		available[i] = a;
		availableTotal += a;
	}
	for (let i = 0; i < names.length; i++) {
		const name = names[i];
		let mask = 0;
		for (let j = 0; j < name.length; j++) mask |= LETTER_BIT[name[j]] ?? 0;
		letterMask[i] = mask;
	}
	return { names, available, availableTotal, letterMask };
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

export function letterMatches(name: string, letters: string[], mode: LetterMode): boolean {
	if (letters.length === 0) return true;
	if (mode === 'exclude') return !letters.some((l) => name.includes(l));
	return letters.every((l) => name.includes(l));
}

export function rowAt(ds: Dataset, i: number): Row {
	return { name: ds.names[i], status: ds.available[i] ? 'available' : 'registered' };
}

export function availableIndices(ds: Dataset): Uint32Array {
	const out = new Uint32Array(ds.availableTotal);
	let k = 0;
	for (let i = 0; i < ds.available.length; i++) {
		if (ds.available[i]) out[k++] = i;
	}
	return out;
}

export function allIndices(ds: Dataset): Uint32Array {
	const out = new Uint32Array(ds.names.length);
	for (let i = 0; i < out.length; i++) out[i] = i;
	return out;
}

export function matchingIndices(ds: Dataset, opts: SearchOptions): Uint32Array {
	const q = opts.query.trim().toLowerCase();
	const required = maskOf(opts.letters ?? []);
	const letterMode = opts.letterMode ?? 'include';
	if (q.length === 0 && required === 0) return new Uint32Array(0);

	const pass = (i: number) => {
		if (opts.onlyAvailable && ds.available[i] === 0) return false;
		return maskMatches(ds.letterMask[i], required, letterMode);
	};

	if (q.length === 0) {
		const out: number[] = [];
		for (let i = 0; i < ds.names.length; i++) if (pass(i)) out.push(i);
		return Uint32Array.from(out);
	}
	if (opts.mode === 'prefix') {
		const out: number[] = [];
		for (let i = lowerBound(ds.names, q); i < ds.names.length; i++) {
			if (!ds.names[i].startsWith(q)) break;
			if (pass(i)) out.push(i);
		}
		return Uint32Array.from(out);
	}
	if (opts.mode === 'postfix') {
		const out: number[] = [];
		for (let i = 0; i < ds.names.length; i++) {
			if (!ds.names[i].endsWith(q)) continue;
			if (pass(i)) out.push(i);
		}
		return Uint32Array.from(out);
	}
	const out: number[] = [];
	for (let i = 0; i < ds.names.length; i++) {
		if (!ds.names[i].includes(q)) continue;
		if (pass(i)) out.push(i);
	}
	return Uint32Array.from(out);
}

export function sampleFromPool(
	pool: ArrayLike<number>,
	n: number,
	rng: () => number = Math.random
): number[] {
	const size = Math.min(Math.max(n, 0), pool.length);
	const arr = Array.from({ length: pool.length }, (_, i) => pool[i]);
	for (let i = 0; i < size; i++) {
		const j = i + Math.floor(rng() * (arr.length - i));
		const tmp = arr[i];
		arr[i] = arr[j];
		arr[j] = tmp;
	}
	return arr.slice(0, size);
}

export function indexOfName(ds: Dataset, name: string): number {
	const i = lowerBound(ds.names, name);
	return i < ds.names.length && ds.names[i] === name ? i : -1;
}
