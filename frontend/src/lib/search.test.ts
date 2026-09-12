import { describe, expect, it } from 'vitest';
import { lowerBound, parseDataset, search, type Dataset } from './search';

const namesRaw = 'aa\nab\nac\nba\nbb\nbc';
const availRaw = '101010';

function makeDataset(): Dataset {
	return parseDataset(namesRaw, availRaw);
}

describe('parseDataset', () => {
	it('splits names and decodes availability bits', () => {
		const ds = makeDataset();
		expect(ds.names).toEqual(['aa', 'ab', 'ac', 'ba', 'bb', 'bc']);
		expect(Array.from(ds.available)).toEqual([1, 0, 1, 0, 1, 0]);
		expect(ds.availableTotal).toBe(3);
	});
});

describe('lowerBound', () => {
	const names = ['aa', 'ab', 'ac'];
	it('finds the first index not less than x', () => {
		expect(lowerBound(names, 'aa')).toBe(0);
		expect(lowerBound(names, 'ab')).toBe(1);
		expect(lowerBound(names, 'ad')).toBe(3);
		expect(lowerBound(names, 'a')).toBe(0);
	});
});

describe('search', () => {
	const ds = makeDataset();

	it('returns nothing for an empty query', () => {
		expect(search(ds, { query: '', mode: 'prefix', onlyAvailable: false })).toEqual({
			total: 0,
			rows: [],
			truncated: false
		});
	});

	it('finds all prefix matches', () => {
		const r = search(ds, { query: 'a', mode: 'prefix', onlyAvailable: false });
		expect(r.total).toBe(3);
		expect(r.rows.map((x) => x.name)).toEqual(['aa', 'ab', 'ac']);
	});

	it('filters to available only', () => {
		const r = search(ds, { query: 'a', mode: 'prefix', onlyAvailable: true });
		expect(r.total).toBe(2);
		expect(r.rows.map((x) => x.name)).toEqual(['aa', 'ac']);
		expect(r.rows.every((x) => x.status === 'available')).toBe(true);
	});

	it('finds contains matches', () => {
		const r = search(ds, { query: 'b', mode: 'contains', onlyAvailable: false });
		expect(r.total).toBe(4);
		expect(r.rows.map((x) => x.name)).toEqual(['ab', 'ba', 'bb', 'bc']);
	});

	it('combines contains with the availability filter', () => {
		const r = search(ds, { query: 'b', mode: 'contains', onlyAvailable: true });
		expect(r.total).toBe(1);
		expect(r.rows.map((x) => x.name)).toEqual(['bb']);
	});

	it('is case-insensitive and trims', () => {
		const r = search(ds, { query: '  A  ', mode: 'prefix', onlyAvailable: false });
		expect(r.total).toBe(3);
	});

	it('caps rows but reports the true total', () => {
		const r = search(ds, { query: 'a', mode: 'prefix', onlyAvailable: false, limit: 2 });
		expect(r.total).toBe(3);
		expect(r.rows).toHaveLength(2);
		expect(r.truncated).toBe(true);
	});

	it('does not mark truncated when under the limit', () => {
		const r = search(ds, { query: 'a', mode: 'prefix', onlyAvailable: false, limit: 10 });
		expect(r.truncated).toBe(false);
	});
});
