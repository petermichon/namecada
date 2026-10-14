import { describe, expect, it } from 'vitest';
import { UNSUPPORTED, type UnsupportedReason } from './unsupported';

const REASONS: UnsupportedReason[] = ['needs key', 'bot-blocked', 'no public API', 'not checked'];

describe('UNSUPPORTED', () => {
	it('is not empty', () => {
		expect(UNSUPPORTED.length).toBeGreaterThan(0);
	});

	it('gives every entry a unique id and a display name', () => {
		const ids = UNSUPPORTED.map((source) => source.id);
		expect(new Set(ids).size).toBe(ids.length);
		expect(UNSUPPORTED.every((source) => source.platformName.length > 0)).toBe(true);
	});

	it('uses only the documented reason tags', () => {
		expect(UNSUPPORTED.every((source) => REASONS.includes(source.reason))).toBe(true);
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
