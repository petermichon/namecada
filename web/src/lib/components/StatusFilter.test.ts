import { render } from 'svelte/server';
import { describe, expect, it } from 'vitest';
import StatusFilter from './StatusFilter.svelte';
import type { Filter } from '$lib/useLookup.svelte';

const TOTALS = { available: 0, taken: 0, unknown: 0, unsupported: 0 };

function chips(counts = TOTALS, filter: Filter = [], total = 0): string {
	return render(StatusFilter, { props: { counts, filter, total } }).body;
}

describe('StatusFilter', () => {
	it('keeps every chip visible when its count is zero', () => {
		const html = chips();

		for (const label of ['available', 'unavailable', 'unknown', 'not checked']) {
			expect(html).toContain(label);
		}
	});

	it('keeps a deselected chip visible so the filter can be turned back on', () => {
		const html = chips(TOTALS, ['available']);

		expect(html).toContain('unknown');
		expect(html).toContain('not checked');
		expect(html.match(/aria-pressed="false"/g)).toHaveLength(3);
	});

	it('shows the counts', () => {
		const html = chips({ available: 1, taken: 2, unknown: 3, unsupported: 4 });

		for (const label of ['available', 'unavailable', 'unknown', 'not checked']) {
			expect(html).toContain(label);
		}
		expect(html).toMatch(/available<\/span>\s*<span[^>]*>1</);
		expect(html).toMatch(/not checked<\/span>\s*<span[^>]*>4</);
	});
});
