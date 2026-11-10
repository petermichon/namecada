import { render } from 'svelte/server';
import { describe, expect, it } from 'vitest';
import ResultsList from './ResultsList.svelte';
import type { Row } from '$lib/useLookup.svelte';

function row(id: string, paid = true, status: Row['status'] = 'available'): Row {
	return {
		id,
		platform: 'github',
		platformName: 'GitHub',
		name: id,
		url: `https://example.com/${id}`,
		status,
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
	it('renders flat section headings with the group count', () => {
		const { body } = render(ResultsList, { props: { paidRows: [row('acema')] } });

		expect(body).toContain('Paid platforms');
		expect(body).toContain('1 of 1 available');
		expect(body).not.toContain('Free platforms');
	});

	it('draws separators as borders on the rows, not a card', () => {
		const { body } = render(ResultsList, { props: { paidRows: [row('a'), row('b')] } });

		for (const li of liClasses(body)) {
			expect(li).toContain('border-b');
			expect(li).toContain('border-border');
			expect(li).not.toContain('bg-card');
			expect(li).not.toContain('lg:border-l');
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

	it('excludes not-checked rows from the section count', () => {
		const { body } = render(ResultsList, {
			props: { paidRows: [row('a'), row('b', true, 'unsupported')] }
		});

		expect(body).toContain('1 of 1 available');
		expect(liClasses(body)).toHaveLength(2);
	});
});
