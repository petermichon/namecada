import { describe, expect, it } from 'vitest';
import { suggestLabel, toLabel } from './rows';

describe('toLabel', () => {
	it('accepts a bare label', () => {
		expect(toLabel('  AceMa ')).toBe('acema');
		expect(toLabel('a-1-b')).toBe('a-1-b');
	});

	it('accepts dots and underscores inside a name', () => {
		expect(toLabel('acema.com')).toBe('acema.com');
		expect(toLabel('a_b.c-d')).toBe('a_b.c-d');
	});

	it('rejects empty and malformed input', () => {
		expect(toLabel('')).toBeNull();
		expect(toLabel('   ')).toBeNull();
		expect(toLabel('a b')).toBeNull();
		expect(toLabel('-abc')).toBeNull();
		expect(toLabel('abc-')).toBeNull();
		expect(toLabel('.abc')).toBeNull();
		expect(toLabel('abc_')).toBeNull();
	});
});

describe('suggestLabel', () => {
	it('extracts a name from a handle or URL', () => {
		expect(suggestLabel('@acema')).toBe('acema');
		expect(suggestLabel('https://acema.com/path')).toBe('acema.com');
		expect(suggestLabel('www.acema.com')).toBe('acema.com');
	});

	it('returns nothing when the input is already a label or cannot be reduced safely', () => {
		expect(suggestLabel('acema')).toBeNull();
		expect(suggestLabel('acema.com')).toBeNull();
		expect(suggestLabel('')).toBeNull();
		expect(suggestLabel('.com')).toBeNull();
		expect(suggestLabel('a b')).toBeNull();
		expect(suggestLabel('-acema.com')).toBeNull();
	});
});
