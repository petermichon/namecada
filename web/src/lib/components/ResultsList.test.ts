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

function liCount(html: string): number {
	return [...html.matchAll(/<li[\s>]/g)].length;
}

describe('ResultsList', () => {
	it('renders flat section headings with the group count', () => {
		const { body } = render(ResultsList, { props: { paidRows: [row('acema')] } });

		expect(body).toContain('Paid platforms');
		expect(body).toContain('1 of 1 available');
		expect(body).not.toContain('Free platforms');
	});

	it('styles rows like the status filters instead of a divided table', () => {
		const { body } = render(ResultsList, { props: { paidRows: [row('a'), row('b')] } });
		const ul = firstUlClass(body);

		expect(ul).toContain('gap-y-0.5');
		expect(body).toContain('rounded-md');
		expect(body).not.toContain('border-b');
		expect(liCount(body)).toBe(2);
	});

	it('keeps the responsive two-column layout', () => {
		const { body } = render(ResultsList, {
			props: { freeRows: [row('a', false), row('b', false)] }
		});
		const ul = firstUlClass(body);

		expect(ul).toContain('grid-cols-1');
		expect(ul).toContain('lg:grid-cols-2');
	});

	it('excludes not-checked rows from the section count and labels the denominator', () => {
		const { body } = render(ResultsList, {
			props: { paidRows: [row('a'), row('b', true, 'unsupported')] }
		});

		expect(body).toContain('1 of 1 checked available');
		expect(liCount(body)).toBe(2);
	});
});
