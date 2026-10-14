import { describe, expect, it } from 'vitest';
import { UNSUPPORTED } from './unsupported';

describe('UNSUPPORTED', () => {
	it('is not empty', () => {
		expect(UNSUPPORTED.length).toBeGreaterThan(0);
	});

	it('gives every entry a unique id and a display name', () => {
		const ids = UNSUPPORTED.map((source) => source.id);
		expect(new Set(ids).size).toBe(ids.length);
		expect(UNSUPPORTED.every((source) => source.platformName.length > 0)).toBe(true);
	});

	it('templates every url with the label placeholder', () => {
		expect(UNSUPPORTED.every((source) => source.url.includes('{name}'))).toBe(true);
	});

	it('classifies every entry as paid or free', () => {
		expect(UNSUPPORTED.every((source) => typeof source.paid === 'boolean')).toBe(true);
		expect(UNSUPPORTED.some((source) => source.paid)).toBe(true);
		expect(UNSUPPORTED.some((source) => !source.paid)).toBe(true);
	});
});
