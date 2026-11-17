import { describe, expect, it } from 'vitest';
import { toLabel } from './rows';

describe('toLabel', () => {
	it('accepts a bare label', () => {
		expect(toLabel('  AceMa ')).toBe('acema');
		expect(toLabel('a-1-b')).toBe('a-1-b');
	});

	it('accepts dots and underscores inside a name', () => {
		expect(toLabel('acema.com')).toBe('acema.com');
		expect(toLabel('a_b.c-d')).toBe('a_b.c-d');
	});

	it('accepts any non-empty text and only rejects blank input', () => {
		expect(toLabel('a b')).toBe('a b');
		expect(toLabel('-abc')).toBe('-abc');
		expect(toLabel('@acema')).toBe('@acema');
		expect(toLabel('')).toBeNull();
		expect(toLabel('   ')).toBeNull();
	});
});
