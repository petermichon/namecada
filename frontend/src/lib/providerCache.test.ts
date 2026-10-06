import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cacheKey, clearProviderCache, getCached, isFresh, setCached } from './providerCache';

function fakeStorage(): Storage {
	const map = new Map<string, string>();
	return {
		getItem: (key: string) => map.get(key) ?? null,
		setItem: (key: string, value: string) => void map.set(key, value),
		removeItem: (key: string) => void map.delete(key),
		clear: () => map.clear(),
		key: (index: number) => [...map.keys()][index] ?? null,
		get length() {
			return map.size;
		}
	} as Storage;
}

describe('providerCache', () => {
	beforeEach(() => vi.stubGlobal('localStorage', fakeStorage()));
	afterEach(() => vi.unstubAllGlobals());

	it('keys entries by provider and label', () => {
		expect(cacheKey('npm', 'acema')).toBe('npm:acema');
	});

	it('stores and retrieves a result', () => {
		setCached('npm', 'acema', 'available');
		expect(getCached('npm', 'acema')).toMatchObject({ status: 'available' });
		expect(getCached('github', 'acema')).toBeNull();
	});

	it('keeps the detail alongside the status', () => {
		setCached('domain.com', 'acema', 'taken', 'registered 1997-09-15');
		expect(getCached('domain.com', 'acema')).toMatchObject({
			status: 'taken',
			detail: 'registered 1997-09-15'
		});
	});

	it('expires entries past the ttl', () => {
		expect(isFresh({ status: 'taken', at: 1_000 }, 1_500, 1_000)).toBe(true);
		expect(isFresh({ status: 'taken', at: 1_000 }, 2_500, 1_000)).toBe(false);
	});

	it('clears the store', () => {
		setCached('npm', 'acema', 'available');
		clearProviderCache();
		expect(getCached('npm', 'acema')).toBeNull();
	});

	it('degrades gracefully without localStorage', () => {
		vi.unstubAllGlobals();
		expect(getCached('npm', 'acema')).toBeNull();
		expect(() => setCached('npm', 'acema', 'taken')).not.toThrow();
	});
});
