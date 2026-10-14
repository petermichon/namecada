import { render } from 'svelte/server';
import { describe, expect, it } from 'vitest';
import ResultsList from './ResultsList.svelte';
import type { Row } from '$lib/useLookup.svelte';

function row(id: string, paid = true): Row {
	return {
		id,
		platform: 'github',
		platformName: 'GitHub',
		name: id,
		url: `https://example.com/${id}`,
		status: 'available',
		paid
	};
}

function firstUlClass(html: string): string {
	return html.match(/<ul[^>]*class="([^"]*)"/)?.[1] ?? '';
}

function liClasses(html: string): string[] {
	return [...html.matchAll(/<li[^>]*class="([^"]*)"/g)].map((match) => match[1]);
}

describe('ResultsList', () => {
	it('paints the grid with the row background so a trailing empty cell matches', () => {
		const { body } = render(ResultsList, { props: { paidRows: [row('acema')] } });
		const ul = firstUlClass(body);

		expect(ul).toContain('bg-card');
		expect(ul).not.toContain('bg-border');
		expect(ul).not.toContain('gap-px');
	});

	it('draws the 1px separators as borders on the rows', () => {
		const { body } = render(ResultsList, { props: { paidRows: [row('a'), row('b')] } });

		for (const li of liClasses(body)) {
			expect(li).toContain('border-b');
			expect(li).toContain('border-border');
			expect(li).toContain('bg-card');
			expect(li).toContain('lg:border-l');
		}
	});

	it('keeps the responsive two-column layout', () => {
		const { body } = render(ResultsList, {
			props: { freeRows: [row('a', false), row('b', false)] }
		});
		const ul = firstUlClass(body);

		expect(ul).toContain('grid-cols-1');
		expect(ul).toContain('lg:grid-cols-2');
	});
});
