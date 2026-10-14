import { onMount } from 'svelte';
import { isAbort, streamFromApi } from './api';
import { toLabel, type Status } from './platforms';

const QUERY_KEY = 'domain-browser:query';

export interface Row {
	id: string;
	platform: string;
	platformName: string;
	name: string;
	url: string;
	status: Status | 'pending';
	detail?: string;
	paid: boolean;
}

export type Filter = 'all' | Status;

// Order rows alphabetically by their display name, independent of status, so
// the list stays stable as results stream in. Ties break on the provider URL so
// the order is deterministic regardless of arrival order — domains sort by TLD
// (.ai .app .com .dev .io .org) and Hex sorts packages/ before users/.
export function byPlatformName(a: Row, b: Row): number {
	if (a.platformName < b.platformName) return -1;
	if (a.platformName > b.platformName) return 1;
	if (a.url < b.url) return -1;
	if (a.url > b.url) return 1;
	return 0;
}

// Order rows paid-first, then alphabetically by display name within each group,
// so the report can split into a Paid section above a Free one.
export function byPaidThenPlatformName(a: Row, b: Row): number {
	if (a.paid !== b.paid) return a.paid ? -1 : 1;
	return byPlatformName(a, b);
}

function queryParam(): string {
	if (typeof location === 'undefined') return '';
	return new URLSearchParams(location.search).get('q') ?? '';
}

function loadQuery(): string {
	const fromUrl = queryParam();
	if (fromUrl) return fromUrl;
	if (typeof localStorage === 'undefined') return '';
	try {
		return localStorage.getItem(QUERY_KEY) ?? '';
	} catch {
		return '';
	}
}

function syncUrl(query: string) {
	const url = new URL(location.href);
	if (url.searchParams.get('q') === query) return;
	url.searchParams.set('q', query);
	history.pushState(null, '', url);
}

// Lookup owns the form state, the abort/generation guards that keep a stale
// stream from winning, and the query's URL plus localStorage persistence.
export class Lookup {
	input = $state(loadQuery());
	rows = $state<Row[]>([]);
	error = $state('');
	loading = $state(false);
	lastLabel = $state<string | null>(null);
	filter = $state<Filter>('all');
	copied = $state(false);
	// busy marks a re-scan running over rows from the previous label. The stale
	// rows stay mounted (so the document keeps its height and scroll cannot
	// clamp, P49) and are dimmed + aria-busy until the first fresh result lands.
	busy = $state(false);
	// received counts results resolved for the current generation, so the
	// summary reports progress against the in-flight scan rather than the
	// stale rows still on screen.
	received = $state(0);

	#controller: AbortController | null = null;
	#generation = 0;

	check = ({ sync = true }: { sync?: boolean } = {}): void => {
		const label = toLabel(this.input);
		const gen = ++this.#generation;

		this.#controller?.abort();
		this.#controller = null;
		this.error = '';
		this.filter = 'all';

		if (!label) {
			this.rows = [];
			this.lastLabel = null;
			this.loading = false;
			this.busy = false;
			this.received = 0;
			const raw = this.input.trim();
			this.error = !raw
				? 'Enter a name to check.'
				: raw.includes('.')
					? 'Enter the name without the dot — e.g. “example”, not “example.com”.'
					: 'Names can use letters, numbers and hyphens, e.g. acema.';
			return;
		}

		try {
			localStorage.setItem(QUERY_KEY, this.input.trim());
		} catch {
			// storage unavailable: persistence is best-effort
		}
		if (sync) syncUrl(this.input.trim());

		const ac = new AbortController();
		this.#controller = ac;
		this.lastLabel = label;
		this.loading = true;
		// Stale-while-revalidate: keep the previous label's rows mounted so the
		// document never collapses below the viewport (a shrink would clamp
		// scroll to 0, P49). They overwrite by id — one result per provider, so
		// the set is constant — and are flagged busy until the first fresh
		// result arrives.
		this.busy = this.rows.length > 0;
		this.received = 0;
		const fresh: string[] = [];

		streamFromApi(
			label,
			(result) => {
				if (gen !== this.#generation || ac.signal.aborted) return;
				fresh.push(result.id);
				this.received += 1;
				this.busy = false;
				const next: Row = {
					id: result.id,
					platform: result.platform,
					platformName: result.platformName,
					name: result.name,
					url: result.url,
					status: result.status,
					detail: result.detail,
					paid: result.paid
				};
				this.rows = this.rows.some((row) => row.id === result.id)
					? this.rows.map((row) => (row.id === result.id ? next : row))
					: [...this.rows, next];
			},
			ac.signal
		)
			.catch((e) => {
				if (isAbort(e)) return;
				if (ac.signal.aborted || gen !== this.#generation) return;
				// Drop rows the failed scan never refreshed; a re-scan that
				// cannot reach the backend must not keep showing the previous
				// label's results.
				this.rows = this.rows.filter((row) => fresh.includes(row.id));
				this.rows = this.rows.map((row) =>
					row.status === 'pending' ? { ...row, status: 'unknown', detail: 'unreachable' } : row
				);
				if (!this.rows.length) this.error = 'Could not reach the backend. Is `serve` running?';
			})
			.finally(() => {
				if (gen === this.#generation) {
					this.loading = false;
					this.busy = false;
				}
			});
	};

	copyResults = async (): Promise<void> => {
		const lines = this.rows
			.toSorted(byPlatformName)
			.map((row) => `${row.platformName}\t${row.name}\t${row.status}`);
		const text = `${this.lastLabel ?? ''}\n${lines.join('\n')}`;
		try {
			await navigator.clipboard.writeText(text);
			this.copied = true;
			setTimeout(() => (this.copied = false), 1500);
		} catch {
			// clipboard unavailable
		}
	};

	searching = $derived(this.loading || this.rows.some((row) => row.status === 'pending'));
	checked = $derived(this.received);

	stale = $derived(
		!this.searching &&
			this.rows.length > 0 &&
			this.lastLabel !== null &&
			toLabel(this.input) !== this.lastLabel
	);

	visibleRows = $derived(
		(this.filter === 'all'
			? this.rows
			: this.rows.filter((row) => row.status === this.filter)
		).toSorted(byPaidThenPlatformName)
	);

	paidRows = $derived(this.visibleRows.filter((row) => row.paid));
	freeRows = $derived(this.visibleRows.filter((row) => !row.paid));

	counts = $derived.by(() => {
		const totals = { available: 0, taken: 0, unknown: 0 };
		for (const row of this.rows) {
			if (row.status !== 'pending') totals[row.status]++;
		}
		return totals;
	});

	announcement = $derived(
		this.searching
			? ''
			: `${this.counts.available} available, ${this.counts.taken} unavailable, ${this.counts.unknown} unknown`
	);
}

export function useLookup(): Lookup {
	const lookup = new Lookup();

	onMount(() => {
		if (toLabel(lookup.input)) lookup.check({ sync: false });

		const onPop = () => {
			lookup.input = queryParam();
			lookup.check({ sync: false });
		};
		addEventListener('popstate', onPop);
		return () => removeEventListener('popstate', onPop);
	});

	return lookup;
}
