import { describe, expect, it } from 'vitest';
import {
	allIndices,
	availableIndices,
	indexOfName,
	letterMatches,
	lowerBound,
	maskMatches,
	maskOf,
	matchingIndices,
	parseDataset,
	rowAt,
	sampleFromPool,
	type Dataset,
	type SearchOptions
} from './search';

const namesRaw = 'aa\nab\nac\nba\nbb\nbc';
const availRaw = '101010';

function makeDataset(): Dataset {
	return parseDataset(namesRaw, availRaw);
}

function pool(ds: Dataset, opts: SearchOptions) {
	const idx = matchingIndices(ds, opts);
	return {
		total: idx.length,
		names: Array.from(idx, (i) => ds.names[i]),
		rows: Array.from(idx, (i) => rowAt(ds, i))
	};
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

describe('matchingIndices', () => {
	const ds = makeDataset();

	it('returns nothing for an empty query', () => {
		expect(matchingIndices(ds, { query: '', mode: 'prefix', onlyAvailable: false })).toHaveLength(
			0
		);
	});

	it('finds all prefix matches', () => {
		expect(pool(ds, { query: 'a', mode: 'prefix', onlyAvailable: false }).names).toEqual([
			'aa',
			'ab',
			'ac'
		]);
	});

	it('filters to available only', () => {
		const r = pool(ds, { query: 'a', mode: 'prefix', onlyAvailable: true });
		expect(r.names).toEqual(['aa', 'ac']);
		expect(r.rows.every((x) => x.status === 'available')).toBe(true);
	});

	it('finds contains matches', () => {
		expect(pool(ds, { query: 'b', mode: 'contains', onlyAvailable: false }).names).toEqual([
			'ab',
			'ba',
			'bb',
			'bc'
		]);
	});

	it('finds postfix (ends-with) matches', () => {
		expect(pool(ds, { query: 'c', mode: 'postfix', onlyAvailable: false }).names).toEqual([
			'ac',
			'bc'
		]);
		expect(pool(ds, { query: 'c', mode: 'postfix', onlyAvailable: true }).names).toEqual(['ac']);
		expect(pool(ds, { query: 'a', mode: 'postfix', onlyAvailable: false }).names).toEqual([
			'aa',
			'ba'
		]);
	});

	it('combines contains with the availability filter', () => {
		expect(pool(ds, { query: 'b', mode: 'contains', onlyAvailable: true }).names).toEqual(['bb']);
	});

	it('is case-insensitive and trims', () => {
		expect(pool(ds, { query: '  A  ', mode: 'prefix', onlyAvailable: false }).total).toBe(3);
	});

	it('can filter by letters with no query', () => {
		expect(
			pool(ds, {
				query: '',
				mode: 'prefix',
				onlyAvailable: false,
				letters: ['a'],
				letterMode: 'include'
			}).names
		).toEqual(['aa', 'ab', 'ac', 'ba']);
	});

	it('filters by excluded letters', () => {
		expect(
			pool(ds, {
				query: '',
				mode: 'prefix',
				onlyAvailable: false,
				letters: ['a'],
				letterMode: 'exclude'
			}).names
		).toEqual(['bb', 'bc']);
	});
});

describe('rowAt', () => {
	it('builds a row with the correct status', () => {
		const ds = makeDataset();
		expect(rowAt(ds, 0)).toEqual({ name: 'aa', status: 'available' });
		expect(rowAt(ds, 1)).toEqual({ name: 'ab', status: 'registered' });
	});
});

describe('availableIndices', () => {
	it('collects the indices flagged available', () => {
		const ds = makeDataset();
		expect(Array.from(availableIndices(ds))).toEqual([0, 2, 4]);
	});
});

describe('allIndices', () => {
	it('returns every index in order', () => {
		const ds = makeDataset();
		expect(Array.from(allIndices(ds))).toEqual([0, 1, 2, 3, 4, 5]);
	});
});

describe('letterMatches', () => {
	it('requires all letters in include mode', () => {
		expect(letterMatches('abc', ['a', 'c'], 'include')).toBe(true);
		expect(letterMatches('abc', ['a', 'z'], 'include')).toBe(false);
	});

	it('forbids letters in exclude mode', () => {
		expect(letterMatches('abc', ['z'], 'exclude')).toBe(true);
		expect(letterMatches('abc', ['a'], 'exclude')).toBe(false);
	});

	it('passes when no letters are selected', () => {
		expect(letterMatches('abc', [], 'include')).toBe(true);
	});
});

describe('letter masks', () => {
	it('matches inclusion and exclusion', () => {
		const name = maskOf(['n', 's']);
		expect(maskMatches(name, maskOf(['n']), 'include')).toBe(true);
		expect(maskMatches(name, maskOf(['n', 'z']), 'include')).toBe(false);
		expect(maskMatches(name, maskOf(['z']), 'exclude')).toBe(true);
		expect(maskMatches(name, maskOf(['n']), 'exclude')).toBe(false);
		expect(maskMatches(name, 0, 'include')).toBe(true);
	});
});

describe('sampleFromPool', () => {
	function mulberry32(seed: number) {
		return function () {
			seed |= 0;
			seed = (seed + 0x6d2b79f5) | 0;
			let t = Math.imul(seed ^ (seed >>> 15), 1 | seed);
			t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
			return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
		};
	}

	const poolArr = new Uint32Array([10, 20, 30, 40, 50, 60]);

	it('returns the requested number of unique pool values', () => {
		const out = sampleFromPool(poolArr, 4, mulberry32(1));
		expect(out).toHaveLength(4);
		expect(new Set(out).size).toBe(4);
		expect(out.every((v) => Array.from(poolArr).includes(v))).toBe(true);
	});

	it('returns the whole pool when n exceeds its size', () => {
		const out = sampleFromPool(poolArr, 99, mulberry32(2));
		expect(new Set(out)).toEqual(new Set(Array.from(poolArr)));
	});

	it('handles n = 0 and negative n', () => {
		expect(sampleFromPool(poolArr, 0, mulberry32(3))).toEqual([]);
		expect(sampleFromPool(poolArr, -5, mulberry32(3))).toEqual([]);
	});
});

describe('indexOfName', () => {
	const ds = makeDataset();

	it('finds an existing name', () => {
		expect(indexOfName(ds, 'aa')).toBe(0);
		expect(indexOfName(ds, 'bc')).toBe(5);
	});

	it('returns -1 for an unknown name', () => {
		expect(indexOfName(ds, 'zz')).toBe(-1);
	});
});
