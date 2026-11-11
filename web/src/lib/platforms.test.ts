import { describe, expect, it } from 'vitest';
import { suggestLabel, toLabel } from './platforms';

describe('toLabel', () => {
	it('accepts a bare label', () => {
		expect(toLabel('  AceMa ')).toBe('acema');
		expect(toLabel('a-1-b')).toBe('a-1-b');
	});

	it('rejects a dotted input rather than reducing it to a label', () => {
		expect(toLabel('acema.com')).toBeNull();
		expect(toLabel('www.bbc.co.uk')).toBeNull();
	});

	it('rejects empty and malformed input', () => {
		expect(toLabel('')).toBeNull();
		expect(toLabel('   ')).toBeNull();
		expect(toLabel('a b')).toBeNull();
		expect(toLabel('-abc')).toBeNull();
		expect(toLabel('abc-')).toBeNull();
	});
});

describe('suggestLabel', () => {
	it('extracts the bare label from a domain, handle, or URL', () => {
		expect(suggestLabel('acema.com')).toBe('acema');
		expect(suggestLabel('Acema.COM')).toBe('acema');
		expect(suggestLabel('@acema')).toBe('acema');
		expect(suggestLabel('www.acema.com')).toBe('acema');
		expect(suggestLabel('https://acema.com/path')).toBe('acema');
	});

	it('returns nothing when the input is already a label or cannot be reduced safely', () => {
		expect(suggestLabel('acema')).toBeNull();
		expect(suggestLabel('')).toBeNull();
		expect(suggestLabel('.com')).toBeNull();
		expect(suggestLabel('a b')).toBeNull();
		expect(suggestLabel('-acema.com')).toBeNull();
	});
});
