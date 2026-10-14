import { describe, expect, it } from 'vitest';
import { toLabel } from './platforms';

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
