import { describe, expect, it } from 'vitest';
import { loadRecent, RECENT_LIMIT, rememberLabel } from './recent';

function memoryStore(initial?: string) {
	const data = new Map<string, string>();
	if (initial !== undefined) data.set('namecada:recent', initial);
	return {
		getItem: (key: string) => data.get(key) ?? null,
		setItem: (key: string, value: string) => void data.set(key, value)
	};
}

describe('loadRecent', () => {
	it('returns nothing without a store', () => {
		expect(loadRecent(null)).toEqual([]);
	});

	it('returns nothing for missing, corrupt or non-array values', () => {
		expect(loadRecent(memoryStore())).toEqual([]);
		expect(loadRecent(memoryStore('{'))).toEqual([]);
		expect(loadRecent(memoryStore('{"label":"acema"}'))).toEqual([]);
	});

	it('drops blanks, repeats and non-strings, then caps the list', () => {
		const store = memoryStore(
			JSON.stringify(['acema', ' AceMa ', '', 7, 'beta', 'BETA', 'gamma', 'delta', 'epsilon'])
		);
		expect(loadRecent(store)).toEqual(['acema', 'beta', 'gamma', 'delta', 'epsilon']);
	});

	it('caps at the limit', () => {
		const labels = Array.from({ length: RECENT_LIMIT + 4 }, (_, index) => `name-${index}`);
		expect(loadRecent(memoryStore(JSON.stringify(labels)))).toHaveLength(RECENT_LIMIT);
	});
});

describe('rememberLabel', () => {
	it('puts the newest label first', () => {
		const store = memoryStore();
		rememberLabel('acema', store);
		expect(rememberLabel('beta', store)).toEqual(['beta', 'acema']);
	});

	it('moves a repeat to the front instead of duplicating it', () => {
		const store = memoryStore();
		rememberLabel('acema', store);
		rememberLabel('beta', store);
		expect(rememberLabel('  AceMa ', store)).toEqual(['AceMa', 'beta']);
	});

	it('caps the stored list', () => {
		const store = memoryStore();
		for (let index = 0; index < RECENT_LIMIT + 3; index += 1) {
			rememberLabel(`name-${index}`, store);
		}
		const stored = loadRecent(store);
		expect(stored).toHaveLength(RECENT_LIMIT);
		expect(stored[0]).toBe(`name-${RECENT_LIMIT + 2}`);
	});

	it('ignores a blank label', () => {
		const store = memoryStore();
		rememberLabel('acema', store);
		expect(rememberLabel('   ', store)).toEqual(['acema']);
	});

	it('survives a store that refuses writes', () => {
		const store = {
			getItem: () => null,
			setItem: () => {
				throw new Error('quota');
			}
		};
		expect(rememberLabel('acema', store)).toEqual(['acema']);
	});
});
