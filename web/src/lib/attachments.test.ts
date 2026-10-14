import { describe, expect, it } from 'vitest';
import { shouldAutofocus, type FocusEnv } from './attachments';

const base: FocusEnv = { connected: true, activeIsBody: true, pointerFine: true };

describe('shouldAutofocus', () => {
	it('focuses when connected, nothing else focused, and a fine pointer', () => {
		expect(shouldAutofocus(base)).toBe(true);
	});

	it('does not steal focus when another element is focused', () => {
		expect(shouldAutofocus({ ...base, activeIsBody: false })).toBe(false);
	});

	it('does not focus a detached node', () => {
		expect(shouldAutofocus({ ...base, connected: false })).toBe(false);
	});

	it('does not pop the keyboard on touch devices', () => {
		expect(shouldAutofocus({ ...base, pointerFine: false })).toBe(false);
	});
});
